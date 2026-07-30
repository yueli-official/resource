# 月离资源

Resource 是独立的资源目录消费者产品，拥有资源元数据、分类、搜索、下载引用、站点设置、权限和管理界面。
`api/` 与 `web/` 是本仓唯一实现真源；仓库不依赖 Platform 源码或工作区私有包。

## 边界

- Resource 自己拥有领域数据、PostgreSQL migration、授权实例、Discovery 发布和界面组件。
- Identity 只通过 OIDC issuer、Discovery 和 JWKS 证明用户身份。
- Asset 只通过公开 HTTP 合同管理资源文件和封面；Resource 不导入 Asset 内部代码。
- Foundation 通过正式 Go module 与 JS Release 提供跨产品协议原语。
- 本地多仓编排属于相邻 `workspace`；生产部署属于本仓 Compose。

不可变依赖与能力绑定记录在：

- `deploy/contracts/requirements.json`：消费者需要的能力；
- `deploy/deployment.lock.json`：能力到具体生产者版本的部署锁。

## 本地开发

推荐从相邻 `workspace` 仓启动：

```powershell
# Identity + Account + Resource 专属 Asset + Resource
.\environments\resource-local\run.ps1 -Mode Complete

# 复用已有 Identity，管理 Resource 专属 Asset
.\environments\resource-local\run.ps1 -Mode Hybrid

# 复用已有 Identity 与 Asset，只启动 Resource
.\environments\resource-local\run.ps1 -Mode Attach

.\environments\resource-local\run.ps1 -Action Down
```

所有端口均可通过 `LOCAL_*_PORT` 覆盖；关闭 Resource target 只停止该 Workspace 会话，不会终止其他项目。

## Docker Compose

本仓提供三种生产/预发布拓扑：

```powershell
# 完整独立部署：Identity + Account + Resource 专属 Asset + Resource
Copy-Item .env.example .env
docker compose -f compose.yaml up -d --wait

# 复用已有 Identity，部署 Resource 专属 Asset
Copy-Item deploy/env/hybrid.env.example .env
docker compose -f compose.hybrid.yaml up -d --wait

# 复用已有 Identity 与 Asset
Copy-Item deploy/env/attach.env.example .env
docker compose -f compose.attach.yaml up -d --wait
```

生成 `.env` 后必须填写所有空 secret、管理员 Subject 和外部服务 URL。宿主端口由
`RESOURCE_API_PORT`、`RESOURCE_WEB_PORT`、`IDENTITY_PORT`、`IDENTITY_ACCOUNT_PORT`、`ASSET_PORT`
配置，不硬编码占用。

Resource PostgreSQL 使用锁定的 `postgres-zhparser` 镜像，因为现有搜索 migration 明确依赖 `zhparser`；
普通 PostgreSQL 不能替代该契约。

## 独立命令

```powershell
cd api
go run ./cmd/resource
go run ./cmd/errorcatalog

cd ..\web
pnpm install --frozen-lockfile --ignore-workspace
pnpm dev
```

运行配置模板位于 `api/manifest/config/config.example.yaml`。本仓不使用 `doctor.yaml`。

## 验收策略

API、Web、Compose 与浏览器合同均由本仓 CI 拥有。当前迁移批次按约定暂停逐产品测试；完成所有消费者迁移后，
再统一运行 API、前端、容器、Compose 和 Playwright 验收。
