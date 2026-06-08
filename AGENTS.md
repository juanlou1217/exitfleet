# AGENTS.md

## 项目概况

ExitFleet 是一个使用 Go 重构的容器原生多出口代理网关。它参考 `Guozh1peng/aimili-vpngate` 的产品思路，但不逐行翻译原 Python 实现。

项目目标：

- 管理 VPNGate-compatible 节点池。
- 自动拉取、解析、测速、评分节点。
- 通过 Docker worker 管理活跃出口。
- 每个活跃出口独立运行 OpenVPN、TUN、HTTP/SOCKS5 代理和健康检查。
- 提供轻量 Web UI、REST API、SSE 日志流、系统诊断和 SQLite 状态持久化。

核心架构约束：

- 一个活跃出口节点必须对应一个 worker 容器。
- Manager 不直接承载用户代理流量。
- Worker 只负责单个出口，不管理全局调度。
- 原 Python 项目只作为行为参考，不作为新项目目录的一部分。

## 目录速查

```text
cmd/
  manager/          Manager 可执行入口
  worker/           Worker 可执行入口
internal/
  api/              REST/health 接口契约
  config/           Manager/Worker 配置、默认值、校验
  harness/          项目目标与架构硬约束
docs/
  architecture.md   架构设计
  api.md            API 草案
  features/         功能开发记录
  designs/          技术方案记录
  tests/            测试方案与验证记录
  decisions/        架构决策记录
  templates/        AI 生成文档模板
AGENTS.md           AI 修改约束与项目导航
README.md           项目说明
go.mod              Go module 定义
```

## 代码模块总览 【功能描述】

### `cmd/`

技术栈概览：

- Go 标准库
- 每个子目录一个可执行程序
- 只做启动、配置加载、依赖装配

目录结构详解：

```text
cmd/manager/main.go  启动 manager，后续接入 Web UI、REST API、调度器
cmd/worker/main.go   启动 worker，后续接入 OpenVPN、代理、健康检查
```

### `internal/api/`

技术栈概览：

- Go 类型定义
- REST API 契约
- 第一阶段不绑定具体 HTTP 框架

目录结构详解：

```text
internal/api/contract.go       Manager/Worker 路由契约
internal/api/contract_test.go  接口契约测试
```

功能描述：

- 统一声明 `/api/v1` 管理接口。
- 统一声明 worker 的 `/healthz` 和 `/readyz`。
- 防止接口散落在不同 package 中。

### `internal/config/`

技术栈概览：

- Go 标准库
- 显式配置类型
- 单元测试覆盖默认端口和校验规则

目录结构详解：

```text
internal/config/config.go       Manager/Worker 配置模型
internal/config/config_test.go  默认配置和校验测试
```

功能描述：

- 默认管理端口：`8787`。
- 默认 worker 代理端口：`7928`。
- 默认 worker TUN 设备：`tun0`。

### `internal/harness/`

技术栈概览：

- Go 常量
- 架构硬约束
- 测试确保项目身份和关键规则存在

目录结构详解：

```text
internal/harness/harness.go       项目名称、目标、运行时规则
internal/harness/harness_test.go  harness 约束测试
```

功能描述：

- 明确项目名 `ExitFleet`。
- 明确项目目标。
- 明确一个活跃出口对应一个 worker 容器。

### `docs/`

技术栈概览：

- Markdown
- 架构、接口、功能、方案、测试和决策文档
- 所有 AI 生成的重要开发产物必须落盘记录

目录结构详解：

```text
docs/architecture.md             Manager/Worker 架构和运行模型
docs/api.md                      REST API 和 worker health API 草案
docs/features/                   功能开发记录，一项功能一个 Markdown
docs/designs/                    技术方案记录，一项方案一个 Markdown
docs/tests/                      测试计划、测试用例、验证结果
docs/decisions/                  ADR 架构决策记录
docs/templates/feature.md        功能记录模板
docs/templates/technical-plan.md 技术方案模板
docs/templates/test-plan.md      测试计划模板
docs/templates/decision.md       架构决策模板
```

功能描述：

- 记录每个新功能的目标、范围、接口、数据流和验收标准。
- 记录技术方案、取舍、风险、回滚策略和后续任务。
- 记录测试用例、测试命令、验证结果和未覆盖风险。
- 记录会影响架构边界的决策，避免口头约定丢失。

## 关键业务约定

