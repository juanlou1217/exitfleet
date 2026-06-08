# 测试计划：Docker Worker 生命周期骨架

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 测试范围

- Docker launcher 命令生成。
- Docker launcher 错误传播。
- Docker worker stop 命令。
- Manager 装配可编译。
- Worker Dockerfile 静态存在。

## 不测试范围

- 当前 Mac 未安装 Docker CLI，不执行真实容器启动。
- 不验证真实 OpenVPN 连接。
- 不验证真实代理转发。
- 不验证 3x-ui/Xray 联动。

## 测试用例

| 用例 | 输入/动作 | 期望结果 | 类型 |
| --- | --- | --- | --- |
| 启动 worker 命令 | fake runner + StartWorker | 生成 `docker run -d`，包含 TUN、NET_ADMIN、端口映射和 worker env | 单元测试 |
| Docker 输出为空 | fake runner 返回空 stdout | worker ID 回退到容器名 | 单元测试 |
| Docker 启动失败 | fake runner 返回错误 | StartWorker 返回错误 | 单元测试 |
| 停止 worker | StopWorker(containerID) | 生成 `docker rm -f containerID` | 单元测试 |
| 完整 Go 测试 | `go test ./...` | 全部通过 | 自动测试 |

## 测试命令

```bash
go test ./...
go fmt ./...
go vet ./...
docker build -f Dockerfile.worker -t exitfleet-worker:local .
```

## 执行结果

- 未执行原因：
  - 当前 Mac 执行环境已安装 Docker CLI，但 Colima VM 首次启动长时间停留在 disk image 下载阶段，已终止卡住的 `colima start`。
  - `docker version` 能显示 client，但 daemon 未运行，错误为无法连接 `/var/run/docker.sock`。
  - 因此 `docker build -f Dockerfile.worker -t exitfleet-worker:local .` 未执行。
- 实际结果：
  - `brew install docker colima docker-buildx` 已执行成功。
  - `command -v docker` 返回 `/opt/homebrew/bin/docker`。
  - `colima status` 返回 `colima is not running`。
  - `docker build -f Dockerfile.worker -t exitfleet-worker:local .` 已执行失败，原因是 Docker daemon 未运行，无法连接 `/var/run/docker.sock`。
  - `go test ./...` 已执行成功。
  - `go fmt ./...` 已执行成功。
  - `go vet ./...` 已执行成功。

## 未覆盖风险

- fake runner 测试不能证明 Docker daemon、镜像构建、TUN 设备和容器权限在真实机器上可用。
