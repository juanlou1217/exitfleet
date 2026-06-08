# 技术方案：Env 配置加载

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 背景

ExitFleet 后续会以二进制、Docker 或 Compose 方式运行。运行端口、VPNGate 数据源、TUN 设备和 OpenVPN 命令不应写死在入口代码中，也不能提交真实 `.env` 或凭据。

## 方案

- 使用 Go 标准库实现轻量 `.env` 解析，不引入第三方依赖。
- 启动时读取当前目录 `.env`，文件不存在时返回空配置。
- 配置优先级：进程环境变量 > `.env` 文件 > Go 默认值。
- 提交 `.env.example`，忽略真实 `.env`。

## 模块边界

- `internal/config` 负责解析、默认值、环境变量覆盖和校验。
- `cmd/manager` 只调用 config loader 并装配依赖。
- `cmd/worker` 只调用 config loader 并装配 worker。

## 数据流

```text
.env.example -> user .env
process environment
  -> internal/config LoadEnvFile + LoadManager/LoadWorker
  -> cmd/manager or cmd/worker dependency wiring
```

## 接口变化

- 新增 `.env.example`。
- 新增 `.gitignore` 忽略真实 `.env`。
- 新增 `LoadEnvFile`、`LoadManager`、`LoadWorker`。

## 取舍

- 不引入 `godotenv`，减少依赖面。
- `.env` 只支持常见 `KEY=value` 和整行注释，不实现复杂 shell expansion。
- 进程环境变量优先，方便容器和 CI 覆盖。

## 风险

- `.env` 解析不是完整 shell 语法，复杂值需要后续增强。
- 当前 `.env` 路径固定为工作目录 `.env`，服务化运行时需要确保 working directory 正确。

## 回滚策略

入口可退回默认配置构造；`.env.example` 和 `.gitignore` 可保留，不影响运行。

## 实施步骤

1. 增加 config loader 测试。
2. 实现 `.env` 解析和环境变量覆盖。
3. 更新 manager/worker 入口。
4. 增加 `.env.example` 和 `.gitignore`。
5. 执行 `go fmt ./...`、`go test ./...`、`go vet ./...`。

## 验证方式

- `go test ./...`
- `go fmt ./...`
- `go vet ./...`
