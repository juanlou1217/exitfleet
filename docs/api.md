# ExitFleet API Draft

## Manager API

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/state` | Return manager, scheduler, node, and worker summary state. |
| `GET` | `/api/v1/nodes` | List candidate and tested VPNGate-compatible nodes. |
| `POST` | `/api/v1/nodes/refresh` | Fetch and parse a fresh node inventory. |
| `POST` | `/api/v1/nodes/test` | Test selected nodes or a scheduler-selected batch. |
| `GET` | `/api/v1/exits` | List active exit workers. |
| `POST` | `/api/v1/exits` | Start a worker for one selected exit node. |
| `DELETE` | `/api/v1/exits/{id}` | Stop and remove one active exit worker. |
| `GET` | `/api/v1/logs/stream` | Stream manager and worker events over SSE. |

## Worker API

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/healthz` | Process liveness. |
| `GET` | `/readyz` | OpenVPN and proxy readiness. |

