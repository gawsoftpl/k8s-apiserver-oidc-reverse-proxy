FROM docker.io/golang:1.24 AS builder

WORKDIR /app

# Copy go module files first for caching
COPY go.mod ./
RUN go mod download

COPY main.go .
RUN go build -o k8s-jwks-proxy-amd64

FROM docker.io/busybox
COPY --from=builder /app/k8s-jwks-proxy-amd64 /usr/local/bin/k8s-jwks-proxy

USER 1000

ENTRYPOINT ["/usr/local/bin/k8s-jwks-proxy"]
