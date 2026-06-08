# 技术方案：worker-proxy-health

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 背景

ExitFleet 当前已经有 Manager 多出口调度、Docker worker launcher 和 OpenVPN 状态骨架，但 worker 内部还缺少可运行的代理与健康检查闭环。`xiaowen-king/vpngate-proxy` 对单容器 OpenVPN + SOCKS5 + health 有参考价值；ExitFleet 本轮只吸收行为边界，不采用其 Python 单体结构。

## 方案

- 新增 `internal/proxy` 独立包实现最小 SOCKS5 server。
- SOCKS5 默认只支持无认证 CONNECT，支持 IPv4 和域名目标。
- 默认 dialer 必须收到 outbound IP 才会拨号，防止 worker 尚未获得 TUN IP 时直连公网。
- 新增 `worker.HealthChecker`，支持多 URL 和失败阈值。
- Worker 的 `/readyz` 同时受 OpenVPN 状态和健康检查阈值约束。
- Worker 进程拆成两个 listener：
  - SOCKS5 proxy：默认 `7928`
  - health HTTP：默认 `8790`
- Docker launcher 把 worker env 注入容器；当前只映射 proxy 端口到宿主机，health 端口用于容器内部 healthcheck。

## 模块边界

- `internal/proxy` 只负责 SOCKS5 协议和 outbound dial，不知道 OpenVPN、节点池或调度。
- `internal/worker` 只编排单 worker 状态和健康检查，不管理多出口。
- `internal/manager` 不直接依赖 `internal/proxy`，仍通过 `WorkerLauncher` 启动 worker。
- `internal/docker` 只负责 Docker CLI 参数和 env 注入。
- `cmd/worker` 负责装配 proxy server、health checker 和 HTTP health server。

## 数据流

```text
client
  -> host proxy port
  -> worker container :7928 SOCKS5
  -> proxy.Dialer(outboundIP)
  -> later: tun0 VPN source IP

Docker healthcheck
  -> worker container :8790 /healthz

GET /readyz
  -> worker.Status
  -> OpenVPN ready && health checker below failure threshold
```

## 接口变化

- 新增 `internal/proxy` Go package。
- 新增 worker status 字段 `health_failures`。
- 新增 worker env 配置项，见 `docs/features/worker-proxy-health.md`。
- `Dockerfile.worker` healthcheck 从 `127.0.0.1:7928/healthz` 改为 `127.0.0.1:8790/healthz`。

## 取舍

- 本轮不引入第三方 SOCKS5 库，先实现可测试的最小 CONNECT 行为。
- 本轮不实现认证，worker 代理端口仍应由 Manager/Docker 暴露策略控制。
- 本轮不自动探测 TUN IP；`EXITFLEET_WORKER_PROXY_OUTBOUND_IP` 留作后续 OpenVPN ready 后注入或动态更新。
- Health checker 先提供状态能力，Manager 自动切换策略后续接入。

## 风险

- SOCKS5 当前只覆盖 CONNECT、IPv4 和域名，不支持 UDP ASSOCIATE、IPv6 和认证。
- outbound IP 为空时代理会拒绝转发，真实 OpenVPN 接线完成前不能作为可用出口。
- Docker daemon 当前 Mac 环境仍未启动，Dockerfile 构建未在本轮验证。
- Manager 尚未消费 worker health 状态做自动切换。

## 回滚策略

可回滚 `cmd/worker` 中 proxy listener 装配和 `internal/proxy` 包；Manager、node pool 和 Docker launcher 主流程仍可保留已有骨架。若 health 端口拆分影响部署，可临时只运行 `/healthz`/`/readyz` HTTP server。

## 实施步骤

1. 先写 SOCKS5、HealthChecker、env 和 Docker launcher 单元测试。
2. 实现 `internal/proxy`。
3. 实现 `worker.HealthChecker` 并接入 Worker status。
4. 扩展 `internal/config` 和 `.env.example`。
5. 扩展 Docker launcher env 注入。
6. 调整 `cmd/worker` 和 `Dockerfile.worker`。
7. 更新文档并运行验证命令。

## 验证方式

- `go fmt ./...`
- `go test ./...`
- Docker 构建在 Docker daemon 可用后再执行：
  - `docker build -f Dockerfile.worker -t exitfleet-worker:local .`
