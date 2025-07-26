# 🔐 JWKS Reverse Proxy for Kubernetes API Server

This is a lightweight reverse proxy written in Go that exposes Kubernetes API Server's OIDC discovery endpoints (`/.well-known/openid-configuration` and `/openid/v1/jwks`) **securely to the public**, **without enabling anonymous access** (`--anonymous-auth=false`).

It authenticates to the Kubernetes API Server using a Service Account token and validates its TLS certificate using the in-cluster CA.

--- 
## Install

Install helm chart
```sh
helm install k8s-jwks-proxy oci://ghcr.io/gawsoft/k8s-jwks-proxy
```

Run docker container
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

---

## 🚀 Running in Kubernetes

1. **Deploy the proxy using a Service Account** with limited read access to:

   * `/openid/v1/jwks`
   * `/.well-known/openid-configuration`

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
