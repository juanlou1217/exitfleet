# 技术方案：AimiliVPN 首轮迁移骨架

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 背景

上游 `Guozh1peng/aimili-vpngate` 是 Python 单进程实现，核心行为包含 VPNGate 节点获取、OpenVPN 进程控制、策略路由、代理服务、管理 API/UI 和日志诊断。ExitFleet 的架构约束要求 Go 重构后保持 Manager/Worker 边界，不能把 Manager、Worker、Proxy、Docker 调度和 OpenVPN 进程管理揉进一个大 package。

## 方案

首轮按行为域拆成四个 Go 模块：

- `internal/node` 负责节点模型、VPNGate CSV 解析和测速抽象。
- `internal/worker` 负责单 worker 的 OpenVPN 命令、runner 接口、状态机和 health/readiness。
- `internal/manager` 负责节点池、出口调度、内存状态和 worker launcher 抽象。
- `internal/api` 负责标准库 HTTP handler，把已有 route contract 接到 manager/worker 能力。

首轮所有外部依赖使用接口隔离：节点源、测速 runner、OpenVPN runner、worker launcher 均可 fake。后续再把 fake 替换为真实 VPNGate、Docker、OpenVPN 和代理实现。

## 模块边界

- Manager 不承载用户代理流量。
- Worker 只负责单个出口，不管理全局节点池。
- Candidate node 只进入 manager inventory，不创建 worker。
- 一个 active exit 对应一个 worker record。
- OpenVPN 第一阶段作为外部命令，不重写 VPN 协议。

## 数据流

```text
VPNGate source
  -> internal/node ParseVPNGateCSV
  -> internal/manager Store candidates
  -> internal/node ProbeNode
  -> internal/manager StartExit
  -> WorkerLauncher
  -> active exit + log event
```

## 接口变化

- `internal/api` 从 route contract 扩展为可执行 `net/http` handler。
- `cmd/manager` 从只打印地址改为启动 manager HTTP server。
- `cmd/worker` 从只打印地址改为启动 worker health/readiness server。

## 取舍

- 选择标准库 `net/http`，避免首轮引入 Gin/Echo。
- 选择内存 store，避免过早设计 SQLite schema。
- 选择 fake runner/launcher，避免测试依赖真实 Docker、OpenVPN 和 VPNGate 网络。
- 暂不实现公网认证；生产暴露前必须补认证或反代保护。

## 风险

- 当前 skeleton 只能验证边界和 API，不代表真实 VPN 链路可用。
- 后续接 Docker 和 OpenVPN 时需要补充超时、日志、权限和系统命令测试。
- API 暂无认证，不适合直接公网暴露。

## 回滚策略

新增模块独立于现有 config/harness/api contract。若首轮实现有问题，可回滚 `internal/node`、`internal/worker`、`internal/manager`、`internal/api` handler 扩展和 cmd wiring，保留原 skeleton。

## 实施步骤

1. 新增功能、方案、测试记录。
2. 先写 `internal/node` 测试，再实现解析和测速骨架。
3. 先写 `internal/worker` 测试，再实现命令构建和状态机。
4. 先写 `internal/manager` 测试，再实现 store 和调度骨架。
5. 先写 `internal/api` handler 测试，再实现 manager API 和 worker health API。
6. 更新 `cmd/manager` 和 `cmd/worker` 启动 HTTP server。
7. 执行 `go test ./...` 和 `go fmt ./...`。

## 验证方式

- `go test ./...`
- `go fmt ./...`
- Docker/OpenVPN 真实构建和真实连通不属于首轮验证。
