<script setup lang="ts">
import { resourceDashboardMessages } from "~/utils/dashboard";
import { SkeletonList } from "~/utils/manageComponents";
import { abs } from "~/utils/date";
import { useMinimumLoading } from '@yueli/ui/feedback'
import { DashboardLayout } from '@yueli/ui/dashboard/pattern'
import type { MyResources, ResourceLifecycleCounts } from '~/types'

definePageMeta({ layout: 'manage', middleware: 'auth' })
useSeoMeta({ title: '状态 · 资源后台' })

const { can } = useResourceMe()
const canManageTaxonomy = computed(() => can('resource.taxonomy.manage'))
const canManageAssets = computed(() => can('resource.asset_settings.manage'))
const { call } = useApi()
const emptyCounts: ResourceLifecycleCounts = { all: 0, published: 0, draft: 0, archived: 0, issues: 0 }
const mounted = ref(false)
onMounted(() => { mounted.value = true })

const { data, pending, error } = await useAsyncData(
  'resource-dashboard',
  () => call<MyResources>('/api/v1/resources/mine', {
    query: { page: 1, size: 6, sortBy: 'updatedAt', sortOrder: 'desc' },
  }),
  { server: false, default: () => ({ items: [], total: 0, page: 1, size: 6, counts: emptyCounts }) },
)
const showSkeleton = useMinimumLoading(computed(() => !mounted.value || pending.value))
const resources = computed(() => data.value?.items ?? [])
const counts = computed(() => data.value?.counts ?? emptyCounts)
const metrics = computed(() => [
  { label: '资源总数', value: counts.value.all, icon: 'i-tabler-package', to: '/manage' },
  { label: '已发布', value: counts.value.published, icon: 'i-tabler-circle-check', to: '/manage?status=published' },
  { label: '草稿', value: counts.value.draft, icon: 'i-tabler-pencil', to: '/manage?status=draft' },
  { label: '待完善', value: counts.value.issues, icon: 'i-tabler-alert-circle', to: '/manage?status=issues' },
])
</script>

<template>
  <DashboardLayout title="资源状态" description="查看发布队列、继续最近更新，并确认当前资源服务可工作。" :messages="resourceDashboardMessages">
    <template #actions><UButton to="/manage" icon="i-tabler-package" label="管理资源" /></template>

    <template #metrics>
      <div v-if="showSkeleton" class="grid grid-cols-2 gap-3 lg:grid-cols-4"><USkeleton v-for="item in 4" :key="item" class="h-24 rounded-xl" /></div>
      <div v-else class="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <NuxtLink v-for="metric in metrics" :key="metric.label" :to="metric.to" class="rounded-xl border border-default bg-default p-4 transition hover:border-primary/40 hover:bg-elevated/30"><div class="flex items-center gap-3"><span class="grid size-10 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary"><UIcon :name="metric.icon" class="size-5" /></span><div class="min-w-0"><p class="text-xl font-semibold text-highlighted tabular-nums sm:text-2xl">{{ metric.value }}</p><p class="truncate text-xs text-muted">{{ metric.label }}</p></div></div></NuxtLink>
      </div>
    </template>

    <template #pending>
      <div v-if="showSkeleton" class="grid gap-3 sm:grid-cols-2"><USkeleton v-for="item in 2" :key="item" class="h-20 rounded-lg" /></div>
      <div v-else-if="counts.issues || counts.draft" class="grid gap-3 sm:grid-cols-2">
        <NuxtLink v-if="counts.issues" to="/manage?status=issues" class="group flex min-h-20 items-center gap-3 rounded-lg border border-default px-4 py-3 transition hover:border-primary/40 hover:bg-elevated/50"><span class="grid size-9 shrink-0 place-items-center rounded-lg bg-warning/10 text-warning"><UIcon name="i-tabler-alert-circle" class="size-5" /></span><div class="min-w-0 flex-1"><p class="text-sm font-medium text-highlighted">{{ counts.issues }} 个资源待完善</p><p class="text-xs text-muted">补齐发布所需信息</p></div><UIcon name="i-tabler-chevron-right" class="size-4 text-dimmed transition group-hover:translate-x-0.5" /></NuxtLink>
        <NuxtLink v-if="counts.draft" to="/manage?status=draft" class="group flex min-h-20 items-center gap-3 rounded-lg border border-default px-4 py-3 transition hover:border-primary/40 hover:bg-elevated/50"><span class="grid size-9 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary"><UIcon name="i-tabler-pencil" class="size-5" /></span><div class="min-w-0 flex-1"><p class="text-sm font-medium text-highlighted">{{ counts.draft }} 个草稿资源</p><p class="text-xs text-muted">继续编辑或发布</p></div><UIcon name="i-tabler-chevron-right" class="size-4 text-dimmed transition group-hover:translate-x-0.5" /></NuxtLink>
      </div>
      <UAlert v-else color="success" variant="subtle" icon="i-tabler-circle-check" title="当前没有待处理资源" />
    </template>

    <template #recent>
      <SkeletonList v-if="showSkeleton" :rows="6" class="p-4" />
      <div v-else-if="resources.length" class="divide-y divide-default">
        <NuxtLink v-for="resource in resources" :key="resource.id" :to="`/manage/${resource.id}`" class="group grid min-h-16 grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-3 px-4 py-3 transition hover:bg-elevated/50">
          <span class="grid size-10 place-items-center overflow-hidden rounded-lg bg-elevated"><img v-if="resource.coverUrl" :src="resource.coverUrl" :alt="resource.title" class="size-full object-cover" /><UIcon v-else name="i-tabler-package" class="size-5 text-muted" /></span>
          <div class="min-w-0"><p class="truncate text-sm font-medium text-highlighted group-hover:text-primary">{{ resource.title || '(无标题)' }}</p><p class="truncate text-xs text-muted">{{ resource.type }} · {{ resource.status === 'published' ? '已发布' : resource.status === 'archived' ? '已归档' : '草稿' }}</p></div>
          <span class="hidden text-xs text-muted sm:block">{{ abs(resource.updatedAt) }}</span>
        </NuxtLink>
      </div>
      <div v-else class="p-8 text-center text-sm text-muted">还没有资源，先创建第一项内容。</div>
    </template>

    <template #health>
      <UAlert v-if="error" color="error" variant="subtle" icon="i-tabler-alert-circle" title="资源服务暂时不可用" description="刷新后仍失败时，请到平台状态检查服务。" />
      <div v-else class="space-y-3"><div class="flex items-center justify-between gap-3 rounded-lg bg-success/10 px-3 py-2.5 text-sm"><span class="flex items-center gap-2 text-success"><UIcon name="i-tabler-circle-check" class="size-4" />资源服务可用</span><span class="text-xs text-muted">正常</span></div><p class="text-xs leading-5 text-muted">当前状态只覆盖资源编辑与发布；存储后端配置由管理员在资源配置页维护。</p></div>
    </template>

    <template #quickActions>
      <div class="grid gap-2">
        <UButton to="/manage" icon="i-tabler-package" label="管理资源" color="neutral" variant="soft" block />
        <UButton v-if="canManageTaxonomy" to="/manage/categories" icon="i-tabler-folder" label="管理分类" color="neutral" variant="soft" block />
        <UButton v-if="canManageAssets" to="/manage/assets" icon="i-tabler-database-cog" label="资源配置" color="neutral" variant="soft" block />
        <UButton to="/" icon="i-tabler-external-link" label="查看站点" color="neutral" variant="ghost" block />
      </div>
    </template>
  </DashboardLayout>
</template>
