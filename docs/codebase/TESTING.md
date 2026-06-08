# Testing Patterns

## 1) Test Stack and Commands

- Primary test framework: Go `testing` package.
- Assertion/mocking tools: direct `testing.T` assertions; no third-party mocking tool.
- Commands:

```bash
go test ./...
go fmt ./...
```

No integration, E2E, or coverage command is configured in the repository.

## 2) Test Layout

- Test file placement pattern: tests are colocated with source packages.
- Naming convention: `*_test.go`, for example `contract_test.go`, `config_test.go`, and `harness_test.go`.
- Setup files and where they run: none found.

## 3) Test Scope Matrix

| Scope | Covered? | Typical target | Notes |
|-------|----------|----------------|-------|
| Unit | Yes | API route contract, config defaults/validation, harness constants | `internal/api`, `internal/config`, `internal/harness` |
| Integration | No | API server, Docker, OpenVPN, SQLite, proxy behavior | Not implemented yet |
| E2E | No | Manager/worker runtime flow | Not implemented yet |

## 4) Mocking and Isolation Strategy

- Main mocking approach: none required for current tests.
- Isolation guarantees: tests call pure functions or inspect constants; no external state is touched.
- Common failure mode in tests: config default drift, route contract drift, or harness rule drift.

## 5) Coverage and Quality Signals

- Coverage tool + threshold: `[TODO]` no threshold configured.
- Current reported coverage: `[TODO]` not measured in this pass.
- Known gaps/flaky areas: no tests cover HTTP handlers, process management, Docker lifecycle, SQLite, OpenVPN, proxying, SSE, or UI because those modules do not exist yet.

## 6) Evidence

- `internal/api/contract_test.go`
- `internal/config/config_test.go`
- `internal/harness/harness_test.go`
- `README.md`
- `AGENTS.md`
- `docs/tests/project-initialization.md`
- `docs/codebase/.codebase-scan.txt`
