# 技术方案：项目初始化与文档留痕机制

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 背景

ExitFleet 当前处于 Go 重构初始化阶段。项目需要在早期建立清晰的目录结构、架构边界、接口草案和 AI 修改约束，避免后续功能开发只存在于聊天记录中。

## 方案

初始化阶段建立以下机制：

- 使用 `AGENTS.md` 作为项目 harness，约束 AI 修改规则。
- 使用 `docs/features/` 记录功能开发。
- 使用 `docs/designs/` 记录技术方案。
- 使用 `docs/tests/` 记录测试计划与验证结果。
- 使用 `docs/decisions/` 记录架构决策。
- 使用 `docs/templates/` 提供统一文档模板。

## 模块边界

- `cmd/` 只放进程入口。
- `internal/` 放 Go 业务模块。
- `docs/` 放架构、接口、方案、测试和决策记录。
- 原 Python 项目不复制进本仓库。

## 数据流

```text
User request
  -> AI drafts feature/design/test docs
  -> AI implements or updates code
  -> AI records verification result
  -> Commit includes code and matching docs
```

## 接口变化

本次初始化不新增运行时接口，只维护文档和项目约束。

## 取舍

- 选择仓库内 Markdown 记录，而不是外部文档系统。
- 选择按功能/方案/测试/决策拆目录，而不是单一大文档。

## 风险

- 文档可能滞后于实现。
- 小改动如果强制记录过细，会增加维护成本。

## 回滚策略

如果记录机制过重，可以保留目录结构，但只要求较大功能和架构变化必须记录。

## 实施步骤

1. 更新 `AGENTS.md` 的 AI 修改规则。
2. 创建 `docs/features/`、`docs/designs/`、`docs/tests/`、`docs/decisions/`。
3. 创建文档模板。
4. 增加 ADR 记录本决策。

## 验证方式

- 检查目录和模板是否存在。
- 检查 `AGENTS.md` 是否声明 AI 产物必须落盘。
- 检查 Git 工作区和提交记录。

