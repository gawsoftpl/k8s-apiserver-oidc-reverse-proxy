package main

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"net/http"
	"os"
)

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func main() {
	// Get file paths from env or fallback to default in-cluster paths
	tokenPath := getEnv("TOKEN_PATH", "/var/run/secrets/kubernetes.io/serviceaccount/token")
	caPath := getEnv("CA_CERT_PATH", "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")

	// Read ServiceAccount token
	token, err := os.ReadFile(tokenPath)
	if err != nil {
		log.Fatalf("Failed to read token from %s: %v", tokenPath, err)
	}

	// Read CA cert
	caCert, err := os.ReadFile(caPath)
	if err != nil {
		log.Fatalf("Failed to read CA certificate from %s: %v", caPath, err)
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatalf("Failed to parse CA certificate from %s", caPath)
	}

	// HTTP client with custom CA
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	// Proxy endpoints
	http.HandleFunc("/openid/v1/jwks", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, client, string(token), "https://kubernetes.default.svc/openid/v1/jwks")
	})

	http.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, client, string(token), "https://kubernetes.default.svc/.well-known/openid-configuration")
	})

	// Health check endpoint
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func proxyRequest(w http.ResponseWriter, client *http.Client, token string, url string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to contact API server: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Forward headers & body
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
