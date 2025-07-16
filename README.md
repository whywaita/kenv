# kenv - Kubernetes Environment Variable Extractor

kenv is a CLI tool that extracts environment variables from Kubernetes manifests and formats them for use with `docker run` or shell commands.

## Motivation

When developing or debugging containerized applications, you often need to run the same container locally with the exact same environment variables as in your Kubernetes cluster. Manually copying environment variables from Kubernetes manifests is tedious and error-prone, especially when dealing with:

- Multiple environment variables across different manifests
- Values from ConfigMaps and Secrets
- Complex deployments with many containers

kenv solves this by automatically extracting all environment variables from your Kubernetes resources and formatting them for immediate use.

## Quick Examples

### Running a container locally with the same environment as in Kubernetes

```bash
# Extract env vars from a deployment and run locally
docker run $(kenv extract -f deployment.yaml --mode docker) myapp:latest

# Or from a live cluster
kubectl get deployment myapp -o yaml | kenv extract -f - --mode docker | xargs docker run myapp:latest
```

### Debugging with local tools using production environment

```bash
# Export Kubernetes env vars to your shell
eval $(kenv extract -f deployment.yaml --mode env)

# Now run your app locally with production config
go run main.go
# or
python app.py
```

### Comparing environments between different deployments

```bash
# Extract and save environment from staging
kenv extract -f staging-deployment.yaml > staging.env

# Extract and save environment from production  
kenv extract -f prod-deployment.yaml > prod.env

# Compare the differences
diff staging.env prod.env
```

## Features

- Extract environment variables from Deployment, StatefulSet, DaemonSet, Job, CronJob, and Pod resources
- Automatically resolve Secret and ConfigMap references when kubeconfig is available
- Support for multiple output formats (docker, env)
- Specify target container in multi-container pods
- Optional redaction of sensitive values
- Read manifests from file or stdin

## Installation

```bash
go install github.com/whywaita/kenv/cmd/kenv@latest
```

Or build from source:

```bash
git clone https://github.com/whywaita/kenv.git
cd kenv
go build -o kenv cmd/kenv/*.go
```

## Usage

### Basic Usage

```bash
# Extract environment variables from a deployment
kenv extract -f deployment.yaml

# Extract from a live cluster
kubectl get deployment myapp -o yaml | kenv extract -f -
```

### Output Modes

**Docker mode** - Format for `docker run`:
```bash
kenv extract -f deployment.yaml --mode docker
# Output: -e DB_HOST=db.example.com -e DB_USER=admin -e DB_PASS=secret ...
```

**Environment mode** - Format for shell export:
```bash
kenv extract -f deployment.yaml --mode env
# Output: DB_HOST="db.example.com" DB_USER="admin" DB_PASS="secret" ...
```

### Advanced Examples

**Working with multi-container pods:**
```bash
# Target a specific container by name
kenv extract -f pod.yaml --container app
kenv extract -f pod.yaml --container sidecar
```

**Security and sensitive data:**
```bash
# Redact secret values in output (useful for sharing configs)
kenv extract -f deployment.yaml --redact
# Output: DB_HOST="db.example.com" DB_PASS="***REDACTED***" ...

# Resolve actual values from ConfigMaps and Secrets
kenv extract -f deployment.yaml --context production --namespace backend
```

**Integration with other tools:**
```bash
# Create an env file for docker-compose
kenv extract -f deployment.yaml --mode env > .env

# Pass environment to local development server
kenv extract -f deployment.yaml --mode env | grep -E "^API_|^DB_" > local.env
source local.env && npm run dev

# Quick environment debugging
kenv extract -f deployment.yaml | grep DATABASE
```

## Command Line Options

```
Usage:
  kenv extract [flags]

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
go build -o kenv cmd/kenv/*.go
```

## License

MIT
