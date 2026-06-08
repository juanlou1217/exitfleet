# 测试计划：worker-proxy-health

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 测试范围

- SOCKS5 CONNECT 握手、目标解析和 outbound IP 传递。
- SOCKS5 不支持命令的错误响应。
- outbound IP 为空时默认 dialer 拒绝直连。
- HealthChecker 多 URL、失败阈值和 Worker 状态联动。
- Worker env 默认值和 env 覆盖。
- Docker launcher 新增 env 注入。

## 不测试范围

- 真实 OpenVPN 连通。
- 真实 VPNGate 节点切换。
- 真实 Docker daemon 构建和容器运行。
- 3x-ui API 注册。
- SOCKS5 UDP、IPv6、认证。

## 测试用例

| 用例 | 输入/动作 | 期望结果 | 类型 |
| --- | --- | --- | --- |
| SOCKS5 CONNECT | no-auth 握手后请求 `example.test:443` | dialer 收到 `tcp`、目标地址和 `10.8.0.2` outbound IP | 单元 |
| SOCKS5 unsupported command | 请求命令 `0x02` | 返回 SOCKS5 command-not-supported | 单元 |
| outbound IP 防直连 | 默认 dialer 且 outbound IP 为空 | CONNECT 返回失败，dialer 返回 `ErrOutboundIPRequired` | 单元 |
| 多 URL 健康检查 | 第一个 URL 失败，第二个成功 | 健康，连续失败次数清零 | 单元 |
| 失败阈值 | 单 URL 连续失败两次，阈值 2 | `Ready()` 返回 false | 单元 |
| Worker 状态联动 | ready worker 健康检查失败且阈值 1 | Worker 状态变为 `failed`，`health_failures=1` | 单元 |
| env 配置 | 设置 health/proxy env | Worker config 读取新值并校验 | 单元 |
| Docker env 注入 | 启动 fake Docker launcher | `docker run` 参数包含 health/proxy env | 单元 |

## 测试命令

```bash
go fmt ./...
go test ./...
go vet ./...
docker build -f Dockerfile.worker -t exitfleet-worker:local .
```

## 执行结果

- 未执行原因：Docker CLI 已安装，但 Docker daemon 当前未运行；`docker version` 无法连接 `/var/run/docker.sock`，`colima status` 返回 not running，本轮未执行真实 `docker build`。
- 实际结果：
  - `go fmt ./...`：通过。
  - `go test ./...`：通过。
  - `go vet ./...`：通过。

## 未覆盖风险

- 真实 OpenVPN ready 后的 TUN IP 获取和 `EXITFLEET_WORKER_PROXY_OUTBOUND_IP` 动态设置尚未完成。
- Manager 尚未基于 worker health 执行自动切换。
- 真实 SOCKS5 代理经 VPN 出口访问公网未验证。
