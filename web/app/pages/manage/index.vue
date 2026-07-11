<script setup lang="ts">
import { abs } from '@platform/ui/date'
import { ManageEmpty, ManageHeader, ManagePageFooter, ManagePagination, ManageTabs, SkeletonList } from '@platform/manage/components'
import { useMinLoading } from '@platform/ui/use-min-loading'
import type { ResourceView, MyResources } from '~/types'

definePageMeta({ layout: 'manage', middleware: 'auth' })
useSeoMeta({ title: '资源管理 · 控制台' })

const { user } = useAuth()
const { call } = useApi()
const toast = useToast()

const ALL = '__all__'
const q = ref('')
const status = ref(ALL)
const sortBy = ref<'title' | 'createdAt' | 'updatedAt' | 'publishedAt' | 'viewCount' | 'downloadCount'>('updatedAt')
const sortOrder = ref<'asc' | 'desc'>('desc')
const viewMode = ref<'list' | 'grid'>('list')
const page = ref(1)
const pageSize = ref(24)
const mounted = ref(false)
onMounted(() => { mounted.value = true })

const { data, pending } = await useAsyncData(
  () => `my-resources-${q.value}-${status.value}-${sortBy.value}-${sortOrder.value}-${page.value}-${pageSize.value}`,
  () => call<MyResources>('/api/v1/resources/mine', {
    query: {
      q: q.value.trim() || undefined,
      status: status.value === ALL ? undefined : status.value,
      sortBy: sortBy.value,
      sortOrder: sortOrder.value,
      page: page.value,
      size: pageSize.value
    }
  }),
  {
    server: false,
    watch: [q, status, sortBy, sortOrder, page, pageSize],
    default: () => ({ items: [] as ResourceView[], total: 0, page: 1, size: 24 })
  }
)

const resources = computed(() => data.value?.items ?? [])
const total = computed(() => data.value?.total ?? 0)
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const showSkeleton = useMinLoading(computed(() => !mounted.value || pending.value))

const types = [
  { label: '软件 / 工具', value: 'software' },
  { label: '设计素材 / 模板', value: 'design' },
  { label: '脚本', value: 'script' },
  { label: '其它', value: 'default' }
]
const statusTabs = computed(() => [
  { key: ALL, label: '全部', count: status.value === ALL ? total.value : undefined },
  { key: 'published', label: '已发布', count: status.value === 'published' ? total.value : undefined },
  { key: 'draft', label: '草稿', count: status.value === 'draft' ? total.value : undefined },
  { key: 'archived', label: '归档', count: status.value === 'archived' ? total.value : undefined }
])
const sortItems = [
  { label: '按名称', value: 'title' },
  { label: '按创建日期', value: 'createdAt' },
  { label: '按更新日期', value: 'updatedAt' },
  { label: '按发布时间', value: 'publishedAt' },
  { label: '按浏览量', value: 'viewCount' },
  { label: '按下载量', value: 'downloadCount' }
]
const pageSizeItems = [
  { label: '12/页', value: 12 },
  { label: '24/页', value: 24 },
  { label: '48/页', value: 48 },
  { label: '96/页', value: 96 }
]

const showCreate = ref(false)
const form = reactive({ title: '', type: 'software', summary: '' })
const creating = ref(false)

watch([q, status, sortBy, sortOrder, pageSize], () => { page.value = 1 })

function openCreateModal() {
  showCreate.value = true
}

async function create() {
  if (!form.title.trim()) return
  creating.value = true
  try {
    const res = await call<{ resource: ResourceView }>('/api/v1/resources', { method: 'POST', body: { ...form } })
    toast.add({ title: '已创建草稿', color: 'success', icon: 'i-tabler-check' })
    showCreate.value = false
    await navigateTo(`/manage/${res.resource.id}`)
  } catch (e: any) {
    toast.add({ title: '创建失败', description: e?.data?.message || '请重试', color: 'error' })
  } finally {
    creating.value = false
  }
}

function typeLabel(value: string) {
  return types.find(t => t.value === value)?.label || value
}

function statusLabel(value: string) {
  if (value === 'published') return '已发布'
  if (value === 'archived') return '归档'
  return '草稿'
}

function statusColor(value: string) {
  if (value === 'published') return 'success'
  if (value === 'archived') return 'neutral'
  return 'warning'
}

