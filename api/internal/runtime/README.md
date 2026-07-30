# Resource API 运行时装配

本模块只拥有 Resource API 的进程装配：环境配置、鉴权、HTTP 中间件、健康检查、OpenAPI 与遥测。

领域规则仍归 `internal/catalog`、`internal/resourceauthz`、`internal/resourcediscovery` 等包；跨产品协议原语直接来自
Foundation。独立 Resource 不依赖 Platform `gokit`，生产者只通过公开 HTTP/OIDC 合同绑定。
