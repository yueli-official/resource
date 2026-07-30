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

## 完成状态

- Go/PNPM 锁文件、错误目录与 OpenAPI 契约已经生成并提交。
- 静态边界扫描、JSON 解析、Compose include 路径和 `git diff --check` 已通过。
- 旧浏览量导入、旧 settings 转换与 Site Profile 降级兼容已删除。
- GitHub 权威仓库是 `https://github.com/yueli-official/resource`，首个独立版本为 `v0.1.0`。
- `v0.1.1` 将默认 Asset profile 修正为真实存在的 `resource`，与专属 Asset 策略和能力声明一致。
- Workspace 接入和 Platform 旧实现退役由上层多仓迁移 Flightdeck 继续记录。

## 约束

- 不兼容旧 Platform 路径，不保留双真源。
- 本批次不启动服务、不跑逐产品测试；所有消费者迁移完成后统一验收。
- `doctor.yaml` 不作为任何运行或部署契约。
