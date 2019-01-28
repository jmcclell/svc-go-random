package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRandomHandlerWithDefaults(t *testing.T) {
	req, err := http.NewRequest("GET", "/random", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler := http.HandlerFunc(randomNumberHandler)

	rand.Seed(1) // Gives series: 81, 87, 47
	vals := [3]int{81, 87, 47}
	for _, val := range vals {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		expected := fmt.Sprintf(`{"values":[%d]}`, val)
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	}
}

func TestRandomHandlerWithNumGreaterThanOne(t *testing.T) {
	req, err := http.NewRequest("GET", "/random?num=3", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler := http.HandlerFunc(randomNumberHandler)

	rand.Seed(1) // Gives series: 81, 87, 47
	vals := [3]int{81, 87, 47}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := fmt.Sprintf(`{"values":[%d,%d,%d]}`, vals[0], vals[1], vals[2])
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestRandomHandlerWithMinGreaterThanMax(t *testing.T) {
	req, err := http.NewRequest("GET", "/random?min=100&max=50", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler := http.HandlerFunc(randomNumberHandler)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}
