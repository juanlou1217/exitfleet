# 测试计划：Go 单元测试位置和提交门禁

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 测试范围

- `AGENTS.md` 声明 Go 单元测试位置。
- `AGENTS.md` 声明测试记录位置。
- `AGENTS.md` 声明全部必需测试通过后才能提交。
- `docs/tests/README.md` 解释 Go 单元测试用例写在哪里。
- ADR 记录测试门禁决策。

## 不测试范围

- Go 单元测试执行。
- Docker 构建。
- OpenVPN 或代理集成测试。

## 测试用例

| 用例 | 输入/动作 | 期望结果 | 类型 |
| --- | --- | --- | --- |
| 检查测试位置规则 | `rg "Go 单元测试：.*\\*_test.go|测试记录文档：docs/tests" AGENTS.md` | 能匹配测试位置说明 | 静态检查 |
| 检查提交门禁 | `rg "全部必需测试必须通过后才能提交|功能代码提交默认不能绕过必需测试" AGENTS.md` | 能匹配提交门禁 | 静态检查 |
| 检查测试 README | `test -f docs/tests/README.md` | 文件存在 | 静态检查 |
| 检查 ADR | `test -f docs/decisions/2026-06-08-test-gate.md` | 文件存在 | 静态检查 |

## 测试命令

```bash
rg "Go 单元测试：.*\\*_test.go|测试记录文档：docs/tests" AGENTS.md
rg "全部必需测试必须通过后才能提交|功能代码提交默认不能绕过必需测试" AGENTS.md
test -f docs/tests/README.md
test -f docs/decisions/2026-06-08-test-gate.md
```

## 执行结果

- 未执行原因：无。
- 实际结果：
  - `rg "Go 单元测试：.*\\*_test.go|测试记录文档：docs/tests" AGENTS.md` 已执行，能匹配测试位置说明。
  - `rg "全部必需测试必须通过后才能提交|功能代码提交默认不能绕过必需测试" AGENTS.md` 已执行，能匹配提交门禁。
  - `test -f docs/tests/README.md` 已执行，文件存在。
  - `test -f docs/decisions/2026-06-08-test-gate.md` 已执行，文件存在。
  - `rg "docs/tests.*不能替代 Go 单元测试|真正的测试用例必须写在.*\\*_test.go" docs/tests/README.md AGENTS.md` 已执行，能匹配 Go 单元测试和测试记录的边界。

## 未覆盖风险

- 本次只更新项目规则和文档，没有执行 Go 测试。
