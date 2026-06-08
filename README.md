# ExitFleet

ExitFleet 是一个使用 Go 重构的容器原生多出口代理网关。它参考
`Guozh1peng/aimili-vpngate` 的产品思路，但不是对原 Python 项目的逐行翻译。

项目的核心目标是把“节点发现、节点测速、出口调度、OpenVPN 连接、代理服务、日志诊断”拆成清晰的 Manager/Worker 架构：

```text
一个活跃出口节点 = 一个 worker 容器
```

## 项目定位

ExitFleet 面向 VPS 上的多出口代理网关场景：

- 从 VPNGate-compatible 数据源拉取候选节点。
- 对节点进行解析、测速、评分和健康检测。
- 按策略启动一个或多个出口节点。
- 每个出口节点通过独立 worker 容器运行。
- 每个 worker 容器内部负责一个 OpenVPN 进程、一个 TUN 设备、一个 HTTP/SOCKS5 代理。
- Manager 提供 Web UI、REST API、SSE 日志流、调度器、Docker 生命周期管理和状态持久化。

第一阶段重点不是做完整功能，而是先把项目边界、接口契约和 harness 约束立住，避免后续重构变成单文件大程序。

## 架构设计

```text
exitfleet-manager
  Web UI
  REST API
  VPNGate 节点池
  节点测速与评分
  出口调度器
  Docker worker 生命周期管理
  SQLite 状态库
  日志与诊断

exitfleet-worker
  OpenVPN 进程
  tun0 设备
  HTTP/SOCKS5 代理
  /healthz 健康检查
  /readyz 就绪检查
```

关键约定：

- Manager 不直接承载用户代理流量。
- Worker 不管理全局节点池和调度策略。
- 候选节点不会创建容器。
- 只有活跃出口才创建 worker 容器。
- Worker 容器内默认使用 `tun0`，因为每个 worker 拥有独立网络命名空间。
- OpenVPN 第一阶段作为外部进程调用，不重写 VPN 协议。

## 目录结构

```text
cmd/
  manager/          Manager 可执行入口
  worker/           Worker 可执行入口
internal/
  api/              REST API 与 worker health 接口契约
  config/           Manager/Worker 配置模型、默认值和校验
  harness/          项目名称、目标和架构硬约束
docs/
  architecture.md   架构设计说明
  api.md            API 草案
AGENTS.md           AI 修改规则和项目 harness
README.md           项目说明
go.mod              Go module 定义
```

## 当前模块

### `cmd/manager`

Manager 入口。当前只加载默认配置并输出监听地址，后续会接入：

- Web UI
- REST API
- 节点拉取
- 调度器
- Docker worker 管理
- SQLite 状态库

### `cmd/worker`

Worker 入口。当前只加载默认配置并输出代理监听地址，后续会接入：

- OpenVPN 进程管理
- TUN 设备检查
- HTTP/SOCKS5 代理
- `/healthz`
- `/readyz`

### `internal/api`

声明第一版接口契约，避免 API 分散在不同模块里。

Manager API：

```text
GET    /api/v1/state
GET    /api/v1/nodes
POST   /api/v1/nodes/refresh
POST   /api/v1/nodes/test
GET    /api/v1/exits
POST   /api/v1/exits
DELETE /api/v1/exits/{id}
GET    /api/v1/logs/stream
```

Worker API：

```text
GET /healthz
GET /readyz
```

### `internal/config`

定义 Manager 和 Worker 的默认配置。

默认值：

```text
Manager 管理端口: 8787
Worker 代理端口: 7928
Worker TUN 设备: tun0
```

### `internal/harness`

声明项目身份和架构硬约束。

当前约束：

```text
ProjectName = ExitFleet
RuntimeRule = One active exit node maps to exactly one worker container.
```

## 技术选型

当前规划：

- 语言：Go
- Manager 路由：轻量 HTTP 路由，优先 `net/http` + 小型路由库
- UI：Go templates + HTMX + Alpine.js + SSE
- 状态库：SQLite
- Worker 管理：Docker SDK
- VPN：OpenVPN 外部进程
- 代理：Go 实现 HTTP CONNECT + SOCKS5
- 日志：结构化日志 + SSE 流式输出

暂不默认引入：

- Gin
- Echo
- React
- Next.js
- Rust 核心实现

## 开发命令

```bash
go test ./...
go fmt ./...
go run ./cmd/manager
go run ./cmd/worker
```

当前机器如果没有安装 Go，会出现：

```text
zsh:1: command not found: go
```

这种情况下需要先安装 Go 工具链后再执行验证。

## GitHub

当前仓库：

```text
https://github.com/juanlou1217/exitfleet
```

GitHub 操作要求使用本地 `gh`：

```bash
gh auth status
git push origin main
git push --force origin main
```

## 与原项目的关系

原项目：

```text
https://github.com/Guozh1peng/aimili-vpngate
```

原项目提供行为参考：

- VPNGate 节点拉取
- OpenVPN 配置解析
- 节点测试
- 多出口代理
- HTTP/SOCKS5 代理
- Web 管理后台
- 日志和诊断

ExitFleet 不会把原 Python 文件复制进新仓库，也不会继续扩展 Python 版本。重构方向是重新设计 Go 的模块边界和容器运行模型。

## 当前阶段

当前项目处于初始化阶段，已经完成：

- Go module 初始化
- Manager/Worker 入口
- 配置模型
- API 契约
- harness 约束
- 架构文档
- API 草案

下一步建议：

1. 安装 Go 工具链并跑通 `go test ./...`。
2. 实现 `internal/api` 的 HTTP router。
3. 设计 SQLite schema。
4. 实现 VPNGate 节点拉取和解析模块。
5. 实现 worker 容器生命周期管理。

## 重要文档

- `AGENTS.md`
- `docs/architecture.md`
- `docs/api.md`
