# ArgoCD Auto-Tuner based on CPU Usage

This project implements a full GitOps flow where CPU usage metrics from Prometheus trigger a dynamic change to an NGINX Helm release. When CPU exceeds a threshold, a Go-based tuner edits `values.yaml`, commits the change, and Argo CD syncs it automatically.

---

## Architecture

- **Go Tuner**  
  Queries Prometheus regularly. If CPU usage > threshold:
  - Updates `replicaCount` in `values.yaml`
  - Commits and pushes to Git

- **Argo CD**  
  Monitors the repo and syncs the NGINX Helm chart from `charts/nginx`

- **Prometheus**  
  Scrapes container CPU metrics

- **Helm**  
  Manages the NGINX deployment with configurable values

---

## Prerequisites

- Kubernetes cluster
- Argo CD installed
- Prometheus scraping pod CPU metrics
- Go 1.20+

---

## Usage

### 1. Deploy NGINX (or any pod) via Argo CD

Use the existing `Application` manifest in the repo to deploy the chart from `charts/nginx`.

```bash
kubectl apply -f argocd-nginx-app.yaml
```

### 2. Run the Go Tuner

Edit `tuner/main.go` if needed:

```go
const (
    prometheusURL = "http://localhost:9090"
    threshold     = 0.75
    repoPath      = "charts/nginx"
    valuesFile    = "values.yaml"
    targetPod     = "nginx-auto-tuned"
)
```

Then:

```bash
go run tuner/main.go
```

---

## Project Structure

```
.
├── charts
│   └── nginx
│       ├── templates/
│       ├── Chart.yaml
│       └── values.yaml
├── tuner
│   └── main.go
├── argocd-nginx-app.yaml
└── README.md
```

---

## License

MIT
