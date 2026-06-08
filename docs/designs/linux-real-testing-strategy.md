# 技术方案：Linux 真实测试策略

## 状态

- 状态：proposed
- 创建日期：2026-06-08
- 负责人：AI/User

## 背景

ExitFleet 的目标运行环境是 Debian、Ubuntu、CentOS-compatible、Alpine 等 Linux 系统。当前开发机是 macOS。macOS 可以完成大部分 Go 单元测试和纯逻辑验证，但不能真实代表 Linux 上的 TUN、OpenVPN、iptables/rp_filter、systemd、Docker worker 网络命名空间行为。

因此项目需要一套分层测试策略：在 macOS 上保持快速反馈，在 Linux 容器中覆盖发行版差异，在真实 Linux VM/VPS 上做至少一次端到端校准，确保 Go 单元测试和 fake/integration test 的方向没有偏。

## 方案

采用四层测试：

```text
L0: Go 单元测试
  运行位置：macOS、Linux CI
  目标：快速验证纯逻辑、解析、状态机、命令构建、错误分类

L1: Linux 容器兼容测试
  运行位置：Docker Desktop / Linux CI
  目标：验证 Debian、Ubuntu、CentOS-compatible、Alpine 的包、路径、命令、脚本差异

L2: Linux privileged 集成测试
  运行位置：Linux VM 或 Linux CI runner
  目标：验证 /dev/net/tun、CAP_NET_ADMIN、OpenVPN fake process、proxy bind、worker 容器模型

L3: 真实 Linux 端到端测试
  运行位置：真实 VPS 或本地 Linux VM
  目标：验证真实 OpenVPN、真实 TUN、真实代理出口、真实健康检查和自动切换链路
```

## 模块边界

### macOS 本地

macOS 只作为开发环境和快速单元测试环境，不作为最终系统行为依据。

适合测试：

- VPNGate CSV 解析。
- OpenVPN 配置解码。
- 节点评分和筛选。
- 调度器状态机。
- Docker 命令参数构建。
- OpenVPN 命令参数构建。
- API handler 的输入输出。

不适合测试：

- 真实 TUN 设备。
- Linux 策略路由。
- `SO_BINDTODEVICE`。
- `rp_filter`。
- systemd 行为。
- Linux 防火墙行为。

### Linux 容器矩阵

使用容器测试发行版兼容性：

```text
debian:bookworm
ubuntu:24.04
rockylinux:9 或 almalinux:9
alpine:3
```

容器矩阵主要验证：

- 包管理器差异。
- OpenVPN 安装路径。
- shell 脚本兼容性。
- 基础命令存在性。
- 配置文件路径。
- manager/worker 二进制能启动。

容器矩阵不作为最终网络行为依据。Docker Desktop 在 macOS 上本质是 Linux VM，能提供部分 Linux 行为，但仍不能完全代表 VPS 网络、防火墙和宿主机 TUN 配置。

### Linux privileged 集成测试

在 Linux VM 或 Linux runner 中运行 privileged worker：

```bash
docker run --rm \
  --cap-add NET_ADMIN \
  --device /dev/net/tun \
  exitfleet-worker:test
```

这一层使用 fake OpenVPN 或可控测试进程，不依赖真实 VPNGate 节点。目标是稳定验证 worker 生命周期：

- worker 能看到 `/dev/net/tun`。
- worker 能启动 fake OpenVPN。
- worker 能启动 HTTP/SOCKS5 代理。
- worker `/healthz` 和 `/readyz` 状态正确。
- worker 退出时能清理进程。

### 真实 Linux 端到端测试

每个大版本或涉及网络运行时的重大变更，必须至少跑一次真实 Linux 端到端测试。

推荐环境：

```text
首选：一台干净 Ubuntu LTS VPS
补充：Debian VPS
补充：Alpine VM 或容器
CentOS-compatible：Rocky/Alma/CentOS Stream 环境
```

### 北京服务器的定位

北京服务器可以作为真实 Linux 测试环境，但它不一定适合验证“VPNGate 成功连接”这条 happy path。原因是中国大陆机房到 `www.vpngate.net`、OpenVPN 节点、出口 IP 查询服务的访问可能受到 DNS、TLS、TCP 连接、UDP/TCP OpenVPN 协议层面的限制。

