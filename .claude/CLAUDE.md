# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

kenv is a tool that extracts environment variables from Kubernetes manifests (Pod, Deployment, StatefulSet, etc.) and converts them to docker run or shell format. It can resolve actual values from Secrets and ConfigMaps when kubeconfig is available.

## Development Commands

### Build
```bash
# Local build
go build -o kenv cmd/kenv/*.go

# Install
go install github.com/whywaita/kenv/cmd/kenv@latest
```

### Test
```bash
# Run all tests
go test ./...

# Run with verbose output, race detection, and coverage
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run specific package tests
go test ./pkg/extractor/
go test ./pkg/formatter/
```

### Lint
```bash
# Format check
gofmt -l .

# Run golangci-lint (if installed)
golangci-lint run

# Check module dependencies
go mod tidy
```

### Security Checks
```bash
# Install and run govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## Architecture

The codebase follows a modular design with clear separation of concerns:

- **cmd/kenv/** - CLI entry point using Cobra framework
  - `main.go` - Application initialization
  - `root.go` - Root command configuration
  - `run.go` - Main command logic

- **pkg/extractor/** - Core logic for extracting environment variables from Kubernetes manifests
  - Supports Pod, Deployment, StatefulSet, DaemonSet, Job, CronJob
  - Handles valueFrom references (ConfigMap, Secret)
  - Multi-container support with container selection

- **pkg/formatter/** - Output formatting logic
  - `docker` format: Generates `-e KEY=VALUE` for docker run
  - `env` format: Generates `export KEY=VALUE` for shell scripts
  - Value masking for sensitive data

- **pkg/resolver/** - Resolves actual values from Kubernetes cluster
  - Uses kubeconfig to connect to cluster
  - Fetches Secret and ConfigMap values
  - Falls back to placeholders when values cannot be resolved

## Key Design Patterns

1. **Interface-based abstraction**: Each package defines clear interfaces for extensibility
2. **Error handling**: Uses error codes (e.g., `ERR001`, `ERR002`) for better error tracking
3. **Dependency injection**: Components are loosely coupled through interfaces
4. **Single responsibility**: Each package has a focused purpose

## Testing Strategy

- Unit tests for each package in `*_test.go` files
- Test data includes various Kubernetes manifest scenarios
- Tests cover edge cases like missing resources, multi-container pods, and error conditions

## Release Process

Releases are automated via GitHub Actions and GoReleaser:
1. Tag push with `v*` pattern triggers release workflow
2. GoReleaser builds multi-platform binaries (Linux/Windows/Darwin, amd64/arm64)
3. Creates GitHub release with checksums and archives

## Important Notes

- The tool is designed to work both with and without cluster access
- When kubeconfig is unavailable, it uses placeholder values
- Supports reading from stdin or file path
- Security feature: can mask sensitive values in output
- Multi-container pods require specifying container name with `-c` flag