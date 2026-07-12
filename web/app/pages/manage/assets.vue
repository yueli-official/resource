<script setup lang="ts">
import { ActionFeedbackButton, ManageHeader, SkeletonList } from '@platform/manage/components'
import { useActionFeedback } from '@platform/manage/use-action-feedback'
import { useMinLoading } from '@platform/ui/use-min-loading'
import type { AssetProfileView, AssetStorageBackendView } from '~/types'

definePageMeta({ layout: 'manage', middleware: 'auth' })
useSeoMeta({ title: '资源配置 · 控制台' })

interface AssetSiteView {
  siteKey: string
  name: string
  defaultStorageBackend: string
  enabled: boolean
  assetCount: number
  profileCount: number
  variantCount: number
}

const { slug: siteSlug, brand: siteBrand } = useSiteRuntime()
const { isAdmin } = useAuth()
const { call } = useAssetAdminApi()
const { status: saveStatus, pending: markSaving, success: markSaved, reset: resetSave } = useActionFeedback()
const saveError = ref('')

const mounted = ref(false)
onMounted(() => { mounted.value = true })

const { data, pending, refresh } = await useAsyncData(
  'resource-manage-assets-settings',
  async () => {
    const [sites, backends, profiles] = await Promise.all([
      call<{ items: AssetSiteView[] }>('/api/v1/admin/assets/sites'),
      call<{ items: AssetStorageBackendView[], defaultName: string }>('/api/v1/admin/assets/storage-backends'),
      call<{ items: AssetProfileView[] }>(`/api/v1/admin/assets/profiles?siteKey=${encodeURIComponent(siteSlug.value)}`)
    ])
    return {
      site: sites.items.find(item => item.siteKey === siteSlug.value) || null,
      backends: backends.items,
      defaultBackend: backends.defaultName,
      profiles: profiles.items
    }
  },
  { server: false, default: () => ({ site: null, backends: [], defaultBackend: '', profiles: [] }) }
)

const siteForm = reactive({
  name: siteBrand.value,
  defaultStorageBackend: 'local',
  enabled: true
})
const profileForms = ref<AssetProfileView[]>([])
const expandedProfiles = ref<Record<string, boolean>>({})
const profileOrder = ['resource-cover', 'resource-content', 'resource']

watch(() => data.value, (value) => {
  const site = value.site
  siteForm.name = site?.name || siteBrand.value
  siteForm.defaultStorageBackend = site?.defaultStorageBackend || value.defaultBackend || value.backends[0]?.name || 'local'
  siteForm.enabled = site?.enabled !== false
  profileForms.value = (value.profiles || [])
    .map(profile => ({ ...profile }))
    .sort((a, b) => {
      const ai = profileOrder.indexOf(a.profileKey)
      const bi = profileOrder.indexOf(b.profileKey)
      return (ai === -1 ? 99 : ai) - (bi === -1 ? 99 : bi) || a.profileKey.localeCompare(b.profileKey)
    })
}, { immediate: true })

const showSkeleton = useMinLoading(computed(() => !mounted.value || pending.value))
const enabledBackends = computed(() => data.value.backends.filter(backend => backend.enabled !== false))
const backendItems = computed(() => enabledBackends.value.map(backend => ({
  label: `${backend.name}${backend.type ? ` (${backend.type})` : ''}`,
  value: backend.name
})))
const profileBackendItems = computed(() => [
  { label: `继承站点默认 (${siteForm.defaultStorageBackend || 'local'})`, value: '' },
  ...backendItems.value
])
const accessLevelItems = [
  { label: '公开资源 · 公开直链', value: 'public' },
  { label: '私有资源 · 签名链接', value: 'private' }
]
const profileGuides: Record<string, { title: string, icon: string, description: string, recommendation: string }> = {
  'resource-cover': {
    title: '资源封面',
    icon: 'i-tabler-photo',
    description: '用于资源卡片、详情页首图和分享预览，应该是公开图片。',
    recommendation: '建议仅允许 jpg、jpeg、png、webp，控制在 10-20 MB 内，并保留原图以便后续重裁剪。'
  },
  'resource-content': {
    title: '正文图片',
    icon: 'i-tabler-file-description',
    description: '用于资源详情正文和展示图集，应该是公开图片。',
    recommendation: '建议仅允许网页图片格式，派生展示尺寸，避免正文引用私有资源。'
  },
  resource: {
    title: '下载文件',
    icon: 'i-tabler-package',
    description: '用于资源下载文件。资源 Profile 只决定公开或私有签名访问；登录门禁、会员和积分规则应由资源站业务配置决定。',
    recommendation: '小文件可以直传；大文件建议使用 OSS/COS/S3 分片或网盘，不建议长期走本地存储。'
  }
}

