<script setup lang="ts">
import {
  ManageCollectionDock,
  ManageCollectionToolbar,
  ManageEmpty,
  ManageHeader,
  ManageLifecycleTabs,
  ManagePageSelection,
  ManagePagination,
  ManageRowShell,
  ManageTaxonomyChips,
  ManageViewToggle,
  SkeletonList
} from '@platform/manage/components'
import {
  manageCollectionQueryFingerprint,
  serializeManageCollectionQuery,
  type ManageCollectionDefinition
} from '@platform/manage/collection'
import { useManageCollectionState } from '@platform/manage/use-manage-collection-state'
import { useManageSelection } from '@platform/manage/use-manage-selection'
import { abs } from '@platform/ui/date'
import { useMinLoading } from '@platform/ui/use-min-loading'
import type { MyResources, ResourceLifecycleCounts, ResourceView } from '~/types'

definePageMeta({ layout: 'manage', middleware: 'auth' })
useSeoMeta({ title: '资源管理 · 控制台' })

const { user } = useAuth()
const { call } = useApi()
const route = useRoute()
const router = useRouter()
const mounted = ref(false)
onMounted(() => { mounted.value = true })

const collectionDefinition = {
  resourceKind: 'resource',
  statuses: ['', 'published', 'draft', 'archived', 'issues'],
  views: ['list', 'grid'],
  sortKeys: ['title', 'createdAt', 'updatedAt', 'publishedAt', 'viewCount', 'downloadCount'],
  pageSizes: [12, 24, 48, 96],
  defaultStatus: '',
  defaultView: 'list',
  defaultSort: 'updatedAt',
  defaultDirection: 'desc',
  defaultPageSize: 24,
  pagination: 'server',
  selection: 'page',
  quickEditFields: ['title', 'slug', 'summary', 'type', 'tags', 'status'],
  bulkActions: ['publish', 'draft', 'archive']
} as const satisfies ManageCollectionDefinition

const {
  status,
  searchInput,
  q,
  sort,
  direction,
  page,
  size,
  view: viewMode,
  state: collectionState
} = useManageCollectionState({
  definition: collectionDefinition,
  routeQuery: computed(() => route.query),
  replaceQuery: query => router.replace({ query })
})

const emptyCounts: ResourceLifecycleCounts = { all: 0, published: 0, draft: 0, archived: 0, issues: 0 }
const { data, pending, error, refresh } = await useAsyncData(
  'my-resources',
  () => call<MyResources>('/api/v1/resources/mine', {
    query: {
      q: q.value || undefined,
      status: status.value || undefined,
      sortBy: sort.value,
      sortOrder: direction.value,
      page: page.value,
      size: size.value
    }
  }),
  {
    server: false,
    watch: [q, status, sort, direction, page, size],
    default: () => ({ items: [], total: 0, page: 1, size: 24, counts: emptyCounts })
  }
)

const resources = computed(() => data.value?.items ?? [])
const total = computed(() => data.value?.total ?? 0)
const counts = computed(() => data.value?.counts ?? emptyCounts)
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / size.value)))
const showSkeleton = useMinLoading(computed(() => !mounted.value || pending.value))

watch(totalPages, (lastPage) => {
  if (page.value > lastPage) page.value = lastPage
}, { flush: 'sync' })

const types = [
  { label: '软件 / 工具', value: 'software' },
  { label: '设计素材 / 模板', value: 'design' },
  { label: '脚本', value: 'script' },
  { label: '其它', value: 'default' }
]
const statusTabs = computed(() => [
  { key: '', label: '全部', count: counts.value.all },
  { key: 'published', label: '已发布', count: counts.value.published },
  { key: 'draft', label: '草稿', count: counts.value.draft },
  { key: 'archived', label: '归档', count: counts.value.archived },
  { key: 'issues', label: '待完善', count: counts.value.issues }
])
const sortItems = [
  { label: '按名称', value: 'title' },
  { label: '按创建日期', value: 'createdAt' },
  { label: '按更新日期', value: 'updatedAt' },
  { label: '按发布时间', value: 'publishedAt' },
  { label: '按浏览量', value: 'viewCount' },
  { label: '按下载量', value: 'downloadCount' }
]
const pageSizeItems = [12, 24, 48, 96].map(value => ({ label: `${value}/页`, value }))

