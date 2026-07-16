# 资源产品

- 生命周期：活跃的可复用产品类型
- 权威来源：Catalog 产品类型 `resource`、`api/` 迁移/OpenAPI、`web/` 界面
- 消费者：`resource-main` 等资源站点实例
- 验证：`pnpm platformctl verify product --file catalog/overlays/local.yaml --root . resource`

Resource 负责可下载资源的元数据、发布、分类、交付引用和设置。文件仍是 Asset 对象，权益与支付仍属于 Commerce。`api/` 拥有领域状态，`web/` 拥有公开和管理体验。
