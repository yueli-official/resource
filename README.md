# 资源产品

- 生命周期：活跃的可复用产品类型
- 权威来源：Catalog 产品类型 `resource`、`api/` 迁移/OpenAPI、`web/` 界面
- 消费者：`resource-main` 等资源站点实例
- 验证：`pnpm platformctl verify product --file catalog/overlays/local.yaml --root . resource`

Resource 负责可下载资源的元数据、发布、分类、交付引用和设置。文件仍是 Asset 对象，权益与支付仍属于 Commerce。资源访问量由实例本地 Foundation Traffic 持有，`resources.view_count` 只是只读查询投影，管理端不可编辑。

分类数据保留在每个 Resource 实例自己的 content PostgreSQL：revisioned Catalog、Policy、Category、扁平 Tag、Tag Lookup 与 Assignment 都由产品事务持有，规则直接消费 `github.com/yueli-official/foundation/go/classification`。`resources.tags` 是资源作者维护的搜索关键词，不是 Classification Tag 或治理身份。`api/` 拥有领域状态，`web/` 拥有公开和管理体验。

权限同样由每个 Resource 实例自己的 PostgreSQL 持有。Identity 只提供登录 Subject；Authorization 使用
`Site → Resource` Scope、`owner` relation、受保护的“管理员”和普通“贡献者”角色。贡献者可以独立创建、发布和维护自己的资源，
管理员可以治理全站内容、分类、设置、角色能力和申请。注册自动成为贡献者是可发布的策略规则，默认开启以兼容“登录即可贡献”，
关闭后仍可通过申请、邀请或管理员直接授权接入。
