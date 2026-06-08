# 测试计划：UI 技术栈文档调整

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 测试范围

- README 不再包含“暂不默认引入”框架清单。
- AGENTS 不再把 React/Next.js 写成默认禁止项。
- 新增 ADR 记录 UI 技术栈可演进决策。

## 不测试范围

- Go 编译。
- 前端构建。
- 浏览器交互。

## 测试用例

| 用例 | 输入/动作 | 期望结果 | 类型 |
| --- | --- | --- | --- |
| 检查 README 框架清单 | `rg "暂不默认引入|Gin|Echo|React|Next.js|Rust 核心实现" README.md` | 无匹配 | 静态检查 |
| 检查 AGENTS 框架规则 | `rg "必须先在 .*docs/designs" AGENTS.md` | 能匹配到引入框架前的设计记录要求 | 静态检查 |
| 检查 ADR | `test -f docs/decisions/2026-06-08-ui-stack-flexibility.md` | 文件存在 | 静态检查 |

## 测试命令

```bash
rg "暂不默认引入|Gin|Echo|React|Next.js|Rust 核心实现" README.md
rg "必须先在 .*docs/designs" AGENTS.md
test -f docs/decisions/2026-06-08-ui-stack-flexibility.md
```

## 执行结果

- 未执行原因：无。
- 实际结果：
  - `rg "暂不默认引入|Gin|Echo|React|Next.js|Rust 核心实现" README.md` 已执行，无匹配。
  - `rg "必须先在 .*docs/designs" AGENTS.md` 已执行，能匹配到框架引入前的设计记录要求。
  - `test -f docs/decisions/2026-06-08-ui-stack-flexibility.md` 已执行，文件存在。

## 未覆盖风险

- 本次只调整文档和规则，没有验证任何实际 React 集成方式。
