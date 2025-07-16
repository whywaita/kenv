# Go CLI Tool: keex – One-Line Environment Variable Extractor for Kubernetes Manifests

## 1. Purpose

Build a standalone CLI (**keex**) that quickly reproduces the exact set of environment variables and credentials used by an existing Kubernetes Deployment so developers can spin up identical containers locally with `docker run`, or export them directly into their shell.

## 2. Background & Motivation

* Copying dozens of `env` / `secret` entries from a live Deployment into a local shell is slow and error-prone.
* Ad-hoc scripts have proliferated for similar tasks; a unified OSS CLI improves reliability and shareability.
* Emitting everything on a single line enables easy copy-paste or script embedding.

## 3. Scope

* **Input:** Kubernetes manifests (YAML/JSON) containing Deployment / Pod / CronJob resources, provided via file (`-f`) or STDIN.
* **Output:** A single-line string of environment variables in two modes: *docker* (`-e KEY=VALUE`) and *env* (`KEY="VALUE"`).
* **Secret Resolution:** Resolve `valueFrom.secretKeyRef` (and `configMapKeyRef`) to actual values using the current kube-context or supplied flags.
* **Target Kubernetes Version:** v1.22 or newer.

## 4. Terminology

| Term            | Description                                                                           |
| --------------- | ------------------------------------------------------------------------------------- |
| **docker mode** | Concatenates `-e KEY=VALUE` pairs suitable for direct use with `docker run`.          |
| **env mode**    | Concatenates `KEY="VALUE"` pairs (optionally prefixed with `export`) for shell usage. |
| **Manifest**    | A YAML/JSON file describing Kubernetes resources.                                     |

## 5. Functional Requirements (FR)

| ID        | Requirement                                                          | Priority |
| --------- | -------------------------------------------------------------------- | -------- |
| **FR-1**  | Accept manifest via `-f/--file`                                      | Must     |
| **FR-2**  | Accept manifest from STDIN (`-`)                                     | Should   |
| **FR-3**  | Parse `spec.template.spec.containers[*].env` and `envFrom` sections  | Must     |
| **FR-4**  | Resolve `secretKeyRef` / `configMapKeyRef` values                    | Must     |
| **FR-5**  | Flatten variables into a single line                                 | Must     |
| **FR-6**  | Switch output mode with `--mode` (`docker` / `env`)                  | Must     |
| **FR-7**  | Select target container via `--container` (default: first container) | Should   |
| **FR-8**  | Override kube-context & namespace via `--context`, `--namespace`     | Could    |
| **FR-9**  | Disable secret resolution with `--no-secret`                         | Could    |
| **FR-10** | Redact sensitive values in output with `--redact`                    | Could    |

### 5.1 docker-mode Example

```bash
kenv extract -f deploy.yaml --mode docker
# -e DB_HOST=db.example.com -e DB_USER=admin -e DB_PASS=secret ...
```

### 5.2 env-mode Example

```bash
kenv extract -f deploy.yaml --mode env
# DB_HOST="db.example.com" DB_USER="admin" DB_PASS="secret" ...
```

## 6. Non-Functional Requirements (NFR)

| ID        | Requirement | Metric                                                     |
| --------- | ----------- | ---------------------------------------------------------- |
| **NFR-1** | Performance | 1 000-line manifest + 50 secrets → ≤ 2 s                   |
| **NFR-2** | Security    | Secret values remain only in memory; never written to disk |
| **NFR-3** | Runtime     | Go 1.24.4; supports Linux / macOS / Windows                |
| **NFR-4** | Portability | Single static binary (CGO disabled)                        |
| **NFR-5** | Logging     | `--verbose` enables debug; default = warn+                 |

## 7. CLI Interface

```text
Usage:
  kenv extract [flags]

Flags:
  -f, --file string        Manifest file path ("-" for stdin)
      --mode string        Output mode: docker|env (default "env")
      --container string   Target container name
      --context string     kubeconfig context (default: current)
      --namespace string   Kubernetes namespace (default: manifest/ns)
      --no-secret          Skip resolving Secret values
      --redact             Mask secret values in output
  -h, --help               Show help
```

## 8. Error Handling

| Scenario                  | Behaviour                           |
| ------------------------- | ----------------------------------- |
| Manifest parse failure    | exit 1, message: `invalid manifest` |
| Secret resolution failure | exit 2, warn log & skip key         |
| Invalid mode              | exit 3, show help                   |

## 9. Assumptions & Constraints

* Secret resolution requires cluster access via kube-config.
* `binaryData` in Secret is ignored.
* Template values like `$(NAME)` remain unexpanded.

## 10. Acceptance Criteria

1. `kenv extract -f sample.yaml --mode docker` produces the expected single line.
2. Secrets are resolved and included when cluster access is available.
3. `--no-secret` outputs placeholders (e.g. `DB_PASS=<secret>`).
4. Unsupported resource types produce clear errors.

## 12. References

* Kubernetes API v1.22: [https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/)
* client-go: [https://github.com/kubernetes/client-go](https://github.com/kubernetes/client-go)
* Docker CLI: [https://docs.docker.com/engine/reference/commandline/run/](https://docs.docker.com/engine/reference/commandline/run/)

