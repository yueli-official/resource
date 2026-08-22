<script setup lang="ts">
import type {
  AdminNavigationItem,
  AdminSearchGroup,
  AdminShellMessages,
} from "@yueli/ui/admin";

const route = useRoute();
const { brand } = useSiteRuntime();
const { can, isAdministrator } = useResourceMe();
const sidebarOpen = ref(false);

const messages: AdminShellMessages = {
  skipToContent: "跳到主要内容",
  search: "搜索资源站后台",
  searchPlaceholder: "搜索页面与常用操作",
};

function closeSidebar() {
  sidebarOpen.value = false;
}

function active(path: string, exact = false) {
  return exact ? route.path === path : route.path.startsWith(path);
}

const navigation = computed<readonly AdminNavigationItem[]>(() => [
  ...(can("resource.item.read") || can("resource.item.create")
    ? [
        {
          label: "状态",
          icon: "i-tabler-dashboard",
          to: "/manage/dashboard",
          active: active("/manage/dashboard"),
          onSelect: closeSidebar,
        },
        {
          label: "资源",
          icon: "i-tabler-package",
          to: "/manage",
          active: active("/manage", true)
            || (/^\/manage\/[^/]+$/.test(route.path)
              && !new Set(["/manage/dashboard", "/manage/categories", "/manage/tags", "/manage/settings", "/manage/assets", "/manage/authorization"]).has(route.path)),
          onSelect: closeSidebar,
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
          onSelect: closeSidebar,
        },
        {
          label: "标签",
          icon: "i-tabler-tags",
          to: "/manage/tags",
          active: active("/manage/tags"),
          onSelect: closeSidebar,
        },
      ]
    : []),
  ...(can("resource.site_settings.manage")
    ? [{
        label: "站点设置",
        icon: "i-tabler-settings",
        to: "/manage/settings",
        active: active("/manage/settings"),
        onSelect: closeSidebar,
      }]
    : []),
  ...(can("resource.asset_settings.manage")
    ? [{
        label: "资源配置",
        icon: "i-tabler-database-cog",
        to: "/manage/assets",
        active: active("/manage/assets"),
        onSelect: closeSidebar,
      }]
    : []),
  ...(isAdministrator.value
    ? [{
        label: "权限与申请",
        icon: "i-tabler-shield-lock",
        to: "/manage/authorization",
        active: active("/manage/authorization"),
        onSelect: closeSidebar,
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
</script>

<template>
  <YAdminShell
    v-model:open="sidebarOpen"
    :navigation="navigation"
    :search-groups="searchGroups"
    :messages="messages"
    sidebar-appearance="commercial"
    storage-key="resource-manage"
    main-id="manage-main"
    :default-size="16"
    :min-size="14"
    :max-size="20"
  >
    <template #brand="{ collapsed }">
      <UButton
        to="/"
        color="neutral"
        variant="ghost"
        :block="!collapsed"
        :square="collapsed"
        :aria-label="`${brand}首页`"
        :class="[
          'min-h-11 gap-2 px-1.5',
          !collapsed && 'w-full justify-start',
          collapsed && 'aspect-square justify-center px-0',
        ]"
        @click="closeSidebar"
      >
        <span
          class="grid size-7 shrink-0 place-items-center rounded-md bg-primary/10 text-primary"
        >
          <UIcon name="i-tabler-package" class="size-4" />
        </span>
        <span
          v-if="!collapsed"
          class="min-w-0 truncate text-sm font-semibold text-highlighted"
        >
          {{ brand }}
        </span>
      </UButton>
    </template>

    <template #sidebar-footer="{ collapsed }">
      <ConsumerManageAccountControl
        home-to=""
        show-appearance
        :trigger-mode="collapsed ? 'collapsed' : 'sidebar'"
      />
    </template>

    <main
      id="manage-main"
      tabindex="-1"
      class="min-w-0 flex-1 overflow-y-auto p-4 outline-none sm:p-6"
    >
      <slot />
    </main>
    <YBackToTop
      target-id="manage-main"
      scroll-container-id="manage-main"
      avoid-selector="[data-manage-dock], [data-back-to-top-avoid]"
      label="返回顶部"
    />
  </YAdminShell>
</template>
