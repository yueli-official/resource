# resource 后台集合布局

## Goal
将本站管理集合接入用户确认的共享分页、网格、评论和标题工具规则，保留权限与查询行为。

## Status
Finished

## Current
2026-09-09 角色授权统一使用 AuthorizationGrantBadge，原始 grant.source 不直接展示；中文来源支持 hover/focus。真实本地双宽度 Playwright、类型检查和构建通过。验收证据 `E:/tmp/yueli-docs-publish-20260909/grants-resource-*`。未部署线上。

已接入共享分页与标题工具区。CLI Playwright 已验证 /manage, /manage/categories, /manage/tags，覆盖 390/1440 宽度，无 pageerror。首次构建遇到 V8 堆上限；仅本次构建设置 8 GiB 后通过，未修改产品运行配置。

## Next
None

## References
- [共享规则](../../../../foundation/flightdeck/knowledge/frontend/compact-admin-collections.md)
- 本次浏览器证据：`E:/tmp/yueli-media-preset-20260908/resource-all-admin-browser.json` 与 screenshots。
- 本地 URL Catalog 从旧 IP Origin 精确迁移到声明的开发域名：先比对旧 version/digest，事务更新 catalog digest/revision，再由新 Catalog 构造器确认；历史 URL 保留，未修改 SQL 迁移文件。备份在本次外部验收目录。

## Final verification
- 最终生产构建通过：`resource-admin-final-build-8gb.log`；本轮仅推广后台布局，未重新验收支付等无关业务。
- 类型检查、CLI Playwright 页面与相关交互、改动文件空白检查通过。额外验收组合已通过 Workspace CLI 停止，原有共享服务与 Gallery / Blog 保留。
- [页面验收结果](references/browser-layout.json)；截图及交互日志保存在本轮外部制品目录。
