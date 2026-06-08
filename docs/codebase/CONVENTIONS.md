# Coding Conventions

## 1) Naming Rules

| Item | Rule | Example | Evidence |
|------|------|---------|----------|
| Files | Lower-case Go filenames; tests use `_test.go` | `config.go`, `config_test.go` | `internal/config/` |
| Functions/methods | Go exported identifiers use PascalCase; unexported helpers use lower camel case | `DefaultManager`, `validatePort` | `internal/config/config.go` |
| Types/interfaces | Exported types use PascalCase | `Manager`, `Worker`, `Route` | `internal/config/config.go`, `internal/api/contract.go` |
| Constants/env vars | Exported constants use PascalCase; no env vars found | `ProjectName`, `OwnerManager` | `internal/harness/harness.go`, `internal/api/contract.go` |

## 2) Formatting and Linting

- Formatter: `gofmt` via `go fmt ./...`, documented in `README.md` and `AGENTS.md`.
- Linter: no linter config detected.
- Most relevant enforced rules: Go compiler and gofmt conventions only, based on files present.
- Run commands: `go fmt ./...`, `go test ./...`.

## 3) Import and Module Conventions

- Import grouping/order: standard-library imports are grouped before project imports in entrypoints.
- Alias vs relative import policy: source uses module-qualified project imports; no relative imports are present.
- Public exports/barrel policy: no barrel pattern exists in this Go codebase.

## 4) Error and Logging Conventions

- Error strategy by layer: config validation returns `error`; CLI entrypoints print configuration errors to `stderr` and exit with status 1.
- Logging style and required context fields: no logging package or structured logging implementation exists yet.
- Sensitive-data redaction rules: `AGENTS.md` forbids committing secrets, tokens, real proxy credentials, and private OpenVPN config; no code-level redaction helper exists.

## 5) Testing Conventions

- Test file naming/location rule: tests are colocated with packages as `*_test.go`.
- Mocking strategy norm: no mocks exist yet; current tests use direct function calls and constants.
- Coverage expectation: `[TODO]` no coverage tool or threshold is configured.

## 6) Evidence

- `README.md`
- `AGENTS.md`
- `cmd/manager/main.go`
- `cmd/worker/main.go`
- `internal/api/contract.go`
- `internal/config/config.go`
- `internal/harness/harness.go`
- `internal/config/config_test.go`
