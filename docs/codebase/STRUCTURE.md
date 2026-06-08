# Codebase Structure

## 1) Top-Level Map

| Path | Purpose | Evidence |
|------|---------|----------|
| `cmd/` | Executable entrypoints for manager and worker | `cmd/manager/main.go`, `cmd/worker/main.go`, `AGENTS.md` |
| `internal/api/` | Shared REST and worker health route contract | `internal/api/contract.go`, `docs/api.md` |
| `internal/config/` | Manager and worker config models, defaults, and validation | `internal/config/config.go` |
| `internal/harness/` | Project identity and architecture guardrail constants | `internal/harness/harness.go` |
| `docs/` | Architecture, API, design, test, decision, template, and codebase documentation | `AGENTS.md`, `docs/architecture.md` |
| `README.md` | Public project overview and current-stage notes | `README.md` |
| `AGENTS.md` | AI modification rules, project navigation, and hard constraints | `AGENTS.md` |
| `go.mod` | Go module declaration | `go.mod` |

## 2) Entry Points

- Main runtime entry: `cmd/manager/main.go`.
- Secondary entry point: `cmd/worker/main.go`.
- How entry is selected: Go package command paths are run directly with `go run ./cmd/manager` or `go run ./cmd/worker`, as documented in `README.md` and `AGENTS.md`.

## 3) Module Boundaries

| Boundary | What belongs here | What must not be here |
|----------|-------------------|------------------------|
| `cmd/` | Startup, config loading, dependency assembly | Manager/Worker business logic, Docker scheduling, proxy traffic |
| `internal/api/` | Route contract types and route lists | Concrete HTTP framework implementation unless the package is intentionally expanded |
| `internal/config/` | Explicit config structs, defaults, validation helpers | Runtime scheduling, process management, persistent state |
| `internal/harness/` | Durable project identity and architecture rules | Feature implementation details |
| `docs/` | Project architecture, API, designs, decisions, tests, and templates | Source code or copied upstream Python project |

## 4) Naming and Organization Rules

- File naming pattern: lower-case Go filenames with underscores for tests, for example `contract.go` and `contract_test.go`.
- Directory organization pattern: layer/module-oriented directories under `cmd/` and `internal/`.
- Import path convention: internal packages use module-qualified imports such as `github.com/juanlou1217/exitfleet/internal/config`.

## 5) Evidence

- `docs/codebase/.codebase-scan.txt`
- `AGENTS.md`
- `README.md`
- `cmd/manager/main.go`
- `cmd/worker/main.go`
- `internal/api/contract.go`
- `internal/config/config.go`
- `internal/harness/harness.go`
