FROM docker.io/golang:1.27 AS builder

WORKDIR /app

# Copy go module files first for caching
COPY go.mod ./
RUN go mod download

COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o k8s-jwks-proxy-amd64 .

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=builder /app/k8s-jwks-proxy-amd64 /usr/local/bin/k8s-jwks-proxy

ENTRYPOINT ["/usr/local/bin/k8s-jwks-proxy"]