因此北京服务器建议承担两类测试：

```text
正向测试：
  Linux 基础环境
  安装脚本
  Docker/worker 权限
  /dev/net/tun
  OpenVPN 命令存在性
  proxy 进程启动
  health/readiness 状态

诊断测试：
  VPNGate API DNS 失败
  API 连接失败
  OpenVPN 节点连接超时
  TLS handshake 失败
  出口代理不可用
  防火墙、安全组、rp_filter、TUN 缺失
```

它不应该是唯一的真实成功路径环境。为了验证真实 VPN 出口成功，至少还需要一个网络更中性的 Linux 环境，例如香港、日本、新加坡、美国或欧洲 VPS。这样可以把“代码问题”和“北京网络限制”区分开。

真实测试覆盖：

- 安装 OpenVPN。
- 开启 TUN。
- 启动 manager。
- 拉取 VPNGate 节点。
- 选择一个节点启动 worker。
- worker 建立真实 OpenVPN 连接。
- 通过 HTTP/SOCKS5 代理访问出口 IP 查询服务。
- 验证出口 IP 与 worker 节点一致。
- 验证断线后健康检查进入失败状态。
- 验证日志可查询。

## 测试环境准备

### 北京服务器基础要求

需要准备：

```text
系统：Debian/Ubuntu/CentOS-compatible/Alpine 之一
权限：root 或具备 sudo 权限
内核能力：支持 /dev/net/tun
容器能力：支持 Docker 或兼容容器运行时
网络：允许出站 TCP 443、TCP/UDP OpenVPN 常见端口
安全组：放行 manager 管理端口和 worker 代理端口
磁盘：至少 2GB 可用空间
内存：建议 1GB 以上
```

需要先采集环境快照：

```bash
uname -a
cat /etc/os-release
id
ip route
ls -l /dev/net/tun
sysctl net.ipv4.ip_forward
sysctl net.ipv4.conf.all.rp_filter
docker version
openvpn --version
```

如果某些命令不存在，也要记录下来。这些缺失本身就是兼容性测试输入。

### 成功路径测试环境

如果北京服务器无法稳定连接 VPNGate，建议额外准备一台境外 Linux VPS，用于成功路径校准：

```text
Ubuntu LTS VPS
TUN 已开启
Docker 可用
OpenVPN 可用
能访问 www.vpngate.net
能连接至少一个 VPNGate OpenVPN 节点
能访问出口 IP 查询服务
```

北京服务器和境外 VPS 的作用不同：

```text
北京服务器：校准受限网络诊断和失败路径
境外 VPS：校准真实连接成功路径和代理出口
```

## 测试数据准备

真实测试前需要准备以下数据。能用假数据的部分先用 fixture，必须真实联网的部分单独标记。

### VPNGate API 样本

用途：校准 `internal/vpngate` 的 CSV 解析和节点模型。

准备：

```text
原始 API 响应样本：vpngate_iphone_sample.csv
空响应样本：vpngate_empty.csv
损坏 CSV 样本：vpngate_malformed.csv
包含 UDP/TCP 节点的样本
包含 Base64 OpenVPN 配置的样本
```

建议保存到：

```text
internal/vpngate/testdata/
```

### OpenVPN 配置样本

用途：校准 OpenVPN 配置解码、remote/proto 解析、命令构建。

准备：

```text
tcp.ovpn
udp.ovpn
missing_remote.ovpn
invalid_base64_config.txt
route_nopull_expected_args.txt
```

建议保存到：

```text
internal/openvpn/testdata/
```

### OpenVPN 日志样本

用途：校准错误分类和诊断提示。

准备：

```text
openvpn_success.log
openvpn_auth_failed.log
openvpn_tun_missing.log
openvpn_dns_failed.log
openvpn_tls_timeout.log
openvpn_connection_timeout.log
openvpn_connection_refused.log
```

这些日志可以从北京服务器和境外 VPS 真实测试中采集，然后脱敏后固化为单元测试 fixture。

### Proxy 协议样本

用途：校准 HTTP CONNECT、普通 HTTP proxy、SOCKS5 握手解析。

准备：

```text
http_connect_request.txt
http_absolute_request.txt
socks5_connect_ipv4.bin
socks5_connect_domain.bin
socks5_connect_ipv6.bin
```

