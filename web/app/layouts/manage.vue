<script setup lang="ts">
import type {
  AdminNavigationItem,
  AdminSearchGroup,
  AdminShellMessages,
} from "@yueli/ui/admin";

const route = useRoute();
const { brand } = useSiteRuntime();
const { can, isAdministrator } = useResourceMe();

const messages: AdminShellMessages = {
  skipToContent: "跳到主要内容",
  search: "搜索控制台",
  searchPlaceholder: "搜索页面与常用操作",
  currentLocation: "当前位置",
};

function active(path: string, exact = false) {
  return exact ? route.path === path : route.path.startsWith(path);
}

const navigation = computed<readonly AdminNavigationItem[]>(() => [
  ...(can("resource.item.read") || can("resource.item.create")
    ? [
        {
          label: "控制台",
          icon: "i-tabler-dashboard",
          to: "/manage/dashboard",
          active: active("/manage/dashboard"),
        },
        {
          label: "资源",
          icon: "i-tabler-package",
          to: "/manage",
          active: active("/manage", true)
            || (/^\/manage\/[^/]+$/.test(route.path)
              && !new Set(["/manage/dashboard", "/manage/categories", "/manage/tags", "/manage/settings", "/manage/assets", "/manage/authorization"]).has(route.path)),
        },
      ]
    : []),
  ...(can("resource.taxonomy.manage")
    ? [
        {
          label: "分类",
          icon: "i-tabler-folder",
          to: "/manage/categories",
          active: active("/manage/categories"),
        },
        {
          label: "标签",
          icon: "i-tabler-tags",
          to: "/manage/tags",
          active: active("/manage/tags"),
        },
      ]
    : []),
  ...(can("resource.site_settings.manage")
    ? [{
        label: "站点设置",
        icon: "i-tabler-settings",
        to: "/manage/settings",
        active: active("/manage/settings"),
      }]
    : []),
  ...(can("resource.asset_settings.manage")
    ? [{
        label: "资源策略",
        icon: "i-tabler-database-cog",
        to: "/manage/assets",
        active: active("/manage/assets"),
      }]
    : []),
  ...(isAdministrator.value
    ? [{
        label: "权限与申请",
        icon: "i-tabler-shield-lock",
        to: "/manage/authorization",
        active: active("/manage/authorization"),
      }]
    : []),
]);

const searchGroups = computed<readonly AdminSearchGroup[]>(() => {
  const pages = navigation.value.map((item, index) => ({
    id: `resource-page-${index}`,
    label: item.label,
    icon: item.icon,
    to: item.to,
  }));
  const actions = [
    ...(can("resource.item.create")
      ? [{
          id: "new-resource",
          label: "新建资源",
          icon: "i-tabler-package-import",
          to: "/manage",
        }]
      : []),
    ...(isAdministrator.value
      ? [{
          id: "authorization",
          label: "处理角色申请",
          icon: "i-tabler-user-check",
          to: "/manage/authorization",
        }]
      : []),
  ];
  return [
    { id: "resource-pages", label: "管理页面", items: pages },
    ...(actions.length
      ? [{ id: "resource-actions", label: "常用操作", items: actions }]
      : []),
  ];
});
const currentLabel = computed(() =>
  String(navigation.value.find((item) => item.active)?.label || "控制台"),
);
</script>

<template>
  <YAdminConsoleLayout
    :class="{ 'yueli-admin-branded': String(route.name || '') !== 'manage-id' }"
    :immersive="String(route.name || '') === 'manage-id'"
    :navigation="navigation"
    :search-groups="searchGroups"
    :messages="messages"
    storage-key="resource-manage"
    main-id="manage-main"
    :brand-label="brand"
    brand-icon="i-tabler-package"
    brand-to="/"
    :context-label="brand"
    :current-label="currentLabel"
    back-to-top-label="返回顶部"
    data-resource-manage-shell
  >
    <template #topbar-right>
      <ConsumerManageAccountControl
        home-to=""
        show-appearance
        trigger-mode="inline"
      />
    </template>
    <slot />
  </YAdminConsoleLayout>
</template>
