package main

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type cacheEntry struct {
	data      []byte
	headers   http.Header
	expiresAt time.Time
}

var (
	cache      = make(map[string]*cacheEntry)
	cacheMutex = sync.RWMutex{}
	cacheTTL   = getCacheTTL() // Set by env or fallback to 5 minutes
)

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getCacheTTL() time.Duration {
	defaultTTL := 2
	ttlStr := os.Getenv("CACHE_TTL_MINUTES")
	if ttlStr == "" {
		return time.Duration(defaultTTL) * time.Minute
	}
	if ttlMinutes, err := strconv.Atoi(ttlStr); err == nil && ttlMinutes > 0 {
		return time.Duration(ttlMinutes) * time.Minute
	}
	log.Printf("Invalid CACHE_TTL_MINUTES='%s', using default %d minutes", ttlStr, defaultTTL)
	return time.Duration(defaultTTL) * time.Minute
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

	// Proxy endpoints with caching
	http.HandleFunc("/openid/v1/jwks", func(w http.ResponseWriter, r *http.Request) {
		handleWithCache(w, client, string(token), "https://kubernetes.default.svc/openid/v1/jwks")
	})

	http.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		handleWithCache(w, client, string(token), "https://kubernetes.default.svc/.well-known/openid-configuration")
	})

	// Health check endpoint
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("Listening on :8080 with cache TTL: %s", cacheTTL)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWithCache(w http.ResponseWriter, client *http.Client, token, url string) {
	cacheMutex.RLock()
	entry, found := cache[url]
	cacheMutex.RUnlock()

	if found && time.Now().Before(entry.expiresAt) {
		// Serve from cache
		for k, vv := range entry.headers {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write(entry.data)
		return
	}

	// Fetch from API
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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read API response", http.StatusInternalServerError)
		return
	}

	newEntry := &cacheEntry{
		data:      data,
		headers:   cloneHeader(resp.Header),
		expiresAt: time.Now().Add(cacheTTL),
	}

	cacheMutex.Lock()
	cache[url] = newEntry
	cacheMutex.Unlock()

	// Return to client
	for k, vv := range newEntry.headers {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(data)
}

func cloneHeader(h http.Header) http.Header {
	copy := make(http.Header, len(h))
	for k, v := range h {
		copy[k] = append([]string(nil), v...)
	}
	return copy
}
