<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'

// Traditional admin shell: fixed sidebar (lg) / drawer (mobile) + flex-1 content.
const route = useRoute()
const mobileOpen = ref(false)
const { user, logout } = useAuth()
const initial = computed(() => (user.value?.name || user.value?.email || '?').charAt(0).toUpperCase())
const accountUrl = computed(() => useRuntimeConfig().public.accountUrl || 'http://localhost:3000')
const { brand: siteBrand } = useSiteRuntime()
const userItems = computed<DropdownMenuItem[][]>(() => [
  [{ label: user.value?.name || user.value?.email || '', type: 'label' }],
  [
    { label: '返回主站', icon: 'i-tabler-arrow-back-up', to: '/' },
    { label: '用户设置', icon: 'i-tabler-user-cog', onSelect: () => navigateTo(accountUrl.value, { external: true }) }
  ],
  [{ label: '退出登录', icon: 'i-tabler-logout', onSelect: () => logout() }]
])
watch(() => route.path, () => { mobileOpen.value = false })
function openMobileNav() {
  mobileOpen.value = true
}
function closeMobileNav() {
  mobileOpen.value = false
}
</script>

<template>
  <div class="resource-app-shell flex min-h-dvh text-default">
    <!-- sidebar: desktop -->
    <aside class="hidden w-60 shrink-0 border-r border-default lg:block">
      <div class="sticky top-0 h-dvh"><ManageSidebar /></div>
    </aside>

    <!-- sidebar: mobile drawer -->
    <div v-if="mobileOpen" class="fixed inset-0 z-40 lg:hidden">
      <div class="absolute inset-0 bg-black/40" @click="closeMobileNav" />
      <div class="absolute inset-y-0 left-0 w-64 border-r border-default bg-default shadow-xl"><ManageSidebar /></div>
    </div>

    <!-- content -->
    <div class="flex min-w-0 flex-1 flex-col">
      <header class="resource-topbar sticky top-0 z-20 flex h-16 items-center gap-3 border-b px-4">
        <UButton class="lg:hidden" icon="i-tabler-menu-2" color="neutral" variant="ghost" aria-label="菜单" @click="openMobileNav" />
        <NuxtLink to="/manage" class="font-display font-semibold text-highlighted lg:hidden">{{ siteBrand }}</NuxtLink>
        <div class="ml-auto flex items-center gap-1">
          <UColorModeButton aria-label="切换夜间模式" />
          <UDropdownMenu :items="userItems" :ui="{ content: 'w-52' }">
            <UButton variant="ghost" color="neutral" class="gap-2 px-1.5">
              <UAvatar :text="initial" size="xs" />
              <span class="hidden max-w-36 truncate text-sm sm:block">{{ user?.name || user?.email }}</span>
            </UButton>
          </UDropdownMenu>
        </div>
      </header>
      <main class="flex-1 px-5 py-6 lg:px-10 lg:py-9">
        <div class="mx-auto max-w-6xl">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>
