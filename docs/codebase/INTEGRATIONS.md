# External Integrations

## 1) Integration Inventory

| System | Type | Purpose | Auth model | Criticality | Evidence |
|--------|------|---------|------------|-------------|----------|
| VPNGate-compatible data source | External API/data feed, planned | Candidate node inventory | `[TODO]` not implemented | High, planned | `README.md`, `AGENTS.md` |
| Docker | Container runtime, planned | Manage one worker container per active exit | `[TODO]` not implemented | High, planned | `README.md`, `AGENTS.md` |
| OpenVPN | External process, planned | Establish VPN tunnel inside worker | External process config; details `[TODO]` | High, planned | `README.md`, `AGENTS.md`, `docs/architecture.md` |
| HTTP/SOCKS5 proxy | Network listener, planned | User-facing proxy inside worker | `[TODO]` not implemented | High, planned | `README.md`, `AGENTS.md` |
| SQLite | Data store, planned | Persist manager state | Local file; details `[TODO]` | Medium, planned | `README.md`, `AGENTS.md` |
| SSE logs | HTTP streaming, planned | Stream manager and worker events | `[TODO]` not implemented | Medium, planned | `README.md`, `docs/api.md` |

No integration wrapper, database connection, HTTP client, Docker SDK dependency, or OpenVPN process invocation is currently implemented in source.

## 2) Data Stores

| Store | Role | Access layer | Key risk | Evidence |
|-------|------|--------------|----------|----------|
| SQLite | Planned manager state persistence | `[TODO]` not implemented | Schema and migrations are not defined yet | `README.md`, `AGENTS.md` |

## 3) Secrets and Credentials Handling

- Credential sources: `[TODO]` none implemented; no env template detected.
- Hardcoding checks: scan found no `.env.example`, `.env.template`, or source env reads; `AGENTS.md` explicitly forbids committing secrets and private OpenVPN config.
- Rotation or lifecycle notes: `[TODO]` not defined.

## 4) Reliability and Failure Behavior

- Retry/backoff behavior: none implemented.
- Timeout policy: `AGENTS.md` requires network, Docker, process, and filesystem operations to have timeouts, error returns, and logs; no such operations exist yet.
- Circuit-breaker or fallback behavior: none implemented.

## 5) Observability for Integrations

- Logging around external calls: no external calls implemented.
- Metrics/tracing coverage: none implemented.
- Missing visibility gaps: planned scheduler state, Docker lifecycle, OpenVPN process state, proxy health, and SSE event semantics still need implementation.

## 6) Evidence

- `README.md`
- `AGENTS.md`
- `docs/api.md`
- `docs/architecture.md`
- `docs/codebase/.codebase-scan.txt`
