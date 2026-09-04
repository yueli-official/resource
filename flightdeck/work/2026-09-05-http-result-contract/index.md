# 统一错误与 HTTP Result 合同

## Goal

让 Resource API、Go Adapter 与 Nuxt 消费者采用 Foundation Project v1、声明式错误目录、标准成功 DTO 和安全失败参数，并通过真实 Resource 浏览与管理流程验收。

## Status

Finished

## Current

Resource 已完成统一 HTTP Result 迁移：49 个 operation 由 canonical OpenAPI 和 Project v1 管理，错误目录、operation-errors、Go/TS/i18n、legacy catalog 与 CI freshness/compatibility diff 均已落地。创建/删除成功语义统一为 201/204，公开 Problem 不再携带 raw detail；前端按生成 code 与 Foundation resolver 提供原位安全反馈。Foundation Go v0.4.1、JS js-v0.7.2 与 gRPC v1.82.1 已锁定。Go test/race/vet/build、govulncheck、Web unit/typecheck/build 全绿；Workspace Shared 真实组合 ready，Playwright journeys 7/7、列表失败恢复与管理写入 Problem 2/2 通过。

## Next

None.

## Progress

- 建立 Foundation Project v1 全量合同和生成门禁，修正 201/204 成功响应。
- 删除公开 raw message/detail 路径，统一 Resource 前端 failure resolver 与反馈载体。
- 修复可达 gRPC 漏洞并完成 Workspace Shared + CLI Playwright 真实验收。

## References

- [稳定上下文](context.md)
- [Foundation HTTP Result Contract](../../../../foundation/flightdeck/knowledge/errors/http-result-contract.md)
- [Foundation Error Catalog](../../../../foundation/flightdeck/knowledge/errors/error-catalog.md)
