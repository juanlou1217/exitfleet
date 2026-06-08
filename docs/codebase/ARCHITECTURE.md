# Architecture

## 1) Architectural Style

- Primary style: small Go skeleton with explicit manager/worker boundaries and early contract modules.
- Why this classification: `cmd/manager` and `cmd/worker` are separate executable entrypoints; shared contracts live in `internal/api`, configuration in `internal/config`, and architecture guardrails in `internal/harness`.
- Primary constraints:
  - One active exit node maps to exactly one worker container.
  - Manager does not carry user proxy traffic.
  - Worker owns one exit and does not manage global scheduling.

## 2) System Flow

Current implemented flow:

```text
go run ./cmd/manager -> DefaultManager -> Validate -> print manager listen address
go run ./cmd/worker -> DefaultWorker -> Validate -> print proxy listen address and TUN device
```

Planned flow described by docs:

```text
manager -> node inventory/testing/scoring -> scheduler -> Docker worker lifecycle -> worker OpenVPN/TUN/proxy/health
```

The planned flow is documented in `README.md`, `AGENTS.md`, and `docs/architecture.md`; only the entrypoint/config/contract skeleton is currently implemented.

## 3) Layer/Module Responsibilities

| Layer or module | Owns | Must not own | Evidence |
|-----------------|------|--------------|----------|
| `cmd/manager` | Manager process startup and default config validation | Proxy traffic, scheduler implementation, Docker lifecycle implementation | `cmd/manager/main.go`, `AGENTS.md` |
| `cmd/worker` | Worker process startup and default config validation | Global node inventory or scheduling | `cmd/worker/main.go`, `AGENTS.md` |
| `internal/api` | Manager and worker route declarations | HTTP server runtime behavior | `internal/api/contract.go`, `docs/api.md` |
| `internal/config` | Default manager and worker config and validation | Env loading, persistence, process orchestration | `internal/config/config.go` |
| `internal/harness` | Project identity and runtime rule constants | Runtime behavior | `internal/harness/harness.go` |
| `docs` | Architecture, API, design, test, and decision records | Executable behavior | `AGENTS.md` |

## 4) Reused Patterns

| Pattern | Where found | Why it exists |
|---------|-------------|---------------|
| Default config constructor | `DefaultManager`, `DefaultWorker` in `internal/config/config.go` | Centralizes initial ports, bind hosts, and TUN defaults |
| Validation method on config struct | `Manager.Validate`, `Worker.Validate` in `internal/config/config.go` | Keeps invalid config rejection close to config data |
| Contract list with owner filter | `Routes`, `ManagerRoutes`, `WorkerRoutes` in `internal/api/contract.go` | Keeps route definitions centralized and separable by owner |
| Harness constants | `internal/harness/harness.go` | Gives tests and future work stable project constraints |

## 5) Known Architectural Risks

- Planned integrations are not implemented yet: Docker worker lifecycle, SQLite persistence, OpenVPN process control, proxy serving, SSE logs, and Web UI are documented goals but absent from source.
- The API is a route contract only; no HTTP server currently enforces behavior, auth, validation, or response schemas.
- README and architecture docs are broader than the current implementation, so future work should keep distinguishing planned behavior from implemented behavior.

## 6) Evidence

- `cmd/manager/main.go`
- `cmd/worker/main.go`
- `internal/api/contract.go`
- `internal/config/config.go`
- `internal/harness/harness.go`
- `README.md`
- `AGENTS.md`
- `docs/architecture.md`
