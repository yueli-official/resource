# 统一错误与 HTTP Result 合同

## Goal

让 Resource API、Go Adapter 与 Nuxt 消费者采用 Foundation Project v1、声明式错误目录、标准成功 DTO 和安全失败参数，并通过真实 Resource 浏览与管理流程验收。

## Status

Finished

## Current

统一 HTTP Result 迁移及 review 修复已完成：`7ffe7ff` 建立 49-operation Project v1 合同，`8d75564` 修复 Location、结构化 feedback 与原位反馈，`6ee85f6` 建立 HTTP-independent `rescause` → application mapping → Problem projection seam，`9b63278` 清理误导别名与重复构造。最终双轴 review Standards 0 / Spec 0；Go、Web、Workspace Shared 与 CLI Playwright 门禁均通过。

## Next

None.

## Progress

- 建立 Foundation Project v1 全量合同和生成门禁，修正 201/204 与 Location 成功语义。
- 分离 domain/provider typed cause、application catalog mapping 与 HTTP projection。
- 保留 violations/summary/trace 技术详情并统一原位反馈；最终双轴 review 无发现。


## References

- [稳定上下文](context.md)
- [Foundation HTTP Result Contract](../../../../foundation/flightdeck/knowledge/errors/http-result-contract.md)
- [Foundation Error Catalog](../../../../foundation/flightdeck/knowledge/errors/error-catalog.md)
