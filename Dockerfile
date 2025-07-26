FROM golang:1.22 AS builder

WORKDIR /app
COPY main.go .
RUN go build -o k8s-jwks-proxy

FROM alpine
COPY --from=builder /app/jwks-proxy /usr/local/bin/k8s-jwks-proxy
ENTRYPOINT ["/usr/local/bin/k8s-jwks-proxy"]
