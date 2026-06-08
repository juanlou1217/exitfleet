# 测试计划：项目初始化与文档留痕机制

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 测试范围

- 文档目录结构存在。
- 文档模板存在。
- `AGENTS.md` 包含 AI 开发记录要求。
- Git 工作区可提交。

## 不测试范围

- Go 编译和单元测试。
- Docker 构建。
- OpenVPN、代理、VPNGate 网络连通性。

## 测试用例

| 用例 | 输入/动作 | 期望结果 | 类型 |
| --- | --- | --- | --- |
| 检查文档目录 | `find docs -maxdepth 2 -type f` | 能看到 templates、designs、tests、decisions 文件 | 静态检查 |
| 检查 AGENTS 规则 | `rg "AI 开发记录要求|docs/features|docs/designs|docs/tests|docs/decisions" AGENTS.md` | 能匹配到记录要求 | 静态检查 |
| 检查本机路径 | `rg "zhao[k]ang|/User[s]/" README.md AGENTS.md docs` | 无匹配 | 静态检查 |

## 测试命令

```bash
find docs -maxdepth 2 -type f
rg "AI 开发记录要求|docs/features|docs/designs|docs/tests|docs/decisions" AGENTS.md
rg "zhao[k]ang|/User[s]/" README.md AGENTS.md docs
go test ./...
```

## 执行结果

- 未执行原因：当前机器缺少 `go` 命令，Go 测试需要安装 Go 工具链后执行。
- 实际结果：
  - `find docs -maxdepth 2 -type f` 已执行，目录和模板文件存在。
  - `rg "AI 开发记录要求|docs/features|docs/designs|docs/tests|docs/decisions" AGENTS.md` 已执行，能匹配到记录要求。
  - `rg "zhao[k]ang|/User[s]/" README.md AGENTS.md docs` 已执行，无匹配。
  - `go test ./...` 已执行失败，错误为 `zsh:1: command not found: go`。

## 未覆盖风险

- 文档规则是否在后续每次开发中被严格执行，需要靠 review 和 AGENTS 约束持续维护。
