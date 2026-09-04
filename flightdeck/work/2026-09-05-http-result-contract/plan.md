# Plan

## P0 — 合同

- [x] 建立 canonical OpenAPI、v1 error catalog、operation-errors 与 Project v1。
- [x] 修正 collection/page、201/202/204 与特殊响应形状。

## P1 — 实现与消费者

- [x] 生成 Go/TypeScript/i18n，移除公开 `detail`，并按用户恢复语义完成 typed cause 与前端 feedback。
- [x] 更新 Foundation 依赖、Nuxt 消费者和 CI freshness/compatibility 门禁。

## P2 — 验证

- [x] Go test/race/vet/govuln 与 Web tests/typecheck/build 全绿。
- [x] Workspace Shared local checkout 与 CLI Playwright 完成公开浏览、管理、列表恢复和写入失败反馈验收。
