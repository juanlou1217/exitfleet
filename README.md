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
  SOCKS5 代理端口，默认 7928
  /healthz 健康检查端口，默认 8790
  /readyz 就绪检查端口，默认 8790
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

Worker 入口。当前会加载配置，启动 SOCKS5 代理 listener 和独立 health
listener。真实 OpenVPN 配置注入仍在后续阶段接入。

- SOCKS5 代理端口默认 `7928`。
- Health/ready HTTP 端口默认 `8790`。
- outbound IP 为空时默认 dialer 拒绝直连，避免未接 VPN 时泄漏直连流量。

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
Worker SOCKS5 代理端口: 7928
Worker health 端口: 8790
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

## 使用方式

ExitFleet 面向普通使用者时，不要求在机器上安装 Go 工具链。后续发布版本应优先提供两类产物：

```text
预编译二进制
Docker 镜像
```

推荐使用方式：

```bash
# 方式一：下载 Release 中的预编译二进制
exitfleet-manager

# 方式二：使用 Docker/Compose 启动 manager 和 worker
docker compose up -d
```

当前项目仍处于初始化阶段，正式的 Release、Dockerfile 和 Compose 文件还没有建立。发布能力会作为后续功能单独设计和实现。

## 开发命令

以下命令只面向开发者和 CI。普通部署应使用 Release 产物或 Docker 镜像。

```bash
go test ./...
go fmt ./...
go run ./cmd/manager
go run ./cmd/worker
```

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

1. 建立 Release 打包方案，支持预编译二进制。
2. 增加 Dockerfile 和 Compose 示例。
3. 实现 `internal/api` 的 HTTP router。
4. 设计 SQLite schema。
5. 实现 VPNGate 节点拉取和解析模块。
6. 实现 worker 容器生命周期管理。

## 重要文档

- `AGENTS.md`
- `docs/architecture.md`
- `docs/api.md`
