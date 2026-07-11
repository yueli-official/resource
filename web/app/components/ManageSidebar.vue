<script setup lang="ts">
const route = useRoute()
const { isAdmin } = useAuth()
const { brand: siteBrand } = useSiteRuntime()

const nav = computed(() => [
  { label: '资源', icon: 'i-tabler-package', to: '/manage' },
  ...(isAdmin.value ? [
    { label: '分类', icon: 'i-tabler-folder', to: '/manage/categories' },
    { label: '标签', icon: 'i-tabler-tags', to: '/manage/tags' },
    {
      label: '设置',
      icon: 'i-tabler-settings',
      to: '/manage/settings',
      children: [
        { label: '首页', icon: 'i-tabler-home-cog', to: '/manage/settings?section=home', section: 'home' },
        { label: '页脚', icon: 'i-tabler-layout-bottombar', to: '/manage/settings?section=footer', section: 'footer' },
        { label: '参数', icon: 'i-tabler-adjustments-horizontal', to: '/manage/settings?section=resource', section: 'resource' }
      ]
    },
    { label: '资源配置', icon: 'i-tabler-database-cog', to: '/manage/assets' }
  ] : [])
])

function isActive(to: string) {
  const path = to.split('?')[0] || to
  if (path === '/manage') {
    return route.path === '/manage' || isResourceEditorPath(route.path)
  }
  return route.path === path || route.path.startsWith(`${path}/`)
}

function isResourceEditorPath(path: string) {
  if (!/^\/manage\/[^/]+$/.test(path)) return false
  return !new Set(['/manage/categories', '/manage/tags', '/manage/settings', '/manage/assets', '/manage/taxonomy']).has(path)
}

function isChildActive(section?: string) {
  return route.path === '/manage/settings' && (route.query.section || 'home') === section
}
</script>

<template>
  <div class="flex h-full flex-col bg-elevated/30">
    <NuxtLink to="/manage" class="font-display flex h-16 items-center gap-2 border-b border-default px-5 font-semibold text-highlighted">
      <span class="grid size-8 place-items-center rounded-lg bg-primary/10 text-primary"><UIcon name="i-tabler-package" class="size-5" /></span>
      {{ siteBrand }}
    </NuxtLink>

    <nav class="flex-1 space-y-1 p-3">
      <div v-for="item in nav" :key="item.to" class="space-y-1">
        <NuxtLink
          :to="item.to"
          class="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition"
          :class="isActive(item.to) ? 'bg-primary/10 text-primary' : 'text-muted hover:bg-elevated hover:text-default'"
        >
          <UIcon :name="item.icon" class="size-5 shrink-0" />
          <span class="min-w-0 flex-1 truncate">{{ item.label }}</span>
          <UIcon v-if="item.children?.length" name="i-tabler-chevron-down" class="size-4 shrink-0 opacity-70" />
        </NuxtLink>
        <div v-if="item.children?.length && isActive(item.to)" class="ml-4 space-y-1 border-l border-default pl-3">
          <NuxtLink
            v-for="child in item.children"
            :key="child.to"
            :to="child.to"
            class="flex items-center gap-2 rounded-md px-2 py-1.5 text-xs font-medium transition"
            :class="isChildActive(child.section) ? 'bg-primary/10 text-primary' : 'text-muted hover:bg-elevated hover:text-default'"
          >
            <UIcon :name="child.icon" class="size-4 shrink-0" />
            <span class="truncate">{{ child.label }}</span>
          </NuxtLink>
        </div>
      </div>
    </nav>
  </div>
</template>
