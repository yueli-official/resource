# Resource 独立消费者迁移

## 目标

让 Resource 成为可独立发布、可由 Workspace 一键开发、可在生产环境独立或复用基础服务部署的原子消费者。

## 已完成

- 从 Platform 的 `products/resource/` 过滤历史创建独立仓库，未采用无历史复制。
- Go module 改为 `github.com/yueli-official/resource/api`，跨产品能力只消费 Foundation 正式版本。
- 产品自有 runtime、错误目录、健康检查、部署绑定检查、OpenAPI 导出和初始化命令。
- Web 移除 `@platform/*`、`workspace:` 与 `catalog:` 依赖，产品组件与站点运行时归本仓所有。
- Dockerfile、依赖能力声明、部署锁与 Complete/Hybrid/Attach Compose 拓扑归本仓所有。
- CI 在迁移期仅保留手动触发，等待所有消费者完成后统一验收。

## 当前断点

- 整理 Go/PNPM 锁文件并生成确定性契约。
- 静态检查仓库边界与部署配置一致性。
- 创建并发布 `yueli-official/resource` 的 `v0.1.0`。
- 将精确 revision 与三种本地模式接入 Workspace。
- 在 Platform 归档并删除 Resource 活跃实现，只保留迁移记录。

## 约束

- 不兼容旧 Platform 路径，不保留双真源。
- 本批次不启动服务、不跑逐产品测试；所有消费者迁移完成后统一验收。
- `doctor.yaml` 不作为任何运行或部署契约。