function toggleSortDirection() {
  direction.value = direction.value === 'desc' ? 'asc' : 'desc'
}

function typeLabel(value: string) {
  return types.find(item => item.value === value)?.label || value
}

function resourceMeta(resource: ResourceView) {
  return [
    resource.createdAt ? `创建 ${abs(resource.createdAt)}` : '',
    resource.updatedAt ? `更新 ${abs(resource.updatedAt)}` : ''
  ].filter(Boolean).join(' · ')
}

const selectionResetKey = computed(() => manageCollectionQueryFingerprint(
  serializeManageCollectionQuery(collectionState.value, collectionDefinition)
))
const {
  selectedIds,
  selectionCount,
  isPageSelected,
  isPageIndeterminate,
  isSelected,
  toggleOne,
  togglePage,
  replace: replaceSelection,
  clear: clearSelection
} = useManageSelection({
  visibleIds: computed(() => resources.value.map(resource => resource.id)),
  filteredTotal: total,
  resetKey: selectionResetKey
})

type BatchFailure = { id: string, code: string, message: string }
type BatchResult = { changed: number, failures: BatchFailure[], interrupted?: boolean, message?: string }
const batchItems = [
  { label: '发布', value: 'publish' },
  { label: '转为草稿', value: 'draft' },
  { label: '归档', value: 'archive' }
]
const batchAction = ref<string>()
const batchBusy = ref(false)
const batchResult = ref<BatchResult>()

async function applyBatch() {
  if (!batchAction.value || !selectedIds.value.length) return
  const requestIds = [...selectedIds.value]
  batchBusy.value = true
  batchResult.value = undefined
  try {
    const result = await call<BatchResult>('/api/v1/resources/mine/batch', {
      method: 'POST',
      body: { ids: requestIds, action: batchAction.value }
    })
    batchResult.value = result
    const failedIds = result.failures.map(item => item.id)
    if (failedIds.length) replaceSelection(failedIds)
    else clearSelection()
    batchAction.value = undefined
    await refresh()
  } catch (batchError) {
    const apiError = batchError as { data?: { message?: string } }
    replaceSelection(requestIds)
    batchResult.value = {
      changed: 0,
      failures: [],
      interrupted: true,
      message: apiError.data?.message || '批量请求中断，已保留选择，请核对当前状态后重试。'
    }
    await refresh()
  } finally {
    batchBusy.value = false
  }
}

const quickEditTarget = ref<ResourceView>()
const showQuickEdit = ref(false)
function openQuickEdit(resource: ResourceView) {
  quickEditTarget.value = resource
  showQuickEdit.value = true
}
async function onQuickEditSaved(resource: ResourceView) {
  const index = data.value.items.findIndex(item => item.id === resource.id)
  if (index >= 0) data.value.items[index] = resource
  quickEditTarget.value = resource
  await refresh()
}

const showCreate = ref(false)
const form = reactive({ title: '', type: 'software', summary: '' })
const creating = ref(false)
const createError = ref('')

function openCreateModal() {
  createError.value = ''
  showCreate.value = true
}

