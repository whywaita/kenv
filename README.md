# kenv - Kubernetes Environment Variable Extractor

kenv is a CLI tool that extracts environment variables from Kubernetes manifests and formats them for use with `docker run` or shell commands.

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

### Basic usage

Extract environment variables from a deployment:

```bash
kenv extract -f deployment.yaml
```

### Output modes

Docker mode (for use with `docker run`):
```bash
kenv extract -f deployment.yaml --mode docker
# Output: -e DB_HOST=db.example.com -e DB_USER=admin -e DB_PASS=secret ...
```

Environment mode (for shell):
```bash
kenv extract -f deployment.yaml --mode env
# Output: DB_HOST="db.example.com" DB_USER="admin" DB_PASS="secret" ...
```

### Examples

Use with docker run:
```bash
docker run $(kenv extract -f deployment.yaml --mode docker) myimage:latest
```

Export to shell:
```bash
eval $(kenv extract -f deployment.yaml --mode env)
```

Read from stdin:
```bash
kubectl get deployment myapp -o yaml | kenv extract -f - --mode docker
```

Target specific container:
```bash
kenv extract -f multi-container-pod.yaml --container app
```

Redact sensitive values:
```bash
kenv extract -f deployment.yaml --redact
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
