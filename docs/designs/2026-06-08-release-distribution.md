# 技术方案：发布产物与使用方式

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 背景

README 之前偏向开发机视角，提到本机没有 Go 工具链时需要先安装 Go。这个表述不适合公开项目的使用者。ExitFleet 作为网关工具，普通部署应优先使用预编译二进制或 Docker 镜像，开发者和 CI 才需要 Go 工具链。

## 方案

README 中区分两类路径：

- 使用者路径：下载 Release 中的预编译二进制，或使用 Docker/Compose。
- 开发者路径：使用 `go test`、`go fmt`、`go run`。

当前项目还没有正式 Release、Dockerfile 和 Compose 文件，因此 README 只声明发布目标和推荐方向，不宣称已有可用产物。

## 模块边界

- `README.md` 说明公开使用方式和当前阶段。
- 后续打包实现应放入 `deploy/`、`.github/workflows/` 或相关构建脚本。
- Go 源码模块不承担发布说明职责。

## 数据流

```text
source code
  -> CI build
  -> release binaries
  -> docker images
  -> user deployment
```

## 接口变化

本次不改变运行时 API。

## 取舍

- 不要求普通用户安装 Go，降低部署门槛。
- 暂不虚构已经存在的 Release 和 Docker 镜像，避免 README 误导。

## 风险

- README 提到的发布方式需要后续真正实现。
- 如果长期没有 Release/镜像，公开说明会停留在规划状态。

## 回滚策略

如果短期内只支持源码运行，可以把 README 的“使用方式”降级为“发布规划”，但仍不应把本机缺少 Go 的错误作为用户安装说明。

## 实施步骤

1. 删除 README 中“没 Go 就安装 Go”的公开用户说明。
2. 删除 README 中 GitHub 操作和原项目关系段落。
3. 增加使用者优先的 Release/Docker 说明。
4. 后续补充 Release workflow、Dockerfile 和 Compose。

## 验证方式

- 检查 README 不再包含 `zsh:1: command not found: go`。
- 检查 README 不再包含 `gh auth status` 或 `git push --force`。
- 检查 README 包含预编译二进制和 Docker 使用方向。

