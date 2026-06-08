# 技术方案：Docker Worker 生命周期骨架

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 背景

ExitFleet 架构要求一个 active exit 对应一个 worker 容器，manager 不直接承载代理流量。当前 manager 已有多出口调度骨架，但 launcher 还是 fake。需要先用 Docker CLI 打通容器生命周期边界。

## 方案

- 新增 `internal/docker.Launcher` 实现 `manager.WorkerLauncher`。
- `StartWorker` 执行 `docker run -d`：
  - 容器名包含 node ID 和 host proxy port。
  - 添加 `exitfleet.managed=true` 和 `exitfleet.node_id` label。
  - 添加 `--cap-add NET_ADMIN` 和 `--device /dev/net/tun`。
  - 映射 `hostProxyPort:7928`。
  - 注入 worker 内部 proxy host/port 和 TUN env。
- `StopWorker` 执行 `docker rm -f <container>`.
- `cmd/manager` 默认使用 Docker launcher。
- `Dockerfile.worker` 构建 worker 镜像，包含 OpenVPN、iproute2 和 worker 二进制。

## 模块边界

- `internal/docker` 只负责容器生命周期命令，不做节点调度。
- `internal/manager` 仍只依赖 `WorkerLauncher` 接口。
- `cmd/manager` 负责把 env 配置装配成 Docker launcher。
- Worker 容器内部仍由 `cmd/worker` 和 `internal/worker` 管理。

## 数据流

```text
POST /api/v1/exits
  -> manager.StartExit
  -> docker.Launcher.StartWorker
  -> docker run -d -p hostPort:7928 exitfleet-worker:local
  -> manager stores active exit
```

## 接口变化

- 新增 `Dockerfile.worker`。
- 新增 Docker env 配置项。
- Manager 启动出口时会尝试调用 Docker CLI。

## 取舍

- 首轮用 Docker CLI 而不是 Docker SDK，减少依赖和 API 面。
- 测试使用 fake command runner，不依赖本机 Docker。
- 当前 Mac 未安装 Docker 时，manager 仍能启动，但 `POST /api/v1/exits` 会返回 Docker 命令错误。

## 风险

- Mac 上没有 Docker CLI 或 Docker daemon 时无法真实启动 worker。
- 当前 worker 还没有真实 OpenVPN 配置注入和代理转发能力。
- Dockerfile 真实构建需要本机安装 Docker。

## 回滚策略

可把 `cmd/manager` launcher 装配切回 skeleton launcher；`internal/docker` 和 Dockerfile 可保留不影响核心包测试。

## 实施步骤

1. 为 Docker launcher 写 fake runner 单元测试。
2. 实现 Docker launcher。
3. 新增 Docker worker 镜像文件。
4. 更新 env 配置和 manager 装配。
5. 执行 `go fmt ./...`、`go test ./...`、`go vet ./...`。
6. 如果本机有 Docker，再执行 `docker build -f Dockerfile.worker -t exitfleet-worker:local .`。

## 验证方式

- `go test ./...`
- `go fmt ./...`
- `go vet ./...`
- `docker build -f Dockerfile.worker -t exitfleet-worker:local .`，仅在 Docker 可用时执行。
