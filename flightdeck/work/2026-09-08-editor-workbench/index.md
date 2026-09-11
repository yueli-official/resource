# resource 内容编辑工作台

Status: Finished

## 已完成

- 接入共享 EditorCommandBar：沉浸、公开页、设置、发布/下架、保存；窄屏标题与操作分行。
- 列表提供公开页、快速编辑、完整编辑与双击编辑；选择框与按钮不会误触发导航。
- 发布前保存当前内容；下架意味着匿名原链接不可读，不提供仅链接可见选项。
- 本地 Playwright 覆盖 390/768/1024/1440px、浅深色、沉浸/设置、快速编辑、双击、上下架实际 200→404→200。
- 类型检查和构建通过；[浏览器证据](references/browser-local.json)。

## 交付边界

保留本地实现，未首次部署该站。临时验收数据已清理，自建隔离组合已由 Workspace CLI 停止。

## 共享编辑器依赖同步（完成）
2026-09-09：固定新版内容组件制品，验证文档/图片导入与外链失败占位；仅本地验收。

## 共享编辑器依赖固定（2026-09-09 完成）
package.json 与 pnpm 锁固定 `content-nuxt 0.2.3-server.20260909.placeholder.1`，将已验证包放入 web/vendor；五个站点的编辑器包 SHA-256 一致。配套 UI、Asset、HTTP 与会话依赖固定为已经部署验收的 WWW 组合，Tiptap 全组 3.31.3，避免旧正式包缺少当前消费接口。此举不代表发布 npm/GitHub Release。
固定包的独立安装、类型检查和生产构建均通过。Workspace 本地真实依赖 Playwright：URL 图片尝试下载失败留下原图链接占位、正文保留、剪贴板 PNG 经 Asset 上传、保存后刷新回读、390/1440 导入弹窗均通过，无 pageerror。只做本地验收，未首次上线；本次临时组合已停止。
Resource 对齐 Nuxt 4.5.1 / Vue 3.5.39 / Router 5.1.0，消除混用 Vue 实例导致的本地 SSR ce 错误。
依赖审计：E:/tmp/yueli-editor-consumers-20260909/dependencies.json。保留其他未提交改动；本轮未提交或推送 Git。
