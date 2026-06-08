# 测试计划：发布产物与 README 使用方式

## 状态

- 状态：accepted
- 创建日期：2026-06-08
- 负责人：AI/User

## 测试范围

- README 删除本机缺少 Go 的错误提示。
- README 删除 GitHub 本地操作和强推命令。
- README 增加预编译二进制和 Docker 镜像使用方向。

## 不测试范围

- 实际 Go 编译。
- 实际 Docker 构建。
- 实际 GitHub Release 发布。

## 测试用例

| 用例 | 输入/动作 | 期望结果 | 类型 |
| --- | --- | --- | --- |
| 检查 Go 缺失错误提示 | `rg "command not found: go|先安装 Go" README.md` | 无匹配 | 静态检查 |
| 检查 GitHub 操作命令 | `rg "gh auth status|git push --force|GitHub 操作" README.md` | 无匹配 | 静态检查 |
| 检查发布产物说明 | `rg "预编译二进制|Docker 镜像|docker compose" README.md` | 能匹配到使用方式 | 静态检查 |

## 测试命令

```bash
rg "command not found: go|先安装 Go" README.md
rg "gh auth status|git push --force|GitHub 操作" README.md
rg "预编译二进制|Docker 镜像|docker compose" README.md
```

## 执行结果

- 未执行原因：无。
- 实际结果：
  - `rg "command not found: go|先安装 Go" README.md` 已执行，无匹配。
  - `rg "gh auth status|git push --force|GitHub 操作" README.md` 已执行，无匹配。
  - `rg "预编译二进制|Docker 镜像|docker compose" README.md` 已执行，能匹配到使用方式。

## 未覆盖风险

- 本次只更新文档，没有实现 Release workflow、Dockerfile 或 Compose。