async function create() {
  if (!form.title.trim() || creating.value) return
  creating.value = true
  createError.value = ''
  try {
    const response = await call<{ resource: ResourceView }>('/api/v1/resources', {
      method: 'POST',
      body: { title: form.title, type: form.type, summary: form.summary }
    })
    showCreate.value = false
    await navigateTo(`/manage/${response.resource.id}`)
  } catch (createFailure) {
    const apiError = createFailure as { data?: { message?: string } }
    createError.value = apiError.data?.message || '创建失败，请检查输入后重试。'
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <ManageHeader title="资源管理">
      <template #subtitle>
        已登录：<span class="text-default">{{ user?.name || user?.email }}</span> · 管理资源、状态、封面和下载文件。
      </template>
      <template #actions>
        <UButton icon="i-tabler-plus" label="新建资源" @click="openCreateModal" />
      </template>
    </ManageHeader>

    <div class="space-y-6" :inert="batchBusy" :aria-busy="batchBusy">
      <ManageLifecycleTabs v-model="status" :items="statusTabs" />

      <ManageCollectionToolbar v-model:search="searchInput" search-placeholder="搜索标题、摘要或描述…">
        <template #filters>
          <USelect v-model="sort" :items="sortItems" value-key="value" icon="i-tabler-arrows-sort" size="sm" />
          <UButton
            :icon="direction === 'desc' ? 'i-tabler-sort-descending' : 'i-tabler-sort-ascending'"
            :label="direction === 'desc' ? '降序' : '升序'"
            color="neutral"
            variant="outline"
            size="sm"
            @click="toggleSortDirection"
          />
        </template>
        <template #actions>
          <ManageViewToggle v-model="viewMode" :items="[
            { key: 'list', label: '列表', icon: 'i-tabler-list' },
            { key: 'grid', label: '网格', icon: 'i-tabler-layout-grid' }
          ]" />
        </template>
      </ManageCollectionToolbar>

      <UAlert v-if="error && !resources.length" color="error" icon="i-tabler-alert-circle" title="资源加载失败" description="无法读取资源列表，请检查服务状态后重试。">
        <template #actions>
          <UButton label="重试" color="error" variant="soft" size="sm" @click="() => refresh()" />
        </template>
      </UAlert>

      <SkeletonList v-else-if="showSkeleton" :rows="6" />
      <ManageEmpty
        v-else-if="!resources.length"
        icon="i-tabler-package-off"
        :text="q ? '没有匹配的资源' : '这个状态下还没有资源'"
      />

      <div v-else-if="viewMode === 'list'" class="overflow-hidden rounded-lg border border-default bg-default">
        <ManageRowShell
          v-for="resource in resources"
          :key="resource.id"
          :selected="isSelected(resource.id)"
          :selection-disabled="batchBusy"
          :selection-label="`选择资源：${resource.title}`"
          @select="toggleOne(resource.id)"
        >
          <template #media>
            <div class="size-16 overflow-hidden rounded-md border border-default bg-elevated">
              <img v-if="resource.coverUrl" :src="resource.coverUrl" :alt="resource.title" class="size-full object-cover">
              <div v-else class="grid size-full place-items-center bg-primary/10 text-primary">
                <UIcon name="i-tabler-package" class="size-6" />
              </div>
            </div>
          </template>

          <div class="min-w-0">
            <NuxtLink :to="`/manage/${resource.id}`" class="block truncate text-sm font-semibold text-highlighted hover:text-primary">
              {{ resource.title }}
            </NuxtLink>
            <p class="mt-0.5 truncate text-xs text-muted">{{ resource.summary || '未填写摘要' }}</p>
            <div class="mt-1.5 flex min-h-5 flex-wrap gap-1">
              <UBadge :label="typeLabel(resource.type)" color="primary" variant="soft" size="sm" />
              <ManageTaxonomyChips :items="resource.tags.map(tag => ({ key: tag, label: tag, kind: 'tag' }))" />
            </div>
            <p v-if="resource.issueCount" class="mt-1 inline-flex items-center gap-1 text-xs text-warning">
              <UIcon name="i-tabler-alert-circle" class="size-3.5" />待完善
            </p>
          </div>

          <template #meta>
            <div class="min-w-0 text-xs md:w-44 md:text-right">
              <p class="text-sm font-medium text-highlighted">{{ resource.downloadCount }} 次下载</p>
              <p class="mt-0.5 truncate font-mono text-muted">/{{ resource.slug }}</p>
              <ClientOnly>
                <p class="mt-1 truncate text-dimmed">{{ resourceMeta(resource) }}</p>
                <template #fallback><p class="mt-1 text-dimmed">…</p></template>
              </ClientOnly>
            </div>
          </template>

          <template #actions>
            <UTooltip text="快速编辑">
              <UButton icon="i-tabler-pencil" color="neutral" variant="ghost" size="sm" square :aria-label="`快速编辑：${resource.title}`" @click="openQuickEdit(resource)" />
            </UTooltip>
            <UTooltip text="完整编辑">
              <UButton :to="`/manage/${resource.id}`" icon="i-tabler-file-pencil" color="neutral" variant="ghost" size="sm" square :aria-label="`完整编辑：${resource.title}`" />
            </UTooltip>
          </template>
        </ManageRowShell>
      </div>

      <div v-else class="grid gap-3 [grid-template-columns:repeat(auto-fill,minmax(min(13rem,100%),1fr))]">
        <div
          v-for="resource in resources"
          :key="resource.id"
          class="group relative flex flex-col overflow-hidden rounded-lg border border-default bg-default transition hover:-translate-y-0.5 hover:shadow-sm focus-within:ring-2 focus-within:ring-primary"
          :class="isSelected(resource.id) ? 'ring-2 ring-primary' : ''"
        >
          <NuxtLink :to="`/manage/${resource.id}`" class="absolute inset-0 z-0 rounded-lg" :aria-label="`查看资源：${resource.title}`">
            <span class="sr-only">查看资源：{{ resource.title }}</span>
          </NuxtLink>
          <div class="pointer-events-none relative z-10">
            <div class="h-[clamp(7.5rem,13vw,9.5rem)] overflow-hidden border-b border-default bg-elevated">
              <img v-if="resource.coverUrl" :src="resource.coverUrl" :alt="resource.title" class="size-full object-cover transition duration-300 group-hover:scale-105">
              <div v-else class="grid size-full place-items-center bg-primary/10 text-primary">
                <UIcon name="i-tabler-package" class="size-8" />
              </div>
            </div>
            <UCheckbox
              class="pointer-events-auto absolute left-2 top-2 rounded-md bg-default/85 p-1 backdrop-blur"
              :model-value="isSelected(resource.id)"
              :disabled="batchBusy"
              :aria-label="`选择资源：${resource.title}`"
              @click.stop
              @update:model-value="toggleOne(resource.id)"
            />
            <UTooltip text="快速编辑">
              <UButton
                icon="i-tabler-pencil"
                color="neutral"
                variant="solid"
                size="xs"
                square
                class="pointer-events-auto absolute right-2 top-2"
                :aria-label="`快速编辑：${resource.title}`"
                @click.stop="openQuickEdit(resource)"
              />
            </UTooltip>
          </div>
          <div class="pointer-events-none relative z-10 flex min-w-0 flex-1 flex-col p-3">
            <h2 class="truncate text-sm font-semibold text-highlighted">{{ resource.title }}</h2>
            <p class="mt-1 truncate text-xs text-muted">{{ resource.summary || '未填写摘要' }}</p>
            <p v-if="resource.issueCount" class="mt-1 inline-flex items-center gap-1 text-xs text-warning">
              <UIcon name="i-tabler-alert-circle" class="size-3.5" />待完善
            </p>
            <div class="mt-2 flex min-h-5 flex-wrap gap-1">
              <UBadge :label="typeLabel(resource.type)" color="primary" variant="soft" size="sm" />
              <ManageTaxonomyChips :items="resource.tags.map(tag => ({ key: tag, label: tag, kind: 'tag' }))" />
            </div>
            <div class="mt-auto flex items-end justify-between gap-2 border-t border-default pt-2.5">
              <div>
                <p class="text-xs text-muted">下载</p>
                <p class="text-sm font-semibold text-highlighted">{{ resource.downloadCount }}</p>
              </div>
              <ClientOnly>
                <p class="text-xs text-dimmed">{{ resource.updatedAt ? abs(resource.updatedAt) : '-' }}</p>
                <template #fallback><span class="text-xs text-dimmed">…</span></template>
              </ClientOnly>
            </div>
          </div>
        </div>
      </div>

      <ManageCollectionDock v-if="total > 0 || resources.length" label="资源批量操作与分页">
        <template #selection>
          <ManagePageSelection :model-value="isPageSelected" :indeterminate="isPageIndeterminate" :disabled="batchBusy" label="选择当前页资源" @update:model-value="togglePage" />
          <div v-if="batchResult" class="flex min-w-0 flex-wrap items-center gap-2 rounded-lg bg-elevated px-2.5 py-1.5">
            <UIcon :name="batchResult.interrupted || batchResult.failures.length ? 'i-tabler-alert-triangle' : 'i-tabler-circle-check'" :class="batchResult.interrupted || batchResult.failures.length ? 'text-warning' : 'text-success'" />
            <span class="text-xs text-default">
              <template v-if="batchResult.interrupted">{{ batchResult.message }}</template>
              <template v-else>已处理 {{ batchResult.changed }} 个<span v-if="batchResult.failures.length">，{{ batchResult.failures.length }} 个未完成</span></template>
            </span>
            <UButton v-if="batchResult.failures[0]" :to="`/manage/${batchResult.failures[0].id}`" label="查看首个失败项" color="warning" variant="link" size="xs" />
            <UButton icon="i-tabler-x" color="neutral" variant="ghost" size="xs" square aria-label="关闭批量结果" @click="batchResult = undefined" />
          </div>
          <template v-if="selectionCount">
            <span class="text-sm text-default">已选 {{ selectionCount }}</span>
            <span class="h-4 w-px bg-default" />
            <USelect v-model="batchAction" :items="batchItems" value-key="value" placeholder="批量操作" size="sm" class="w-28" :disabled="batchBusy" />
            <UButton size="sm" color="primary" variant="soft" :disabled="!batchAction" :loading="batchBusy" @click="applyBatch">应用</UButton>
            <UButton size="sm" color="neutral" variant="ghost" :disabled="batchBusy" @click="clearSelection">取消</UButton>
          </template>
          <span v-else class="text-xs">共 {{ total }} 个资源</span>
        </template>
        <template #pagination>
          <USelect v-model="size" :items="pageSizeItems" value-key="value" size="sm" class="w-20" :disabled="batchBusy" />
          <ManagePagination v-model="page" :total-pages="totalPages" class="!mt-0" />
        </template>
      </ManageCollectionDock>
    </div>

    <ResourceQuickEditModal v-model:open="showQuickEdit" :resource="quickEditTarget" :types="types" @saved="onQuickEditSaved" />

    <UModal v-model:open="showCreate" title="新建资源（草稿）" :ui="{ footer: 'justify-end' }">
      <template #body>
        <div class="space-y-4">
          <UAlert v-if="createError" title="暂时无法创建" :description="createError" icon="i-tabler-alert-circle" color="error" variant="soft" />
          <UFormField label="标题" required>
            <UInput v-model="form.title" placeholder="例如：My CLI Tool" class="w-full" autofocus />
          </UFormField>
          <UFormField label="类型">
            <USelect v-model="form.type" :items="types" value-key="value" class="w-full" />
          </UFormField>
          <UFormField label="简介">
            <UTextarea v-model="form.summary" :rows="3" class="w-full" placeholder="一句话介绍这个资源" />
          </UFormField>
        </div>
      </template>
      <template #footer="{ close }">
        <UButton label="取消" color="neutral" variant="outline" :disabled="creating" @click="close" />
        <UButton label="创建并编辑" icon="i-tabler-arrow-right" trailing :loading="creating" :disabled="!form.title.trim()" @click="create" />
      </template>
    </UModal>
  </div>
</template>
