# 功能记录：worker-proxy-health

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 背景

ExitFleet 的 worker 需要逐步从 OpenVPN 状态骨架演进到真实出口运行时。参考 `xiaowen-king/vpngate-proxy` 后，本轮吸收其 SOCKS5 代理、健康检查阈值和可配置运行参数思路，但不复制 Python 源码，也不改变 ExitFleet 的 Manager/Worker 多出口边界。

## 目标

- 新增 Go SOCKS5 代理骨架，支持 CONNECT、域名/IPv4 目标、最大连接数和 outbound IP 绑定。
- Worker 增加健康检查器，支持多 URL、任一成功即健康、连续失败阈值。
- 将 worker proxy 和 worker health 拆成两个监听地址，避免 SOCKS5 与 HTTP health 争用端口。
- 将健康检查、SOCKS5 并发、outbound IP 等参数放入 env 配置。
- Docker worker launcher 注入新增 worker 运行参数。

## 非目标

- 不实现真实 VPNGate 节点自动切换。
- 不实现 3x-ui API 注册。
- 不实现 HTTP CONNECT 代理。
- 不让 Manager 承载用户代理流量。
- 不复制 `xiaowen-king/vpngate-proxy` 的 Python 实现。

## 用户场景

- 用户启动多个 active exits 后，每个 worker 容器都有独立 SOCKS5 代理端口。
- 运维可以通过 `.env` 调整健康检查 URL、失败阈值、检查间隔、超时和 SOCKS5 最大并发。
- Worker 的 `/readyz` 可以在健康失败达到阈值后返回不可用。

## 功能范围

- `internal/proxy`：SOCKS5 server、默认 dialer、outbound IP 保护。
- `internal/worker`：HealthChecker、状态暴露、ready 状态联动。
- `internal/config`：Worker health/proxy env 配置。
- `internal/docker`：Docker run env 注入。
- `cmd/worker`：SOCKS5 listener 和 health listener 分离启动。
- `Dockerfile.worker`：暴露 7928 代理端口和 8790 health 端口。

## 接口影响

- Worker SOCKS5 默认监听：`0.0.0.0:7928`。
- Worker health 默认监听：`0.0.0.0:8790`。
- Worker HTTP API 保持：
  - `GET /healthz`
  - `GET /readyz`

## 数据模型影响

- Worker status 增加 `health_failures`。
- 新增 worker env：
  - `EXITFLEET_WORKER_PROXY_OUTBOUND_IP`
  - `EXITFLEET_WORKER_SOCKS_MAX_CONNECTIONS`
  - `EXITFLEET_WORKER_HEALTH_HOST`
  - `EXITFLEET_WORKER_HEALTH_PORT`
  - `EXITFLEET_WORKER_HEALTH_URLS`
  - `EXITFLEET_WORKER_HEALTH_FAIL_THRESHOLD`
  - `EXITFLEET_WORKER_HEALTH_INTERVAL_SECONDS`
  - `EXITFLEET_WORKER_HEALTH_TIMEOUT_SECONDS`

## 验收标准

- SOCKS5 CONNECT 会把目标地址和 outbound IP 交给 dialer。
- outbound IP 为空时默认 dialer 拒绝直连，避免未接 VPN 时泄漏直连流量。
- 健康检查任一 URL 成功即清零失败次数。
- 连续失败达到阈值后 Worker ready 状态下沉为 failed。
- Docker launcher 注入新增 worker env。
- `go fmt ./...` 和 `go test ./...` 通过。

## 相关文档

- `docs/designs/worker-proxy-health.md`
- `docs/tests/worker-proxy-health.md`
- `docs/architecture.md`
- `docs/api.md`
