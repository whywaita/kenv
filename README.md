# keex - Kubernetes Environment Extractor

[![Test](https://github.com/whywaita/keex/actions/workflows/test.yml/badge.svg)](https://github.com/whywaita/keex/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/whywaita/keex/branch/main/graph/badge.svg)](https://codecov.io/gh/whywaita/keex)
[![Go Report Card](https://goreportcard.com/badge/github.com/whywaita/keex)](https://goreportcard.com/report/github.com/whywaita/keex)

keex is a CLI tool that extracts environment variables from Kubernetes manifests and formats them for use with `docker run` or shell commands.

## Motivation

When developing or debugging containerized applications, you often need to run the same container locally with the exact same environment variables as in your Kubernetes cluster. Manually copying environment variables from Kubernetes manifests is tedious and error-prone, especially when dealing with:

- Multiple environment variables across different manifests
- Values from ConfigMaps and Secrets
- Complex deployments with many containers

keex solves this by automatically extracting all environment variables from your Kubernetes resources and formatting them for immediate use.

## Quick Examples

### Running a container locally with the same environment as in Kubernetes

```bash
# Extract env vars from a deployment and run locally
$ keex extract -f examples/deployment.yaml --mode docker
-e APP_ENV="production" -e LOG_LEVEL="info" -e DB_HOST="db.example.com" -e DB_PORT="5432" -e DB_USER="<db-secret:username>" -e DB_PASS="<db-secret:password>" -e API_KEY="<api-secret:key>" -e CONFIG_PATH="<app-config:config-path>"

# Use it with docker run
$ docker run $(keex extract -f examples/deployment.yaml --mode docker) myapp:latest

# Or from a live cluster (with actual secret values resolved)
$ kubectl get deployment myapp -o yaml | keex extract -f - --mode docker | xargs docker run myapp:latest
```

### Debugging with local tools using production environment

```bash
# Export Kubernetes env vars to your shell
$ keex extract -f examples/deployment.yaml --mode env
APP_ENV="production" LOG_LEVEL="info" DB_HOST="db.example.com" DB_PORT="5432" DB_USER="<db-secret:username>" DB_PASS="<db-secret:password>" API_KEY="<api-secret:key>" CONFIG_PATH="<app-config:config-path>"

# Evaluate in your shell
$ eval $(keex extract -f examples/deployment.yaml --mode env)

# Now run your app locally with production config
$ go run main.go
# or
$ python app.py
```

### Comparing environments between different deployments

```bash
# Extract and save environment from staging
$ keex extract -f staging-deployment.yaml > staging.env

# Extract and save environment from production  
$ keex extract -f prod-deployment.yaml > prod.env

# Compare the differences
$ diff staging.env prod.env
3c3
< DB_HOST="staging-db.example.com"
---
> DB_HOST="prod-db.example.com"
```

## Features

- Extract environment variables from Deployment, StatefulSet, DaemonSet, Job, CronJob, and Pod resources
- Automatically resolve Secret and ConfigMap references when kubeconfig is available
- Support for multiple output formats (docker, env)
- Specify target container in multi-container pods
- Optional redaction of sensitive values
- Read manifests from file or stdin

## Installation

### Standalone CLI

```bash
go install github.com/whywaita/keex/cmd/keex@latest
```

Or build from source:

```bash
git clone https://github.com/whywaita/keex.git
cd keex
go build -o keex cmd/keex/*.go
```

### kubectl Plugin

keex can also be used as a kubectl plugin, allowing you to extract environment variables directly from live Kubernetes resources:

```bash
# Install the kubectl plugin
go install github.com/whywaita/keex/cmd/kubectl-eex@latest

# Or build from source
git clone https://github.com/whywaita/keex.git
cd keex
make install-plugin
```

Once installed, you can use it with kubectl:

```bash
# Extract env vars from a live deployment
kubectl eex deployment/myapp

# Extract from a specific container
kubectl eex deployment/myapp -c app

# Different output formats
kubectl eex deployment/myapp --format docker
kubectl eex deployment/myapp --format shell --export
kubectl eex deployment/myapp --format dotenv > .env
kubectl eex deployment/myapp --format compose

# Extract from other resource types
kubectl eex statefulset/database
kubectl eex pod/mypod-xyz123
kubectl eex job/migrate-db
kubectl eex cronjob/backup
```

## Usage

### Basic Usage

```bash
# Extract environment variables from a deployment
keex extract -f deployment.yaml

# Extract from a live cluster
kubectl get deployment myapp -o yaml | keex extract -f -
```

### Output Modes

**Docker mode** - Format for `docker run`:
```bash
keex extract -f deployment.yaml --mode docker
# Output: -e DB_HOST=db.example.com -e DB_USER=admin -e DB_PASS=secret ...
```

**Environment mode** - Format for shell export:
```bash
keex extract -f deployment.yaml --mode env
# Output: DB_HOST="db.example.com" DB_USER="admin" DB_PASS="secret" ...
```

### Advanced Examples

**Working with multi-container pods:**
```bash
# Target a specific container by name
keex extract -f pod.yaml --container app
keex extract -f pod.yaml --container sidecar
```

**Security and sensitive data:**
```bash
# Redact secret values in output (useful for sharing configs)
keex extract -f deployment.yaml --redact
# Output: DB_HOST="db.example.com" DB_PASS="***REDACTED***" ...

# Resolve actual values from ConfigMaps and Secrets
keex extract -f deployment.yaml --context production --namespace backend
```

**Integration with other tools:**
```bash
# Create an env file for docker-compose
keex extract -f deployment.yaml --mode env > .env

# Pass environment to local development server
keex extract -f deployment.yaml --mode env | grep -E "^API_|^DB_" > local.env
source local.env && npm run dev

# Quick environment debugging
keex extract -f deployment.yaml | grep DATABASE
```

## Command Line Options

```
Usage:
  keex extract [flags]

Flags:
  -f, --file string        Manifest file path ("-" for stdin)
      --mode string        Output mode: docker|env (default "env")
      --container string   Target container name
      --context string     kubeconfig context (default: current)
      --namespace string   Kubernetes namespace (default: manifest/ns)
      --redact             Mask secret values in output
  -h, --help               Show help
```

## Requirements

- Go 1.24.4 or higher
- Access to Kubernetes cluster (optional - only needed for secret/configmap resolution)
- Valid kubeconfig file (optional - secrets/configmaps will show placeholders without it)

## Development

Run tests:
```bash
go test ./...
```

Build:
```bash
go build -o keex cmd/keex/*.go
```

## License

MIT
