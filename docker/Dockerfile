FROM golang:1.11.4-alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates

RUN adduser -D -g '' random

COPY . /tmp/random/
WORKDIR /tmp/random/

RUN go mod vendor

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build \
    -mod=vendor -o /tmp/random/random

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /tmp/random/random /go/bin/random

USER random

ENV PORT 8000
EXPOSE 8000

ENTRYPOINT ["/go/bin/random"]
