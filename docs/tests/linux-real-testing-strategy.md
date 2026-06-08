# 测试计划：Linux 真实测试策略

## 状态

- 状态：proposed
- 创建日期：2026-06-08
- 负责人：AI/User

## 测试范围

- 文档是否明确 macOS、Linux 容器、Linux privileged、真实 Linux E2E 的职责。
- 文档是否说明真实测试如何反向校准 Go 单元测试。
- 文档是否给出真实 Linux 测试的 release gate。
- 文档是否说明北京服务器的定位和测试数据准备清单。

## 不测试范围

- 本次不执行真实 Linux VPS 测试。
- 本次不执行 Docker privileged 测试。
- 本次不执行 Go 单元测试。

## 测试用例

| 用例 | 输入/动作 | 期望结果 | 类型 |
| --- | --- | --- | --- |
| 检查设计文档 | `test -f docs/designs/linux-real-testing-strategy.md` | 文件存在 | 静态检查 |
| 检查分层策略 | `rg "L0: Go 单元测试|L1: Linux 容器兼容测试|L2: Linux privileged 集成测试|L3: 真实 Linux 端到端测试" docs/designs/linux-real-testing-strategy.md` | 能匹配四层测试 | 静态检查 |
| 检查真实测试校准单元测试 | `rg "真实测试后要把这些样本转成 Go 单元测试|testdata" docs/designs/linux-real-testing-strategy.md` | 能匹配校准策略 | 静态检查 |
| 检查 release gate | `rg "每个 release 前|至少一次真实 Linux 端到端测试" docs/designs/linux-real-testing-strategy.md` | 能匹配 release 前要求 | 静态检查 |
| 检查北京服务器定位 | `rg "北京服务器|受限网络诊断|境外 VPS" docs/designs/linux-real-testing-strategy.md` | 能匹配北京服务器定位 | 静态检查 |
| 检查测试数据准备 | `rg "测试数据准备|VPNGate API 样本|OpenVPN 日志样本|Proxy 协议样本" docs/designs/linux-real-testing-strategy.md` | 能匹配测试数据清单 | 静态检查 |

## 测试命令

```bash
test -f docs/designs/linux-real-testing-strategy.md
rg "L0: Go 单元测试|L1: Linux 容器兼容测试|L2: Linux privileged 集成测试|L3: 真实 Linux 端到端测试" docs/designs/linux-real-testing-strategy.md
rg "真实测试后要把这些样本转成 Go 单元测试|testdata" docs/designs/linux-real-testing-strategy.md
rg "每个 release 前|至少一次真实 Linux 端到端测试" docs/designs/linux-real-testing-strategy.md
rg "北京服务器|受限网络诊断|境外 VPS" docs/designs/linux-real-testing-strategy.md
rg "测试数据准备|VPNGate API 样本|OpenVPN 日志样本|Proxy 协议样本" docs/designs/linux-real-testing-strategy.md
```

## 执行结果

- 未执行原因：无。
- 实际结果：
  - `test -f docs/designs/linux-real-testing-strategy.md` 已执行，文件存在。
  - `rg "L0: Go 单元测试|L1: Linux 容器兼容测试|L2: Linux privileged 集成测试|L3: 真实 Linux 端到端测试" docs/designs/linux-real-testing-strategy.md` 已执行，能匹配四层测试。
  - `rg "真实测试后要把这些样本转成 Go 单元测试|testdata" docs/designs/linux-real-testing-strategy.md` 已执行，能匹配校准策略。
  - `rg "每个 release 前|至少一次真实 Linux 端到端测试" docs/designs/linux-real-testing-strategy.md` 已执行，能匹配 release 前要求。
  - `rg "北京服务器|受限网络诊断|境外 VPS" docs/designs/linux-real-testing-strategy.md` 已执行，能匹配北京服务器定位。
  - `rg "测试数据准备|VPNGate API 样本|OpenVPN 日志样本|Proxy 协议样本" docs/designs/linux-real-testing-strategy.md` 已执行，能匹配测试数据清单。

## 未覆盖风险

- 当前只是方案测试，没有跑真实 Linux。
- 真实 Linux E2E 需要后续准备 VM/VPS、TUN、OpenVPN、worker 镜像和可脱敏的真实测试样本。