function toggleSortOrder() {
  sortOrder.value = sortOrder.value === 'desc' ? 'asc' : 'desc'
}

function setViewMode(mode: 'list' | 'grid') {
  viewMode.value = mode
}

function resourceMeta(resource: ResourceView) {
  return [
    resource.createdAt ? `创建 ${abs(resource.createdAt)}` : '',
    resource.updatedAt ? `更新 ${abs(resource.updatedAt)}` : ''
  ].filter(Boolean).join(' · ')
}
</script>

<template>
  <div class="space-y-6">
    <ManageHeader title="资源管理">
      <template #subtitle>已登录:<span class="text-default">{{ user?.name || user?.email }}</span> · 管理资源、状态、封面和下载文件。</template>
      <template #actions>
        <UButton icon="i-tabler-plus" label="新建资源" @click="openCreateModal" />
      </template>
    </ManageHeader>

    <ManageTabs v-model="status" :items="statusTabs" />

    <div class="rounded-lg border border-default bg-default p-3">
      <div class="flex flex-wrap items-center gap-2">
        <UInput v-model="q" icon="i-tabler-search" placeholder="搜索标题、摘要或描述" size="sm" class="w-full sm:w-64" />
        <USelectMenu v-model="sortBy" :items="sortItems" value-key="value" icon="i-tabler-arrows-sort" size="sm" class="w-[calc(100%-2.5rem)] sm:w-40" />
        <UButton
          :icon="sortOrder === 'desc' ? 'i-tabler-sort-descending' : 'i-tabler-sort-ascending'"
          color="neutral"
          variant="soft"
          size="sm"
          square
          :aria-label="sortOrder === 'desc' ? '切换为正序' : '切换为反序'"
          @click="toggleSortOrder"
        />
        <USelect v-model="pageSize" :items="pageSizeItems" size="sm" class="w-full sm:w-24" />
        <div class="ml-auto flex items-center gap-0.5 rounded-lg bg-elevated/60 p-0.5 ring-1 ring-default">
          <UButton :variant="viewMode === 'list' ? 'soft' : 'ghost'" :color="viewMode === 'list' ? 'primary' : 'neutral'" size="xs" icon="i-tabler-list" square aria-label="列表视图" @click="setViewMode('list')" />
          <UButton :variant="viewMode === 'grid' ? 'soft' : 'ghost'" :color="viewMode === 'grid' ? 'primary' : 'neutral'" size="xs" icon="i-tabler-layout-grid" square aria-label="网格视图" @click="setViewMode('grid')" />
        </div>
      </div>
    </div>

    <SkeletonList v-if="showSkeleton" :rows="6" />
    <ManageEmpty v-else-if="!resources.length" icon="i-tabler-package-off" text="没有匹配的资源" />

    <div v-else-if="viewMode === 'list'" class="overflow-hidden rounded-lg border border-default bg-default">
      <NuxtLink
        v-for="resource in resources"
        :key="resource.id"
        :to="`/manage/${resource.id}`"
        class="grid grid-cols-[88px_minmax(0,1fr)] gap-3 border-b border-default p-3 transition last:border-b-0 hover:bg-elevated/50 md:grid-cols-[104px_minmax(0,1fr)_13rem] md:gap-4 md:p-4"
      >
        <div class="h-20 overflow-hidden rounded-md border border-default bg-elevated md:h-[78px]">
          <img v-if="resource.coverUrl" :src="resource.coverUrl" :alt="resource.title" class="size-full object-cover">
          <div v-else class="grid size-full place-items-center bg-primary/10 text-primary">
            <UIcon name="i-tabler-package" class="size-6" />
          </div>
        </div>
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <h2 class="line-clamp-1 text-base font-semibold text-highlighted">{{ resource.title }}</h2>
            <UBadge v-if="resource.status !== 'published'" :label="statusLabel(resource.status)" :color="statusColor(resource.status)" variant="subtle" />
          </div>
          <p class="mt-1 line-clamp-1 text-sm text-muted md:line-clamp-2">{{ resource.summary || '未填写摘要' }}</p>
          <div class="mt-3 hidden flex-wrap gap-1.5 sm:flex">
            <UBadge :label="typeLabel(resource.type)" color="primary" variant="soft" />
            <UBadge v-for="tag in resource.tags.slice(0, 3)" :key="tag" :label="`#${tag}`" color="neutral" variant="subtle" />
          </div>
        </div>
        <div class="col-span-2 flex items-end justify-between gap-3 text-sm md:col-span-1 md:flex-col md:items-end md:justify-between md:text-right">
          <div>
            <p class="font-medium text-highlighted">{{ resource.downloadCount }} 次下载</p>
            <p class="mt-1 text-xs text-muted">/{{ resource.slug }}</p>
          </div>
          <ClientOnly>
            <p class="hidden text-xs text-muted sm:block">{{ resourceMeta(resource) }}</p>
          </ClientOnly>
        </div>
      </NuxtLink>
    </div>

    <div v-else class="grid gap-3 [grid-template-columns:repeat(auto-fill,minmax(min(13rem,100%),1fr))]">
      <NuxtLink
        v-for="resource in resources"
        :key="resource.id"
        :to="`/manage/${resource.id}`"
        class="group overflow-hidden rounded-lg border border-default bg-default transition hover:-translate-y-0.5 hover:shadow-sm"
      >
        <div class="h-[clamp(7.5rem,13vw,9.5rem)] overflow-hidden border-b border-default bg-elevated">
          <img v-if="resource.coverUrl" :src="resource.coverUrl" :alt="resource.title" class="size-full object-cover transition duration-300 group-hover:scale-105">
          <div v-else class="grid size-full place-items-center bg-primary/10 text-primary">
            <UIcon name="i-tabler-package" class="size-8" />
          </div>
        </div>
        <div class="space-y-2.5 p-3">
          <div class="min-w-0">
            <div class="flex items-start justify-between gap-2">
              <h2 class="line-clamp-1 text-sm font-semibold text-highlighted">{{ resource.title }}</h2>
              <UBadge v-if="resource.status !== 'published'" :label="statusLabel(resource.status)" :color="statusColor(resource.status)" variant="subtle" size="sm" />
            </div>
            <p class="mt-1 line-clamp-1 text-xs text-muted">{{ resource.summary || '未填写摘要' }}</p>
          </div>
          <div class="flex min-h-5 flex-wrap gap-1">
            <UBadge :label="typeLabel(resource.type)" color="primary" variant="soft" size="sm" />
            <UBadge v-for="tag in resource.tags.slice(0, 1)" :key="tag" :label="`#${tag}`" color="neutral" variant="subtle" size="sm" />
          </div>
          <div class="flex items-end justify-between gap-2 border-t border-default pt-2.5">
            <div>
              <p class="text-xs text-muted">下载</p>
              <p class="text-sm font-semibold text-highlighted">{{ resource.downloadCount }}</p>
            </div>
            <UIcon name="i-tabler-arrow-up-right" class="size-4 text-muted opacity-0 transition group-hover:opacity-100" />
          </div>
        </div>
      </NuxtLink>
    </div>

    <ManagePageFooter>
      <template #left>
        <span>共 {{ total }} 条资源</span>
        <span v-if="resources.length">当前页 {{ resources.length }} 条</span>
      </template>
      <template #right>
        <ManagePagination v-model="page" :total-pages="totalPages" />
      </template>
    </ManagePageFooter>

    <UModal v-model:open="showCreate" title="新建资源(草稿)" :ui="{ footer: 'justify-end' }">
      <template #body>
        <div class="space-y-4">
          <UFormField label="标题" required>
            <UInput v-model="form.title" placeholder="例如:My CLI Tool" class="w-full" autofocus />
          </UFormField>
          <UFormField label="类型">
            <USelect v-model="form.type" :items="types" class="w-full" />
          </UFormField>
          <UFormField label="简介">
            <UTextarea v-model="form.summary" :rows="2" class="w-full" placeholder="一句话介绍这个资源" />
          </UFormField>
        </div>
      </template>
      <template #footer="{ close }">
        <UButton label="取消" color="neutral" variant="outline" @click="close" />
        <UButton
          label="创建并编辑"
          icon="i-tabler-arrow-right"
          trailing
          :loading="creating"
          :disabled="!form.title.trim()"
          @click="create"
        />
      </template>
    </UModal>
  </div>
</template>
