<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'

const { user, loggedIn, login, logout } = useAuth()
const { siteSettings, error: settingsError } = useResourceSettings()
const { brand: siteBrand } = useSiteRuntime()

const router = useRouter()
const searchQ = ref('')
function goSearch() {
  const v = searchQ.value.trim()
  if (v) void router.push({ path: '/search', query: { q: v } })
}
function doLogin() {
  void login()
}
function doLogout() {
  void logout()
}

const initial = computed(() => (user.value?.name || user.value?.email || '?').charAt(0).toUpperCase())
const site = computed(() => siteSettings.value.site)
const footer = computed(() => siteSettings.value.footer)
const userItems = computed<DropdownMenuItem[][]>(() => [
  [{ label: user.value?.name || user.value?.email || '', type: 'label' }],
  [
    { label: '管理控制台', icon: 'i-tabler-layout-dashboard', to: '/manage' },
    { label: '退出登录', icon: 'i-tabler-logout', onSelect: doLogout }
  ]
])
</script>

<template>
  <div class="resource-app-shell flex min-h-dvh flex-col text-default">
    <header class="resource-topbar sticky top-0 z-20 border-b">
      <div class="mx-auto flex h-16 w-full max-w-6xl items-center justify-between gap-4 px-4">
        <NuxtLink
          to="/"
          class="font-display flex items-center gap-2 text-base font-semibold text-highlighted"
        >
          <span class="grid size-8 place-items-center rounded-lg bg-primary/10 text-primary">
            <UIcon name="i-tabler-package" class="size-5" />
          </span>
          {{ site.siteName || siteBrand }}
        </NuxtLink>

        <div class="flex items-center gap-1.5">
          <UInput
            v-model="searchQ"
            icon="i-tabler-search"
            placeholder="搜索资源"
            size="sm"
            class="hidden w-32 sm:block md:w-44"
            @keyup.enter="goSearch"
          />
          <UButton to="/search" icon="i-tabler-search" color="neutral" variant="ghost" class="sm:hidden" aria-label="搜索" />
          <UColorModeButton />
          <template v-if="loggedIn">
            <UButton to="/manage" variant="ghost" color="neutral" icon="i-tabler-upload" label="管理" class="hidden sm:flex" />
            <UDropdownMenu :items="userItems" :ui="{ content: 'w-48' }">
              <UButton variant="ghost" color="neutral" class="gap-2 px-1.5">
                <UAvatar :text="initial" size="xs" />
                <span class="hidden max-w-32 truncate text-sm sm:block">{{ user?.name || user?.email }}</span>
              </UButton>
            </UDropdownMenu>
          </template>
          <UButton v-else variant="ghost" color="neutral" icon="i-tabler-login-2" label="登录" @click="doLogin" />
        </div>
      </div>
    </header>

    <main class="mx-auto w-full max-w-6xl flex-1 px-4 py-8 sm:py-10">
      <UAlert v-if="settingsError" color="error" icon="i-tabler-alert-circle" title="站点配置不可用" description="请先完成当前站点的配置 provision。" />
      <slot v-else />
    </main>

    <footer class="border-t border-default bg-default/45">
      <div class="mx-auto grid w-full max-w-6xl gap-5 px-4 py-6 text-xs text-muted md:grid-cols-[minmax(0,1fr)_auto] md:items-end">
        <div>
          <p class="font-medium text-default">{{ site.siteName || siteBrand }}</p>
          <p class="mt-1">{{ footer.tagline }}</p>
          <div class="mt-2 flex flex-wrap gap-x-3 gap-y-1">
            <NuxtLink v-if="footer.compliance?.icpRecord" :to="footer.compliance.icpUrl" target="_blank" class="hover:text-primary">{{ footer.compliance.icpRecord }}</NuxtLink>
            <NuxtLink v-if="footer.compliance?.policeRecord" :to="footer.compliance.policeUrl" target="_blank" class="hover:text-primary">{{ footer.compliance.policeRecord }}</NuxtLink>
            <span v-if="footer.compliance?.extraText">{{ footer.compliance.extraText }}</span>
          </div>
        </div>
        <div class="flex flex-wrap gap-3 md:justify-end">
          <NuxtLink
            v-for="link in footer.socialLinks?.filter(item => item.label && item.to)"
            :key="link.label + link.to"
            :to="link.to"
            target="_blank"
            class="inline-flex items-center gap-1 hover:text-primary"
          >
            <UIcon :name="link.icon || 'i-tabler-link'" class="size-3.5" />{{ link.label }}
          </NuxtLink>
          <span>{{ footer.copyright }}</span>
        </div>
      </div>
    </footer>
  </div>
</template>
