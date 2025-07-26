Sure! Here's a clean and concise `README.md` for your Go-based JWKS reverse proxy project:

---

# 🔐 JWKS Reverse Proxy for Kubernetes API Server

This is a lightweight reverse proxy written in Go that exposes Kubernetes API Server's OIDC discovery endpoints (`/.well-known/openid-configuration` and `/openid/v1/jwks`) **securely to the public**, **without enabling anonymous access** (`--anonymous-auth=false`).

It authenticates to the Kubernetes API Server using a Service Account token and validates its TLS certificate using the in-cluster CA.

--- 
## Install

Install helm chart
```sh
helm repo add k8s-jwks-proxy https://gawsoft.pl.github.io/k8s-jwks-proxy
helm repo update
```

```sh
docker run -it --rm ghcr.io/gawsoftpl/k8s-jwks-proxy:latest
```
---

## ✨ Features

* Securely proxies OIDC endpoints
* Uses in-cluster Service Account for authentication
* TLS validation via Kubernetes CA bundle
* Lightweight, production-ready, and easy to deploy

---

## 🔧 Configuration

This proxy expects to run **inside a Kubernetes cluster** and relies on:

* The Service Account token (`/var/run/secrets/kubernetes.io/serviceaccount/token`)
* The Kubernetes CA certificate (`/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`)

No extra configuration is needed.

---

## 📦 Building

```bash
go build -o jwks-proxy main.go
```

Or use the provided Dockerfile:

```bash
docker build -t your-registry/jwks-proxy:latest .
```

---

## 🚀 Running in Kubernetes

1. **Deploy the proxy using a Service Account** with limited read access to:

   * `/openid/v1/jwks`
   * `/.well-known/openid-configuration`

2. **Example Deployment:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: jwks-proxy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: jwks-proxy
  template:
    metadata:
      labels:
        app: jwks-proxy
    spec:
      serviceAccountName: jwks-proxy
      containers:
      - name: proxy
        image: your-registry/jwks-proxy:latest
        ports:
        - containerPort: 8080
```

3. **Ingress Example:**

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: jwks-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx
  rules:
  - host: jwks.example.com
    http:
      paths:
      - path: /openid/v1/jwks
        pathType: Prefix
        backend:
          service:
            name: jwks-proxy
            port:
              number: 8080
      - path: /.well-known/openid-configuration
        pathType: Prefix
        backend:
          service:
            name: jwks-proxy
            port:
              number: 8080
```

---

## 🔐 Security

* Works with `--anonymous-auth=false`
* Service Account should have **minimal RBAC permissions**:

  ```yaml
  rules:
  - nonResourceURLs: ["/openid/v1/jwks", "/.well-known/openid-configuration"]
    verbs: ["get"]
  ```

---

## 📜 License

MIT — use freely, modify responsibly.

---

Let me know if you want me to package this into a GitHub-ready repo (`main.go`, `Dockerfile`, `manifests/`, etc.).
