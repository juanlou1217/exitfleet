# 功能记录：Env 配置加载

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 背景

首轮迁移骨架中 manager 默认写死 VPNGate URL 和 proxy base port，worker 默认配置也只能从代码默认值获得。后续接 Docker、OpenVPN 和部署产物前，需要把运行配置移到 `.env` 和进程环境变量中。

## 目标

- 支持从 `.env` 文件读取 manager 和 worker 配置。
- 支持进程环境变量覆盖 `.env` 文件。
- 提供 `.env.example` 说明可配置项。
- 保留当前默认值，避免没有 `.env` 时无法启动。

## 非目标

- 不引入第三方 dotenv 依赖。
- 不提交真实 `.env`、密钥、代理凭据或私有 OpenVPN 配置。
- 不实现完整配置热更新。

## 用户场景

- 开发者复制 `.env.example` 后调整管理端口、VPNGate 数据源、worker 代理端口和 TUN 设备。
- CI 或容器运行时直接通过进程环境变量覆盖 `.env`。

## 功能范围

- `internal/config` 增加 `.env` 解析、manager/worker 配置加载和校验。
- `cmd/manager` 和 `cmd/worker` 启动时读取 `.env`。
- `.gitignore` 忽略真实 `.env` 文件。

## 接口影响

- 新增环境变量：
  - `EXITFLEET_MANAGER_HOST`
  - `EXITFLEET_MANAGER_PORT`
  - `EXITFLEET_VPNGATE_URL`
  - `EXITFLEET_PROXY_BASE_PORT`
  - `EXITFLEET_WORKER_PROXY_HOST`
  - `EXITFLEET_WORKER_PROXY_PORT`
  - `EXITFLEET_WORKER_TUN_DEVICE`
  - `EXITFLEET_OPENVPN_CMD`
  - `EXITFLEET_OPENVPN_AUTH_FILE`

## 数据模型影响

- Manager 配置增加 `VPNGateURL` 和 `ProxyBasePort`。
- Worker 配置增加 `OpenVPNCommand` 和 `OpenVPNAuthFile`。

## 验收标准

- 没有 `.env` 时使用现有默认值。
- `.env` 文件能覆盖默认值。
- 进程环境变量优先于 `.env`。
- 非法端口会返回配置错误。

## 相关文档

- `.env.example`
- `docs/designs/2026-06-08-env-configuration.md`
- `docs/tests/2026-06-08-env-configuration.md`
