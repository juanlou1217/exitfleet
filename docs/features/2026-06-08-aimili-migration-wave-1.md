# 功能记录：AimiliVPN 首轮迁移骨架

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 背景

ExitFleet 以 `Guozh1peng/aimili-vpngate` 的产品行为为参考，使用 Go 重构为 Manager/Worker 架构。首轮迁移目标是建立可运行、可测试的骨架，不复制上游 Python 源码，也不要求真实 VPN 连通。

## 目标

- 建立 VPNGate-compatible 节点拉取、解析、测速的 Go 包边界。
- 建立 worker OpenVPN 命令构建和状态机骨架。
- 建立 manager 内存状态、节点刷新、节点测试、出口启动/停止骨架。
- 建立 `net/http` REST API 和 worker health/readiness 端点。
- 使用 fake runner、fake source 和 fake launcher 完成测试闭环。

## 非目标

- 不复制上游 Python 文件到本仓库。
- 不实现真实 Docker worker 生命周期。
- 不要求真实 OpenVPN、VPNGate 或代理网络连通。
- 不复刻上游生产 UI。
- 不实现系统安装脚本或 systemd 服务。

## 用户场景

- 开发者可以运行 manager，看到 `/api/v1` 骨架接口。
- 开发者可以用 fake 节点源验证刷新、测试、启动出口和停止出口流程。
- 开发者可以运行 worker health/readiness handler 的测试，确认单 worker 状态可观测。

## 功能范围

- `internal/node`：节点模型、VPNGate CSV 解析、OpenVPN 配置解码、节点测速接口。
- `internal/worker`：OpenVPN command builder、runner 接口、worker 状态机、health/readiness handler。
- `internal/manager`：内存 store、调度骨架、日志事件、fake worker launcher 适配。
- `internal/api`：基于标准库 `net/http` 的 manager REST handler。

## 接口影响

- 保持现有 `/api/v1` route contract。
- 新增可执行 handler 覆盖：
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

## 数据模型影响

- 节点模型包含 ID、IP、国家、主机名、score、ping、speed、OpenVPN config text、probe status 和 last error。
- 出口模型包含 exit ID、node ID、worker ID、proxy port 和状态。
- 日志事件包含时间、级别、模块和消息。

## 验收标准

- 候选节点不会创建 worker。
- 只有 `StartExit(nodeID)` 会创建一个 active exit 和对应 worker record。
- `StopExit(exitID)` 会释放 active exit。
- OpenVPN 命令构建覆盖 tun0、route-nopull 和 auth-user-pass。
- HTTP API 使用 fake source/launcher 能完成刷新节点、启动出口、停止出口、SSE 输出。
- 所有新增行为包有同目录 `*_test.go`。

## 相关文档

- `docs/designs/2026-06-08-aimili-migration-wave-1.md`
- `docs/tests/2026-06-08-aimili-migration-wave-1.md`
- `AGENTS.md`
