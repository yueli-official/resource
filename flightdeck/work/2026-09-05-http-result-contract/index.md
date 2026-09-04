# 统一错误与 HTTP Result 合同

## Goal

让 Resource API、Go Adapter 与 Nuxt 消费者采用 Foundation Project v1、声明式错误目录、标准成功 DTO 和安全失败参数，并通过真实 Resource 浏览与管理流程验收。

## Status

Open

## Current

初次实现已提交为 `7ffe7ff`，但双轴 review 发现阻断问题：domain/adapter 尚未形成真正 typed cause 映射 seam，前端 feedback 被压成字符串而丢失 violations/traceId，Asset 与编辑器失败仍有 Toast-only 反馈，资源创建 201 缺少 Location。Work 已重新打开，Nav 暂未开始。

## Next

修复 review 的 typed cause、结构化 feedback、原位反馈与 Location 四项发现，重新完成全量验证和双轴 review；通过前不得进入 Nav。


## References

- [稳定上下文](context.md)
- [Foundation HTTP Result Contract](../../../../foundation/flightdeck/knowledge/errors/http-result-contract.md)
- [Foundation Error Catalog](../../../../foundation/flightdeck/knowledge/errors/error-catalog.md)
