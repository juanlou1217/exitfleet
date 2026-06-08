# 测试计划：文档文件命名规则

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 测试范围

- 已跟踪 docs 文件不再使用日期前缀。
- `AGENTS.md` 声明文档文件名不带日期。
- 旧日期前缀文档已重命名为稳定主题名。

## 不测试范围

- 未跟踪文件。
- Go 单元测试。
- Docker 构建。

## 测试用例

| 用例 | 输入/动作 | 期望结果 | 类型 |
| --- | --- | --- | --- |
| 检查已跟踪日期前缀文件 | `git ls-files docs | rg "/[0-9]{4}-[0-9]{2}-[0-9]{2}-"` | 无匹配 | 静态检查 |
| 检查 AGENTS 命名规则 | `rg "文件名.*不带日期|不使用日期前缀" AGENTS.md docs/decisions/document-naming.md` | 能匹配命名规则 | 静态检查 |
| 检查重命名目标 | `test -f docs/designs/release-distribution.md` | 文件存在 | 静态检查 |

## 测试命令

```bash
git ls-files docs | rg "/[0-9]{4}-[0-9]{2}-[0-9]{2}-"
rg "文件名.*不带日期|不使用日期前缀" AGENTS.md docs/decisions/document-naming.md
test -f docs/designs/release-distribution.md
```

## 执行结果

- 未执行原因：无。
- 实际结果：
  - `git ls-files docs | rg "/[0-9]{4}-[0-9]{2}-[0-9]{2}-"` 已执行，无匹配。
  - `rg "文件名.*不带日期|不使用日期前缀" AGENTS.md docs/decisions/document-naming.md` 已执行，能匹配命名规则。
  - `test -f docs/designs/release-distribution.md` 已执行，文件存在。

## 未覆盖风险

- 当前仍存在未跟踪的 `docs/codebase/` 文件，本次不纳入命名迁移提交。