建议保存到：

```text
internal/proxy/testdata/
```

### 调度和状态样本

用途：校准自动切换、固定 IP、固定国家、IP 类型偏好。

准备：

```text
nodes_mixed_countries.json
nodes_all_failed.json
nodes_fixed_region_jp.json
nodes_ip_types.json
workers_running.json
workers_unhealthy.json
```

建议保存到：

```text
internal/node/testdata/
internal/scheduler/testdata/
internal/worker/testdata/
```

### 真实端到端记录

用途：校准真实测试是否成功，后续回归时可对照。

每次真实测试必须记录：

```text
测试时间
服务器地区和系统版本
公网 IP
TUN 状态
Docker/OpenVPN 版本
VPNGate API 拉取结果
选中的节点国家/IP/proto/port
OpenVPN 连接日志尾部
代理监听端口
通过代理查询到的出口 IP
健康检查结果
失败诊断结果
```

这些记录写入：

```text
docs/tests/real-linux-e2e.md
```

## Go 单元测试如何用真实测试校准

真实测试不能替代单元测试。真实测试负责生成事实样本，反过来校准单元测试。

需要沉淀为 fixture：

```text
internal/vpngate/testdata/vpngate_iphone_sample.csv
internal/openvpn/testdata/openvpn_success.log
internal/openvpn/testdata/openvpn_auth_failed.log
internal/openvpn/testdata/openvpn_tun_missing.log
internal/proxy/testdata/http_connect_request.txt
internal/proxy/testdata/socks5_handshake.bin
```

真实测试后要把这些样本转成 Go 单元测试：

- VPNGate CSV 解析测试。
- OpenVPN 日志错误分类测试。
- OpenVPN 命令构建测试。
- proxy 握手和 CONNECT 解析测试。
- scheduler 状态迁移测试。

## 测试执行策略

### 每次提交前

```bash
go test ./...
go fmt ./...
```

### 每个功能分支

```bash
go test ./...
go test -race ./...
```

如果涉及 Docker worker：

```bash
docker build -f Dockerfile.manager .
docker build -f Dockerfile.worker .
```

### 每个 release 前

必须跑：

- Go 单元测试。
- Linux 容器矩阵。
- Linux privileged 集成测试。
- 至少一次真实 Linux 端到端测试。

## 测试开关

真实测试不能默认跑在普通单元测试中，避免网络不稳定和外部服务波动。

建议使用：

```bash
go test ./...
go test -tags=integration ./...
EXITFLEET_REAL_E2E=1 go test -tags=real_e2e ./...
```

约定：

- `go test ./...` 只跑稳定单元测试。
- `integration` 允许使用 Docker/fake OpenVPN。
- `real_e2e` 才允许访问真实 VPNGate、真实 OpenVPN、真实公网出口。

## 取舍

- macOS 不能作为 Linux 网络行为的真实依据，但适合快速开发。
- Docker Desktop 可以做容器兼容测试，但不能完全替代真实 VPS。
- 真实 VPNGate 测试不适合每次提交都跑，应作为手动或 release gate。
- 单元测试必须稳定，不能依赖实时 VPNGate 节点。

## 风险

- VPNGate 节点质量波动会导致真实测试不稳定。
- VPS 防火墙、安全组、TUN 开关会影响结果。
- Alpine、CentOS-compatible 的包和服务管理差异会增加维护成本。
- Docker Desktop 和真实 Linux VPS 的网络行为不同。

## 回滚策略

如果真实测试不稳定：

- 保留真实测试为手动 release gate。
- 把真实日志和响应固化为 fixture。
- 优先保证 Go 单元测试和 fake integration 稳定。

## 实施步骤

1. 建立 `testdata` fixture 目录。
2. 为 VPNGate 解析、OpenVPN 日志分类、proxy 协议解析写 Go 单元测试。
3. 增加 fake OpenVPN 测试进程。
4. 增加 Linux 容器矩阵脚本或 CI。
5. 增加 privileged worker 集成测试。
6. 准备一台 Linux VPS 做真实端到端校准。
7. 将真实测试样本沉淀回 Go 单元测试。

## 验证方式

- 静态检查本文档存在。
- 检查测试记录存在。
- 后续实现时，真实 E2E 测试结果必须记录到 `docs/tests/`。
