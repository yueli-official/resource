# Plan

## P0 — 合同

- [ ] 建立 canonical OpenAPI、v1 error catalog、operation-errors 与 Project v1。
- [ ] 修正 collection/page、201/202/204 与特殊响应形状。

## P1 — 实现与消费者

- [ ] 生成 Go/TypeScript/i18n，收紧 typed cause 与 Provider raw error seam。
- [ ] 更新 Foundation 依赖、Nuxt 消费者和 CI freshness/compatibility 门禁。

## P2 — 验证

- [ ] Go test/race/vet/govuln 与 Web tests/typecheck/build 全绿。
- [ ] Workspace local checkout 与 CLI Playwright 完成公开浏览、管理和失败反馈验收。
