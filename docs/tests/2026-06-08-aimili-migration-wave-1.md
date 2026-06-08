# 测试计划：AimiliVPN 首轮迁移骨架

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 测试范围

- VPNGate CSV 解析和 OpenVPN config 解码。
- 节点测速 fake runner 成功和失败路径。
- Worker OpenVPN 命令构建和状态机。
- Manager 节点刷新、节点测试、出口启动/停止约束。
- `/api/v1` manager HTTP API 和 worker `/healthz`、`/readyz`。

## 不测试范围

- 真实 VPNGate 网络请求。
- 真实 OpenVPN 连接。
- 真实 Docker worker 生命周期。
- 真实 HTTP/SOCKS5 代理转发。
- 生产 Web UI。

## 测试用例

| 用例 | 输入/动作 | 期望结果 | 类型 |
| --- | --- | --- | --- |
| 解析正常 VPNGate CSV | 含 `#HostName,IP,Score,...,OpenVPN_ConfigData_Base64` 的静态文本 | 返回节点并解码 config | 单元测试 |
| 解析缺列 CSV | OpenVPN config 列缺失 | 返回错误 | 单元测试 |
| 解码非法 base64 | `DecodeOpenVPNConfig("bad")` | 返回错误 | 单元测试 |
| 节点测速成功 | fake runner 返回 latency | probe status 为 available | 单元测试 |
| 节点测速失败 | fake runner 返回错误 | probe status 为 unavailable，last error 有值 | 单元测试 |
| OpenVPN 命令构建 | config path、auth path、tun0、route-nopull | 参数包含预期 flags | 单元测试 |
| Worker 状态机 | fake runner ready/fail/stop | 状态正确迁移 | 单元测试 |
| Manager 调度约束 | refresh 后不 start exit | worker 未创建 | 单元测试 |
| Manager 启停出口 | StartExit/StopExit | active exit 创建并释放 | 单元测试 |
| Manager 禁止重复出口 | 同一 node ID 连续 StartExit | 第二次启动返回错误且不创建新 worker | 单元测试 |
| Manager 端口复用安全 | 启动两个出口，停止第一个，再启动第三个 | 分配第一个空闲端口，不占用仍运行出口端口 | 单元测试 |
| API 节点脱敏 | GET/refresh/test nodes | 响应不包含 `config_text` 或 OpenVPN 配置正文 | handler 测试 |
| HTTP API fake 流程 | refresh/start/stop/logs stream | 返回预期状态码和 JSON/SSE shape | handler 测试 |
| Worker health API | ready/stopped worker | `/healthz` 200，`/readyz` 按状态返回 | handler 测试 |

## 测试命令

```bash
go test ./...
go fmt ./...
```

## 执行结果

- 未执行原因：Docker/OpenVPN/真实代理链路不属于首轮 fake 骨架验证范围。
- 实际结果：
  - 初次检查发现当前机器缺少 `go` 和 `gofmt`。
  - 已通过 Homebrew 安装 Go：`go version go1.26.4 darwin/arm64`。
  - `gofmt -w cmd internal` 已执行成功。
  - `go test ./...` 已执行成功。
  - `go fmt ./...` 已执行成功。

## 未覆盖风险

- 首轮 fake 测试不能证明真实 Docker/OpenVPN/代理链路可用。
- 生产暴露前必须补充认证或反代保护测试。