- 项目名固定为 `ExitFleet`。
- Go module 固定为 `github.com/juanlou1217/exitfleet`。
- 第一阶段使用 Go 重构 manager/worker 边界，不迁移完整 VPNGate 行为。
- 一个活跃出口节点对应一个 worker 容器。
- 候选节点不创建容器。
- Manager 负责节点池、调度、Docker 生命周期、状态库、日志聚合和 UI。
- Worker 负责单个 OpenVPN 进程、单个 TUN、单个代理监听、单个健康检查。
- Worker 容器内 TUN 设备默认叫 `tun0`。
- OpenVPN 第一阶段作为外部进程调用，不重写 VPN 协议。
- 默认管理端口是 `8787`。
- 默认代理端口从 `7928` 开始。
- UI 第一阶段优先 Go templates + HTMX + Alpine.js + SSE，不默认使用 React/Next.js。
- 持久化第一选择 SQLite。
- 网络、Docker、进程、文件系统操作必须有超时、错误返回和日志。

## 常用命令

```bash
go test ./...
go fmt ./...
go run ./cmd/manager
go run ./cmd/worker
```

GitHub 操作必须使用本地 `gh`：

```bash
gh auth status
gh repo create exitfleet --private --source=. --remote=origin --push
git push --force origin main
```

## AI 修改规则

- 修改前必须阅读本文件和相关目录。
- 新行为先写测试，纯文档和机械配置除外。
- 新功能开发前必须在 `docs/features/` 新建或更新功能记录。
- 涉及架构、模块边界、依赖、运行模型、接口设计时，必须在 `docs/designs/` 新建或更新技术方案。
- 涉及长期架构取舍时，必须在 `docs/decisions/` 新建 ADR。
- 新增或修改行为后，必须在 `docs/tests/` 记录测试计划、测试用例、执行结果和未覆盖风险。
- AI 生成的方案、接口、测试用例、验证结论不能只停留在聊天中，必须同步进项目文档。
- 文档命名使用 `YYYY-MM-DD-short-topic.md`，短标题用小写英文和连字符。
- 不允许把 Manager、Worker、Proxy、Docker 调度、OpenVPN 进程管理揉进一个大 package。
- 不允许把项目重新写成单文件程序。
- 不允许把原 Python 项目复制进本仓库。
- 不允许默认引入 Gin、Echo、React、Next.js 或其他大型框架。
- 不允许重写 VPN 协议；OpenVPN 作为外部进程保留到架构稳定后再评估。
- 不允许提交密钥、token、真实代理凭据、私有 OpenVPN 配置。
- 长循环必须支持退出信号。
- 调度状态必须可观测、可测试、可持久化。

## 验证要求

最低验证：

```bash
go test ./...
go fmt ./...
```

涉及 Docker worker 时追加：

```bash
docker build -f Dockerfile.manager .
docker build -f Dockerfile.worker .
```

涉及 OpenVPN 或代理行为时追加：

- 使用 fake OpenVPN 或可控测试进程做集成测试。
- 覆盖命令构建、端口分配、状态转换、超时和错误路径。
- 默认测试不依赖真实 VPNGate 节点连通性。

如果当前机器缺少 `go`、Docker 或网络权限，最终回复必须明确说明哪些验证没有执行。

## AI 开发记录要求

每次进入实质开发前，AI 必须判断本次改动属于哪类记录：

```text
功能行为变化      -> docs/features/
技术方案或模块边界 -> docs/designs/
测试计划和结果    -> docs/tests/
架构长期决策      -> docs/decisions/
```

记录要求：

- 功能记录必须包含目标、范围、非目标、相关接口、数据模型、验收标准。
- 技术方案必须包含背景、方案、取舍、风险、回滚策略、实施步骤。
- 测试记录必须包含测试用例、测试命令、实际结果、未执行原因、剩余风险。
- ADR 必须包含状态、背景、决策、后果、替代方案。
- 如果实现和文档不一致，必须先更新文档再完成最终回复。
- 如果只是修正错字或格式，可以不新增记录，但不能违反已有记录。

## 重要文档

- `AGENTS.md`
- `docs/architecture.md`
- `docs/api.md`
- `docs/features/`
- `docs/designs/`
- `docs/tests/`
- `docs/decisions/`
- 参考仓库：`https://github.com/Guozh1peng/aimili-vpngate`
- 原项目代码不属于本仓库；需要对照时单独克隆，不要复制进 ExitFleet。
