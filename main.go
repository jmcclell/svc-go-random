package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/schema"
	"github.com/heptiolabs/healthcheck"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var paramDecoder = schema.NewDecoder()
var config Config
var metrics = prometheus.NewRegistry()
var health = healthcheck.NewMetricsHandler(metrics, "random")
var serverStatus = Starting
var version = "dev"

type Config struct {
	Port                    int16         `default:"80"`
	AdminPort               int16         `default:"9000"`
	GracefulShutdownTimeout time.Duration `default:"30s"`
}

func main() {
	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, os.Interrupt)

	initConfig()
	initAdminServer()

	router := http.NewServeMux()
	router.HandleFunc("/", randomNumberHandler)

	server := &http.Server{
		Handler: router,
	}

	log.Printf("Starting HTTP server on 0.0.0.0:%d", config.Port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		log.Fatal(err)
	}

	go server.Serve(listener)
	log.Println("Ready to serve requesets")
	serverStatus = Running

	<-shutdown

	serverStatus = ShuttingDown
	log.Println("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.GracefulShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}

	log.Println("Graceful shutdown complete.")
}

func initConfig() {
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err)
	}
}

func initAdminServer() {
	initHealthcheck()

	adminRouter := http.NewServeMux()
	adminRouter.Handle("/metrics", promhttp.HandlerFor(metrics, promhttp.HandlerOpts{}))
	adminRouter.HandleFunc("/live", health.LiveEndpoint)
	adminRouter.HandleFunc("/ready", health.ReadyEndpoint)
	adminRouter.HandleFunc("/about", aboutHandler)

	adminServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.AdminPort),
		Handler: adminRouter,
	}

	log.Printf("Starting admin server on 0.0.0.0:%d", config.AdminPort)
	go func() {
		err := adminServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Println(err.Error())
		}
	}()
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	response := AboutResponse{Name: "random", Version: version}
	response.render(w)
}

func initHealthcheck() {
	health.AddReadinessCheck("http", func() error {
		if serverStatus == Running {
			return nil
		} else {
			return fmt.Errorf("HTTP server is %s", serverStatus)
		}
	})
}

func randomNumberHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		ErrorResponse{Error: err.Error()}.render(w, http.StatusBadRequest)
		return
	}

	params := RandomRequest{Min: 0, Max: 100}

	err = paramDecoder.Decode(&params, r.Form)
	if err != nil {
		ErrorResponse{Error: err.Error()}.render(w, http.StatusBadRequest)
		return
	}

	if params.Min > params.Max {
		ErrorResponse{Error: "min cannot be greater than max"}.render(w, http.StatusBadRequest)
		return
	}

	val := rand.Intn(params.Max-params.Min) + params.Min

	NumberResponse{Value: int64(val)}.render(w)
}

type RandomRequest struct {
	Min int `schema:"min"`
	Max int `schema:"max"`
}

type NumberResponse struct {
	Value int64 `json:"value"`
}

func (v NumberResponse) render(w http.ResponseWriter) {
	encoded, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(encoded)
}

type AboutResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (a AboutResponse) render(w http.ResponseWriter) {
	encoded, err := json.Marshal(a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(encoded)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (e ErrorResponse) render(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "application/json")

	encoded, err := json.Marshal(e)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{ "error": "%s" }`, err.Error()), http.StatusInternalServerError)
	}

	http.Error(w, string(encoded), code)
}

type ServerStatus int

const (
	Starting ServerStatus = iota
	Running
	ShuttingDown
)

func (s ServerStatus) String() string {
	return [...]string{"starting", "running", "shutting down"}[s]
}
