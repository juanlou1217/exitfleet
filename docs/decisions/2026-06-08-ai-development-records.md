# ADR：AI 开发产物必须项目内留痕

## 状态

- 状态：accepted
- 日期：2026-06-08

## 背景

ExitFleet 会长期使用 AI 辅助开发。功能方案、技术方案、测试用例和验证结论如果只留在聊天记录中，会导致后续开发缺少可追踪上下文，也会让 `AGENTS.md` 的约束难以执行。

## 决策

AI 生成的重要开发产物必须记录在项目仓库中：

- 功能行为变化记录到 `docs/features/`。
- 技术方案和模块边界记录到 `docs/designs/`。
- 测试计划、测试用例和验证结果记录到 `docs/tests/`。
- 长期架构决策记录到 `docs/decisions/`。

## 后果

- 正面影响：项目上下文可追踪，后续 AI 和开发者可以直接从仓库恢复决策背景。
- 负面影响：每次开发会多一步文档维护成本。
- 约束：实现和文档不一致时，必须先更新文档再完成任务。

## 替代方案

- 只依赖聊天记录：不可接受，聊天上下文不可长期稳定复用。
- 只依赖 commit message：信息密度不够，无法表达测试和架构取舍。

## 相关文档

- `AGENTS.md`
- `docs/templates/feature.md`
- `docs/templates/technical-plan.md`
- `docs/templates/test-plan.md`

