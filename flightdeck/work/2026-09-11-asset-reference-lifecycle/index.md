# resource 资产引用生命周期

## Status
Finished

## Goal
补齐 resource 自有业务素材登记、替换/删除撤销与失败恢复，并如实提交注册声明。

## Current
2026-09-11 本地源码及真实 PostgreSQL 测试通过。Workspace 独立组合 20260911T064224Z-59732 使用当前源码启动，迁移校验通过。真实 API 上传、同图正文共享与重复、封面移除、删除资源的引用数 3→2→1→0；附件移除 1→0 后文件仍存在。引用中删除素材返回 409。
CLI Playwright 桌面/移动端验证缩略图实际解码、引用弹窗和无横向溢出/脚本错误。本次创建的资源和素材已通过 API 清理，原有未被当前正文使用的 1 份素材保留。未发布 SDK、未部署、未提交。

## Next
None。本站本地验收完成。

## References
- [上下文](context.md)
- [验证和接入](validation.md)
- [共享 SDK 与待接线补丁](../../../../asset/flightdeck/work/2026-09-11-reference-source-sdk/index.md)

## Git 交付（2026-09-11）
用户已授权本地提交。本次纳入本 Work 的实现、相关验证和部署记录；其他工作改动保留，未推送或发布正式 SDK。此前“未提交”为对应阶段的历史状态。提交及范围汇总见 Workspace `flightdeck/work/2026-09-11-target-stop-isolation/commits.md`。