function formatBytes(value: number) {
  if (!value) return '不限制'
  const units = ['B', 'KB', 'MB', 'GB']
  let size = value
  let unit = 0
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024
    unit += 1
  }
  return `${size >= 10 || unit === 0 ? Math.round(size) : size.toFixed(1)} ${units[unit]}`
}

function profileGuide(profile: AssetProfileView) {
  return profileGuides[profile.profileKey] || {
    title: profile.profileKey,
    icon: profile.defaultVisibility === 'private' ? 'i-tabler-lock' : 'i-tabler-folder-cog',
    description: profile.purpose || '按用途配置资源上传和交付规则。',
    recommendation: '按真实用途限制后缀、大小、可见性和存储后端。'
  }
}

function profileStorageLabel(profile: AssetProfileView) {
  return profile.storageBackend || `继承 ${siteForm.defaultStorageBackend || 'local'}`
}

function profileExpanded(profile: AssetProfileView) {
  return expandedProfiles.value[profile.profileKey] === true
}

function toggleProfile(profile: AssetProfileView) {
  expandedProfiles.value = {
    ...expandedProfiles.value,
    [profile.profileKey]: !profileExpanded(profile)
  }
}

function accessLevelLabel(profile: AssetProfileView) {
  return profile.defaultVisibility === 'private' ? '私有签名链接' : '公开直链'
}

function normalizeProfileDeliveryPolicy(profile: AssetProfileView) {
  if (profile.defaultVisibility === 'public') {
    profile.defaultDeliveryPolicy = 'public'
    return
  }
  if (!profile.defaultDeliveryPolicy || profile.defaultDeliveryPolicy !== 'signed') {
    profile.defaultDeliveryPolicy = 'signed'
  }
}

watch(profileForms, (profiles) => {
  for (const profile of profiles) normalizeProfileDeliveryPolicy(profile)
}, { deep: true })

async function save() {
  markSaving()
  saveError.value = ''
  try {
    await call('/api/v1/admin/assets/sites', {
      method: 'POST',
      body: {
        siteKey: siteSlug.value,
        name: siteForm.name,
        defaultStorageBackend: siteForm.defaultStorageBackend,
        enabled: siteForm.enabled
      }
    })
    await Promise.all(profileForms.value.map(profile => call('/api/v1/admin/assets/profiles', {
      method: 'POST',
      body: {
        siteKey: profile.siteKey,
        profileKey: profile.profileKey,
        purpose: profile.purpose,
        allowedExt: profile.allowedExt,
        maxSizeBytes: Number(profile.maxSizeBytes || 0),
        defaultVisibility: profile.defaultVisibility,
        defaultDeliveryPolicy: profile.defaultVisibility === 'public' ? 'public' : 'signed',
        storageBackend: profile.storageBackend || '',
        keepOriginal: profile.keepOriginal
      }
    })))
    markSaved()
    await refresh()
  } catch (err) {
    resetSave()
    saveError.value = (err as Error).message
  }
}
</script>

