# 本地验证与接入

## 范围
资源封面、正文、SEO 图片、附件；删除和替换只撤销关系，不删除素材文件。

## 已通过
- 当前产品业务回归和 API main 编译；使用明确本地 Asset checkout 及已记录 Foundation 源码快照的临时 modfile。
- `internal/assetreferences` 真实 PostgreSQL 临时表测试：重复/共享素材、替换、删除或撤销、恢复；Shop 另验证组合交付、图集与合集，Gallery 另验证撤稿与公开授权失败重试，Commerce 另验证多文件及跨 site 隔离。
- SDK 真实 PostgreSQL + HTTP 夹具验证：读取失败不清空、503 后恢复、不让某类未解析素材阻断其他类型对账；MediaKey 节点解析及禁止重定向泄露机器凭据。
- 消费者 JSON Schema 校验通过。无新数据库迁移。

## 运行约束
Asset Go v0.3.0 尚无 referencesync；本轮是本地源码覆盖验证，未发布 SDK。SDK 依赖的新间接版本和校验和已同步到产品 go.mod/go.sum，无硬编码本地 replace。
共用环境变量：ASSET_BASE_URL、ASSET_REFERENCE_TOKEN_URL、ASSET_REFERENCE_CLIENT_ID、ASSET_REFERENCE_CLIENT_SECRET、ASSET_PUBLIC_ORIGIN。机器身份需要 asset:sign 及此消费者的 Registration binding；缺失配置拒绝启动，不能默默跳过登记。
Commerce 另需 ASSET_REFERENCE_SITE_KEY 指定订单业务 site（默认 shop）；本仓库演示目录显式 COMMERCE_REFERENCE_DEMO_CATALOG=.data/consumer-catalog.json，不能将其他消费者目录交给 Commerce 扫描。Gallery/Licensing 复用原 Asset 机器配置。

## 实际运行验收
2026-09-11 真实 API 生命周期 3→2→1→0、附件 1→0、删除保护 409 通过；附件移除不删 Blob。CLI Playwright 验证桌面和移动端引用弹窗、缩略图解码和布局。仅清理本次测试数据。
证据 E:/tmp/yueli-consumer-references-20260911/resource-lifecycle-report.json、resource-references-{desktop,mobile}.png。无正式 SDK 发布、生产部署或 Git 提交。
