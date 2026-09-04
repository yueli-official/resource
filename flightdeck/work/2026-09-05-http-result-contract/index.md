# 统一错误与 HTTP Result 合同

## Goal

让 Resource API、Go Adapter 与 Nuxt 消费者采用 Foundation Project v1、声明式错误目录、标准成功 DTO 和安全失败参数，并通过真实 Resource 浏览与管理流程验收。

## Status

Open

## Current

Resource checkout 干净；目前只建立并提交了本 Work，产品实现尚未修改。仓库已有旧错误 catalog、OpenAPI 快照和部分标准 `items/page/size/total` DTO，但尚未接入 Foundation Project v1。初步发现 Authorization Adapter 和公开 controller 仍把 `err.Error()` 交给通用错误构造，需要区分 typed cause、内部日志与公开参数。此前为完成 Shortlink review 暂停，现在可从本页直接恢复。

## Next

依据 [执行计划](plan.md) 清点 API operation、现有错误目录和成功形状，建立 v1 catalog、operation-errors 与 Project 配置；先运行严格生成器暴露模型偏差，不发布版本。

## References

- [稳定上下文](context.md)
- [Foundation HTTP Result Contract](../../../../foundation/flightdeck/knowledge/errors/http-result-contract.md)
- [Foundation Error Catalog](../../../../foundation/flightdeck/knowledge/errors/error-catalog.md)