<template>
  <div class="space-y-5">
    <ManageHeader title="资源配置">
      <template #subtitle>配置资源站如何消费资源中心：默认存储、上传用途、大小限制和访问级别。</template>
      <template #actions>
        <ActionFeedbackButton v-if="isAdmin" :status="saveStatus" idle-label="保存" pending-label="保存中" success-label="已保存" @click="save" />
      </template>
    </ManageHeader>

    <UAlert v-if="saveError" color="error" variant="subtle" icon="i-tabler-alert-circle" title="保存失败" :description="saveError" role="alert" />

    <SkeletonList v-if="showSkeleton" :rows="4" />

    <UAlert
      v-else-if="!isAdmin"
      color="warning"
      icon="i-tabler-shield-lock"
      title="需要管理员权限"
      description="资源配置会影响全站上传、存储和访问级别，请使用管理员账户操作。"
    />

    <template v-else>
      <section class="rounded-lg border border-default bg-default p-5">
        <div class="mb-4 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 class="font-medium text-highlighted">站点资源策略</h2>
            <p class="mt-1 text-sm text-muted">站点默认后端可被每个用途 profile 单独覆盖。</p>
          </div>
          <UBadge :label="siteForm.enabled ? '已启用' : '已停用'" :color="siteForm.enabled ? 'success' : 'warning'" variant="soft" />
        </div>
        <div class="grid gap-4 md:grid-cols-[minmax(0,1fr)_18rem_8rem]">
          <UFormField label="站点名称">
            <UInput v-model="siteForm.name" class="w-full" />
          </UFormField>
          <UFormField label="默认存储后端">
            <USelectMenu v-model="siteForm.defaultStorageBackend" :items="backendItems" value-key="value" class="w-full" />
          </UFormField>
          <UFormField label="启用">
            <USwitch v-model="siteForm.enabled" />
          </UFormField>
        </div>
      </section>

      <section class="overflow-hidden rounded-lg border border-default bg-default">
        <div class="border-b border-default px-5 py-4">
          <h2 class="font-medium text-highlighted">用途 Profile</h2>
          <p class="mt-1 text-sm text-muted">Profile 决定某类资源使用哪个后端、允许上传什么、最大多大、以及默认访问级别。</p>
        </div>
        <div class="divide-y divide-default">
          <article v-for="profile in profileForms" :key="profile.profileKey" class="p-5">
            <div class="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
              <div class="flex min-w-0 items-start gap-3">
                <span class="grid size-10 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
                  <UIcon :name="profileGuide(profile).icon" class="size-5" />
                </span>
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-2">
                    <h3 class="font-medium text-highlighted">{{ profileGuide(profile).title }}</h3>
                    <UBadge :label="profile.profileKey" color="neutral" variant="soft" size="sm" />
                    <UBadge :label="accessLevelLabel(profile)" :color="profile.defaultVisibility === 'private' ? 'warning' : 'success'" variant="soft" size="sm" />
                  </div>
                  <p class="mt-1 max-w-3xl text-sm leading-6 text-muted">{{ profileGuide(profile).description }}</p>
                  <div class="mt-3 flex flex-wrap gap-2 text-xs text-muted">
                    <span class="rounded-md bg-elevated/60 px-2 py-1">后端：{{ profileStorageLabel(profile) }}</span>
                    <span class="rounded-md bg-elevated/60 px-2 py-1">上限：{{ formatBytes(profile.maxSizeBytes) }}</span>
                    <span class="rounded-md bg-elevated/60 px-2 py-1">后缀：{{ profile.allowedExt || '未限制' }}</span>
                    <span class="rounded-md bg-elevated/60 px-2 py-1">{{ profile.keepOriginal ? '保留原始文件' : '只保留处理结果' }}</span>
                    <span class="rounded-md bg-elevated/60 px-2 py-1">素材 {{ profile.assetCount }} · 派生 {{ profile.variantCount }}</span>
                  </div>
                </div>
              </div>

              <UButton
                :icon="profileExpanded(profile) ? 'i-tabler-chevron-up' : 'i-tabler-adjustments-horizontal'"
                :label="profileExpanded(profile) ? '收起' : '编辑配置'"
                color="neutral"
                variant="soft"
                size="sm"
                @click="toggleProfile(profile)"
              />
            </div>

            <div v-if="profileExpanded(profile)" class="mt-5 rounded-lg border border-default bg-elevated/30 p-4">
              <p class="mb-4 text-xs leading-5 text-muted">{{ profileGuide(profile).recommendation }}</p>
              <div class="grid gap-3 sm:grid-cols-2">
                <UFormField label="存储后端">
                  <USelectMenu v-model="profile.storageBackend" :items="profileBackendItems" value-key="value" class="w-full" />
                </UFormField>
                <UFormField label="最大大小（bytes）">
                  <UInput v-model.number="profile.maxSizeBytes" type="number" min="0" class="w-full" />
                </UFormField>
                <UFormField label="访问级别" help="公开资源返回稳定公开地址；私有资源由业务授权后生成签名链接。">
                  <USelectMenu
                    v-model="profile.defaultVisibility"
                    :items="accessLevelItems"
                    value-key="value"
                    class="w-full"
                  />
                </UFormField>
              </div>
              <UFormField class="mt-3" label="允许后缀">
                <UInput v-model="profile.allowedExt" class="w-full" placeholder="zip,7z,pdf,jpg,png,webp" />
              </UFormField>
              <UFormField class="mt-3" label="原始文件策略">
                <div class="flex min-h-8 items-center gap-3">
                  <USwitch v-model="profile.keepOriginal" />
                  <span class="text-sm text-default">{{ profile.keepOriginal ? '保留原始文件' : '只保留处理结果' }}</span>
                </div>
              </UFormField>
            </div>
          </article>
        </div>
      </section>

      <UAlert
        color="neutral"
        variant="subtle"
        icon="i-tabler-route"
        title="资源规则按用途生效"
        description="默认后端用于普通资源；每个 Profile 可以单独覆盖为 local、OSS、COS 或 S3 兼容后端。上传时会按该用途校验后缀、大小和访问级别。"
      />
    </template>
  </div>
</template>
