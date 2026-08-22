# 管理后台首屏 SSR

## Goal

Resource 管理后台在有效产品会话下刷新时直接输出共享后台壳，不再等待客户端水合后才替换整页加载文案。

## Status

Complete

## Result

- `web/app/layouts/manage.vue` 直接渲染 SSR-safe `YAdminShell`，浏览器专属行为继续由共享壳的局部边界拥有。
- 补齐 Resource 的 `useApi` 共享 runtime 适配器，避免移除 ClientOnly 后暴露 SSR `useApi is not defined`。
- 恢复历史误删的 `0010_site_profile_cutover` marker、Foundation Go v0.2.1 依赖和 Web `/healthz`，标准 Workspace
  Attach 组合可以准备并进入 ready。
- Playwright 在有效产品会话下连续刷新三次；每次服务端 HTML 都包含 `data-admin-shell`、不含“正在打开控制台”，
  页面中后台壳立即可见。

## Follow-up

- Docs、Shop、Gallery、Nav、Identity Account 与 Shortlink 已在同一 Workspace 工作中完成整壳 ClientOnly 迁移，并以
  真实 OIDC 刷新过程验收。
- 旧 Identity Nuxt BFF catch-all 的 `/api/v1/*` 404 与既有 Asset upload package export 缺口属于独立共享问题。
