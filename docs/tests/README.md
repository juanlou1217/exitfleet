# Go 单元测试用例放在哪里

ExitFleet 的测试用例指 **Go 单元测试**。测试记录是辅助文档，不能替代 Go 单元测试。

## Go 单元测试

Go 单元测试必须写在被测 Go package 同目录，文件名必须使用 `*_test.go`。

示例：

```text
internal/api/contract.go
internal/api/contract_test.go

internal/config/config.go
internal/config/config_test.go

internal/openvpn/command.go
internal/openvpn/command_test.go
```

要求：

- 新功能必须先写对应 Go 单元测试，再写实现。
- 行为变更必须补充或更新对应 Go 单元测试。
- 成功路径和关键失败路径都要覆盖。
- 测试命名要描述行为，不要只写 `TestWorks` 这类模糊名称。

## 测试记录

测试计划、Go 单元测试用例清单、测试命令、执行结果和未覆盖风险必须写到：

```text
docs/tests/short-topic.md
```

注意：`docs/tests/` 只是记录，不是测试用例本体。真正的测试用例必须写在 `*_test.go` 中。

模板：

```text
docs/templates/test-plan.md
```

## 提交门禁

提交前必须执行：

```bash
go test ./...
go fmt ./...
```

涉及 Docker worker 时追加：

```bash
docker build -f Dockerfile.manager .
docker build -f Dockerfile.worker .
```

涉及 OpenVPN 或代理行为时，需要补充 fake OpenVPN、可控测试进程或集成测试。

## 不能通过测试时怎么办

- 不能声称测试通过。
- 必须在 `docs/tests/` 记录未执行或失败原因。
- 功能代码默认不能提交。
- 只有用户明确批准临时例外时，才能提交未完全通过测试的变更。
