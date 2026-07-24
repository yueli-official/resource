# 资源产品

- 生命周期：活跃的可复用产品类型
- 权威来源：Catalog 产品类型 `resource`、`api/` 迁移/OpenAPI、`web/` 界面
- 消费者：`resource-main` 等资源站点实例
- 验证：`pnpm platformctl verify product --file catalog/overlays/local.yaml --root . resource`

Resource 负责可下载资源的元数据、发布、分类、交付引用和设置。文件仍是 Asset 对象，权益与支付仍属于 Commerce。资源访问量由实例本地 Foundation Traffic 持有，`resources.view_count` 只是只读查询投影，管理端不可编辑。

分类数据保留在每个 Resource 实例自己的 content PostgreSQL：revisioned Catalog、Policy、Category、扁平 Tag、Tag Lookup 与 Assignment 都由产品事务持有，规则直接消费 `github.com/yueli-official/foundation/go/classification`。`resources.tags` 是资源作者维护的搜索关键词，不是 Classification Tag 或治理身份。`api/` 拥有领域状态，`web/` 拥有公开和管理体验。
