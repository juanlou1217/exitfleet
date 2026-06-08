# ExitFleet Architecture

ExitFleet is a clean Go redesign of the AimiliVPN/VPNGate gateway idea. The
project does not translate the original Python files line by line. It keeps the
useful product behavior and replaces the runtime model with clearer manager and
worker boundaries.

## Runtime Model

```text
exitfleet-manager
  Web UI
  REST API
  VPNGate node inventory
  node scoring and scheduling
  Docker worker lifecycle
  SQLite state
  logs and diagnostics

exitfleet-worker
  one OpenVPN process
  one tun0 device inside the container namespace
  one HTTP/SOCKS5 proxy listener
  /healthz and /readyz endpoints
```

One active exit node maps to exactly one worker container. Candidate nodes stay
in the manager inventory until the scheduler starts them.

## Interface Boundaries

Manager-facing API routes live under `/api/v1`. Worker health routes stay small
and local to each worker.

- `GET /api/v1/state`
- `GET /api/v1/nodes`
- `POST /api/v1/nodes/refresh`
- `POST /api/v1/nodes/test`
- `GET /api/v1/exits`
- `POST /api/v1/exits`
- `DELETE /api/v1/exits/{id}`
- `GET /api/v1/logs/stream`
- `GET /healthz`
- `GET /readyz`

## First Milestone

The first milestone is a compiling project skeleton with stable package
boundaries, explicit API contracts, and harness rules strong enough for future
AI-assisted changes to stay inside the intended architecture.

