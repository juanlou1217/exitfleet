# 功能记录：Docker Worker 生命周期骨架

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 背景

首轮迁移已经具备多出口调度模型，但 manager 仍使用 fake worker launcher。下一步需要把 active exit 映射到真实 Docker worker 容器，为后续 OpenVPN、TUN、HTTP/SOCKS5 代理和 3x-ui/Xray upstream 对接做准备。

## 目标

- 新增 Docker worker launcher，使用 Docker CLI 启动/停止 worker 容器。
- 新增 `Dockerfile.worker`，构建可运行 worker 镜像。
- Manager 默认使用 Docker launcher。
- 保持测试不依赖本机 Docker daemon。

## 非目标

- 不要求当前 Mac 真实启动容器。
- 不实现真实 OpenVPN 连接。
- 不实现真实代理转发。
- 不直接修改 3x-ui/Xray 配置。

## 用户场景

- 开发者构建 `exitfleet-worker:local` 镜像。
- Manager 启动一个 active exit 时执行 `docker run -d`，把 host 动态端口映射到 worker 内部 `7928`。
- Manager 停止一个 active exit 时执行 `docker rm -f`。

## 功能范围

- `internal/docker`：Docker CLI launcher 和命令 runner。
- `cmd/manager`：使用 Docker launcher 替代 skeleton launcher。
- `Dockerfile.worker`：worker 镜像构建文件。
- `.env.example`：Docker 命令、镜像、容器名前缀和网络配置。

## 接口影响

- 新增环境变量：
  - `EXITFLEET_DOCKER_CMD`
  - `EXITFLEET_DOCKER_WORKER_IMAGE`
  - `EXITFLEET_DOCKER_CONTAINER_PREFIX`
  - `EXITFLEET_DOCKER_NETWORK`

## 数据模型影响

- `manager.WorkerRecord.ID` 现在记录 Docker 返回的 container ID；若 Docker 输出为空则回退到容器名。

## 验收标准

- Docker launcher 生成包含 `--cap-add NET_ADMIN`、`--device /dev/net/tun` 和端口映射的 `docker run` 命令。
- Stop worker 使用 `docker rm -f`。
- Docker CLI 错误能返回给 manager。
- 单元测试不需要 Docker daemon。

## 相关文档

- `docs/designs/2026-06-08-docker-worker-lifecycle.md`
- `docs/tests/2026-06-08-docker-worker-lifecycle.md`
