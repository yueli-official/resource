<script setup lang="ts">
import { ManageShell } from "@platform/manage/components";

const route = useRoute();
const { brand: siteBrand } = useSiteRuntime();

const reservedRoutes = new Set([
  "/manage/dashboard",
  "/manage/categories",
  "/manage/tags",
  "/manage/settings",
  "/manage/assets",
  "/manage/taxonomy",
]);
const isEditor = computed(
  () => /^\/manage\/[^/]+$/.test(route.path) && !reservedRoutes.has(route.path),
);
const showBackToTop = computed(() =>
  ["/manage/dashboard", "/manage/settings", "/manage/assets"].includes(
    route.path,
  ),
);
const contextLabel = computed(() => {
  if (route.path === "/manage/dashboard") return "状态";
  if (route.path === "/manage") return "资源";
  if (isEditor.value) return "编辑资源";
  if (route.path === "/manage/categories") return "分类";
  if (route.path === "/manage/tags") return "标签";
  if (route.path === "/manage/assets") return "资源配置";
  if (route.path === "/manage/settings") return "站点设置";
  return "资源后台";
});
</script>

<template>
  <ManageShell
    :site-name="siteBrand"
    :context-label="contextLabel"
    home-to="/manage/dashboard"
    storage-key="resource-manage"
    shell-class="resource-app-shell"
    :show-back-to-top="showBackToTop"
  >
    <template #sidebar><ManageSidebar /></template>
    <template #user>
      <ConsumerManageAccountControl home-to="/" />
    </template>
    <slot />
  </ManageShell>
</template>
