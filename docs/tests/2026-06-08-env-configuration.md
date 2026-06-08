# 测试计划：Env 配置加载

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 测试范围

- `.env` 文件解析。
- Manager 配置默认值、文件覆盖、进程环境变量覆盖。
- Worker 配置默认值和文件覆盖。
- 非法端口错误。

## 不测试范围

- Docker/Compose 注入环境变量。
- systemd 工作目录配置。
- 复杂 shell 风格 `.env` expansion。

## 测试用例

| 用例 | 输入/动作 | 期望结果 | 类型 |
| --- | --- | --- | --- |
| 解析 `.env` | `KEY=value`、空行、注释 | 返回 env map | 单元测试 |
| Manager 文件覆盖 | 设置 host、port、VPNGate URL、proxy base port | 配置字段被覆盖 | 单元测试 |
| Worker 文件覆盖 | 设置 proxy host/port、tun、OpenVPN command/auth | 配置字段被覆盖 | 单元测试 |
| 进程 env 优先 | `t.Setenv` 覆盖同名 `.env` 值 | 使用进程 env | 单元测试 |
| 非法端口 | port=`bad` | 返回配置错误 | 单元测试 |

## 测试命令

```bash
go test ./...
go fmt ./...
go vet ./...
```

## 执行结果

- 未执行原因：无。
- 实际结果：
  - `go test ./...` 已执行成功。
  - `go fmt ./...` 已执行成功。
  - `go vet ./...` 已执行成功。

## 未覆盖风险

- 当前没有真实部署环境验证 `.env` 工作目录。
