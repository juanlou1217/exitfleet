# Codebase Concerns

## 1) Top Risks (Prioritized)

| Severity | Concern | Evidence | Impact | Suggested action |
|----------|---------|----------|--------|------------------|
| High | Planned runtime capabilities are mostly documentation-only today | `README.md`, `docs/architecture.md`, `docs/codebase/.codebase-scan.txt` | Users may overestimate current readiness | Keep README explicit about current stage; implement next milestone behind tests |
| Medium | No CI pipeline is present | `docs/codebase/.codebase-scan.txt` | `go test ./...` and formatting can be skipped accidentally | Add a minimal GitHub Actions workflow once repository hosting is settled |
| Medium | Release/Docker usage is planned but no Dockerfile, Compose, or release workflow exists | `README.md`, `docs/designs/release-distribution.md`, `docs/codebase/.codebase-scan.txt` | Deployment path remains manual or unavailable | Add release workflow and Dockerfiles as separate documented work |
| Medium | API is a route list, not an HTTP implementation | `internal/api/contract.go`, `docs/api.md` | No request/response behavior is executable or testable yet | Implement router with handler tests before adding UI or clients |
| Low | No env/config loading beyond hardcoded defaults | `internal/config/config.go` | Runtime settings cannot be changed without code edits | Add explicit config sources when first runtime server is implemented |

## 2) Technical Debt

| Debt item | Why it exists | Where | Risk if ignored | Suggested fix |
|-----------|---------------|-------|-----------------|---------------|
| Skeleton-only command processes | Project is in initialization stage | `cmd/manager/main.go`, `cmd/worker/main.go` | Commands only print addresses and do not serve APIs | Add server/process lifecycle incrementally with tests |
| Missing persistence schema | SQLite is selected in docs but not implemented | `README.md`, `AGENTS.md` | State model can drift if implementation starts ad hoc | Create schema design and migration tests before persistence code |
| Missing integration boundaries | Docker, OpenVPN, proxy, and node source adapters are not present | `README.md`, `AGENTS.md` | Future code may mix responsibilities | Add separate packages following AGENTS module boundary rules |

## 3) Security Concerns

| Risk | OWASP category | Evidence | Current mitigation | Gap |
|------|----------------|----------|--------------------|-----|
| Future management API has no auth model defined | A01/A07, future risk | `docs/api.md` | No HTTP server exists yet | Define auth or local binding policy before exposing manager API beyond localhost/private network |
| Future OpenVPN/proxy credentials could be mishandled | N/A | `AGENTS.md` | AGENTS forbids committing secrets and private OpenVPN config | No runtime secret-loading or redaction mechanism exists |
| External process and Docker operations need strict timeout/error handling | N/A | `AGENTS.md` | Project rule requires timeouts, errors, logs | No process/container code exists yet |

## 4) Performance and Scaling Concerns

| Concern | Evidence | Current symptom | Scaling risk | Suggested improvement |
|---------|----------|-----------------|-------------|-----------------------|
| Node testing/scoring strategy is undefined | `README.md`, `AGENTS.md` | Not implemented | Naive sequential tests may be slow once node inventory exists | Design bounded concurrency, timeouts, and observable state transitions |
| Logs and SSE stream behavior is undefined | `docs/api.md`, `README.md` | Not implemented | Unbounded buffering or fanout can become unstable | Define event retention and backpressure policy |

## 5) Fragile/High-Churn Areas

| Area | Why fragile | Churn signal | Safe change strategy |
|------|-------------|--------------|----------------------|
| `README.md` | It describes both current skeleton and future product direction | Scan shows 6 recent commits in last 90 days | Cross-check every README claim against source or mark as planned |
| `AGENTS.md` | It is the active project harness for future AI changes | Scan shows 4 recent commits in last 90 days | Read before edits and update related docs/tests with rule changes |

## 6) `[ASK USER]` Questions

1. [ASK USER] Should the next implementation milestone prioritize the HTTP API router, Docker/Release packaging, SQLite schema, or VPNGate node ingestion?
2. [ASK USER] Should the manager API initially bind only to localhost/private networks, or should an authentication model be designed before the first HTTP server?

## 7) Evidence

- `docs/codebase/.codebase-scan.txt`
- `README.md`
- `AGENTS.md`
- `docs/api.md`
- `docs/architecture.md`
- `docs/designs/release-distribution.md`
- `internal/api/contract.go`
- `internal/config/config.go`
