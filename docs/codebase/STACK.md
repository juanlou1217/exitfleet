# Technology Stack

## 1) Runtime Summary

| Area | Value | Evidence |
|------|-------|----------|
| Primary language | Go | `go.mod`, `cmd/manager/main.go`, `cmd/worker/main.go` |
| Runtime + version | Go 1.22 | `go.mod` |
| Package manager | Go modules | `go.mod` |
| Module/build system | Go module `github.com/juanlou1217/exitfleet` | `go.mod` |

## 2) Production Frameworks and Dependencies

| Dependency | Version | Role in system | Evidence |
|------------|---------|----------------|----------|
| Go standard library | Go 1.22 runtime | CLI entrypoints, config validation, address formatting, tests | `go.mod`, `internal/config/config.go` |

No third-party production dependency is declared in `go.mod`.

## 3) Development Toolchain

| Tool | Purpose | Evidence |
|------|---------|----------|
| `go test` | Run Go tests | `README.md`, `AGENTS.md` |
| `go fmt` | Format Go source | `README.md`, `AGENTS.md` |
| `rg` | Static checks referenced in test records | `docs/tests/project-initialization.md` |

No root linter config, CI config, Makefile, Dockerfile, or container orchestration file is present in the scan output.

## 4) Key Commands

```bash
go test ./...
go fmt ./...
go run ./cmd/manager
go run ./cmd/worker
```

## 5) Environment and Config

- Config sources: `internal/config/config.go`.
- Required env vars: none found in source; no `.env.example` or `.env.template` was detected.
- Deployment/runtime constraints: README describes future Release binaries and Docker/Compose usage, but also states formal Release, Dockerfile, and Compose files do not exist yet.

## 6) Evidence

- `go.mod`
- `README.md`
- `AGENTS.md`
- `internal/config/config.go`
- `docs/codebase/.codebase-scan.txt`
