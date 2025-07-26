FROM golang:1.24 AS builder

WORKDIR /app

# Copy go module files first for caching
COPY go.mod ./
RUN go mod download

COPY main.go .
RUN go build -o k8s-jwks-proxy

FROM busybox
COPY --from=builder /app/k8s-jwks-proxy /usr/local/bin/k8s-jwks-proxy
ENTRYPOINT ["/usr/local/bin/k8s-jwks-proxy"]
