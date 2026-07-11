<script setup lang="ts">
import { ManageEmpty, ManageHeader, ManagePageFooter, ManagePagination, SkeletonList } from '@platform/manage/components'
import type { ListTaxonomies, TaxonomyView } from '~/types'

const { kind } = defineProps<{ kind: 'category' | 'tag' }>()

const { isAdmin } = useAuth()
const { call } = useApi()
const toast = useToast()

const ROOT = '__root__'
const mounted = ref(false)
onMounted(() => { mounted.value = true })

const { data, pending, refresh } = await useAsyncData(
  `resource-taxonomies-${kind}`,
  () => call<ListTaxonomies>('/api/v1/taxonomies'),
  { server: false, default: () => ({ items: [] as TaxonomyView[] }) }
)

const categories = computed(() => (data.value?.items || []).filter(t => t.taxonomy === 'category'))
const items = computed(() => (data.value?.items || []).filter(t => t.taxonomy === kind))
const tree = computed(() => {
  if (kind !== 'category') return []
  const byParent = new Map<string, TaxonomyView[]>()
  for (const item of items.value) {
    const key = item.parentId || ''
    if (!byParent.has(key)) byParent.set(key, [])
    byParent.get(key)!.push(item)
  }
  const out: { tax: TaxonomyView, depth: number }[] = []
  const walk = (parent: string, depth: number) => {
    for (const item of byParent.get(parent) || []) {
      out.push({ tax: item, depth })
      walk(item.id, depth + 1)
    }
  }
  walk('', 0)
  return out
})

const q = ref('')
const page = ref(1)
const pageSize = ref(30)
const sortBy = ref<'count' | 'name' | 'slug'>('count')
const sortOrder = ref<'asc' | 'desc'>('desc')
const filteredItems = computed(() => {
  const keyword = q.value.trim().toLowerCase()
  const source = kind === 'category' ? tree.value.map(row => row.tax) : items.value
  const filtered = keyword
    ? source.filter(item => `${item.name} ${item.slug} ${item.description || ''}`.toLowerCase().includes(keyword))
    : source
  return [...filtered].sort((a, b) => {
    const dir = sortOrder.value === 'asc' ? 1 : -1
    if (sortBy.value === 'name') return a.name.localeCompare(b.name, 'zh-CN') * dir
    if (sortBy.value === 'slug') return a.slug.localeCompare(b.slug) * dir
    return ((a.count || 0) - (b.count || 0) || a.name.localeCompare(b.name, 'zh-CN')) * dir
  })
})
const totalPages = computed(() => Math.max(1, Math.ceil(filteredItems.value.length / pageSize.value)))
const pagedItems = computed(() => filteredItems.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value))
watch([q, sortBy, sortOrder, pageSize], () => { page.value = 1 })
watch(totalPages, (value) => { if (page.value > value) page.value = value })

const sortItems = [
  { label: '按资源数', value: 'count' },
  { label: '按名称', value: 'name' },
  { label: '按 Slug', value: 'slug' }
]
const pageSizeItems = [
  { label: '15/页', value: 15 },
  { label: '30/页', value: 30 },
  { label: '60/页', value: 60 },
  { label: '120/页', value: 120 }
]

const panelOpen = ref(false)
const current = ref<TaxonomyView | null>(null)
const saving = ref(false)
const deleting = ref(false)
const deleteArmed = ref(false)
const mergingBusy = ref(false)
const mergeTarget = ref('')
const form = reactive({ name: '', slug: '', description: '', parentId: ROOT })

const title = computed(() => kind === 'category' ? '分类' : '标签')
const headerTitle = computed(() => kind === 'category' ? '资源分类' : '资源标签')
const headerSubtitle = computed(() => kind === 'category'
  ? '维护资源站目录层级，帮助用户按用途浏览资源。'
  : '维护资源站标签，合并重复词，保持搜索和筛选清晰。')
const parentItems = computed(() => [
  { label: '顶级分类', value: ROOT },
  ...categories.value
    .filter(category => category.id !== current.value?.id)
    .map(category => ({ label: category.name, value: category.id }))
])
const mergeTargets = computed(() => items.value
  .filter(item => item.id !== current.value?.id)
  .map(item => ({ label: item.name, value: item.id })))

function rowDepth(item: TaxonomyView) {
  if (kind !== 'category') return 0
  return tree.value.find(row => row.tax.id === item.id)?.depth || 0
}

function openCreate() {
  current.value = null
  form.name = ''
  form.slug = ''
  form.description = ''
  form.parentId = ROOT
  mergeTarget.value = ''
  deleteArmed.value = false
  panelOpen.value = true
}

function openEdit(item: TaxonomyView) {
  current.value = item
  form.name = item.name
  form.slug = item.slug
  form.description = item.description || ''
  form.parentId = item.parentId || ROOT
  mergeTarget.value = ''
  deleteArmed.value = false
  panelOpen.value = true
}

async function save() {
  if (!form.name.trim()) return
  saving.value = true
  try {
    if (!current.value) {
      const body: Record<string, unknown> = { name: form.name.trim(), taxonomy: kind }
      if (kind === 'category' && form.parentId !== ROOT) body.parentId = form.parentId
      if (form.description.trim()) body.description = form.description.trim()
      await call('/api/v1/taxonomies', { method: 'POST', body })
      toast.add({ title: '已创建', color: 'success', icon: 'i-tabler-check' })
    } else {
      const body: Record<string, unknown> = { name: form.name, slug: form.slug, description: form.description }
      if (kind === 'category') body.parentId = form.parentId === ROOT ? '' : form.parentId
      await call(`/api/v1/taxonomies/${current.value.id}`, { method: 'PATCH', body })
      toast.add({ title: '已更新', color: 'success', icon: 'i-tabler-check' })
    }
    panelOpen.value = false
    await refresh()
  } catch (e: any) {
    toast.add({ title: current.value ? '更新失败' : '创建失败', description: e?.data?.message || '请重试', color: 'error' })
  } finally {
    saving.value = false
  }
}

async function doMerge() {
  if (!current.value || !mergeTarget.value) return
  mergingBusy.value = true
  try {
    await call(`/api/v1/taxonomies/${current.value.id}/merge`, { method: 'POST', body: { targetId: mergeTarget.value } })
    toast.add({ title: '已合并', color: 'success', icon: 'i-tabler-check' })
    panelOpen.value = false
    await refresh()
  } catch (e: any) {
    toast.add({ title: '合并失败', description: e?.data?.message || '请重试', color: 'error' })
  } finally {
    mergingBusy.value = false
  }
}

async function doDelete() {
  if (!current.value) return
  deleting.value = true
  try {
    await call(`/api/v1/taxonomies/${current.value.id}`, { method: 'DELETE' })
    toast.add({ title: `已删除「${current.value.name}」`, color: 'success', icon: 'i-tabler-check' })
    panelOpen.value = false
    await refresh()
  } catch (e: any) {
    toast.add({ title: '删除失败', description: e?.data?.message || '可能仍有子分类或关联资源', color: 'error' })
  } finally {
    deleting.value = false
  }
}

function closePanel() {
  panelOpen.value = false
}

function armDelete() {
  deleteArmed.value = true
}

function cancelDelete() {
  deleteArmed.value = false
}

function toggleSortOrder() {
  sortOrder.value = sortOrder.value === 'desc' ? 'asc' : 'desc'
}
</script>

<template>
  <div class="space-y-6">
    <ManageHeader :title="headerTitle">
      <template #subtitle>{{ headerSubtitle }}</template>
      <template #actions>
        <UButton :icon="kind === 'category' ? 'i-tabler-folder-plus' : 'i-tabler-hash'" :label="`新建${title}`" @click="openCreate" />
      </template>
    </ManageHeader>

    <SkeletonList v-if="!mounted || pending" :rows="8" />

    <UAlert
      v-else-if="!isAdmin"
      color="warning"
      icon="i-tabler-shield-lock"
      title="需要管理员权限"
      :description="`${title}治理会影响全站资源目录，请使用管理员账户操作。`"
    />

    <template v-else>
      <section class="rounded-lg border border-default bg-default p-3">
        <div class="flex flex-wrap items-center gap-2">
          <UInput v-model="q" icon="i-tabler-search" :placeholder="`搜索${title}名称、slug 或描述`" size="sm" class="w-full sm:w-72" />
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
          <USelect v-model="pageSize" :items="pageSizeItems" size="sm" class="w-full sm:w-28" />
        </div>
      </section>

      <section class="overflow-hidden rounded-lg border border-default bg-default">
        <ManageEmpty v-if="!filteredItems.length" :icon="kind === 'category' ? 'i-tabler-folder-off' : 'i-tabler-hash-off'" :text="`没有匹配的${title}`" />
        <div v-else class="divide-y divide-default">
          <button
            v-for="item in pagedItems"
            :key="item.id"
            type="button"
            class="group grid w-full grid-cols-[minmax(0,1fr)_auto] items-center gap-3 px-4 py-3 text-left transition hover:bg-elevated/50"
            @click="openEdit(item)"
          >
            <div class="flex min-w-0 items-center gap-3" :style="{ paddingLeft: rowDepth(item) * 22 + 'px' }">
              <UIcon v-if="kind === 'category' && rowDepth(item) > 0" name="i-tabler-corner-down-right" class="size-4 shrink-0 text-dimmed" />
              <span class="grid size-8 shrink-0 place-items-center rounded-lg bg-elevated text-muted">
                <UIcon :name="kind === 'category' ? 'i-tabler-folder' : 'i-tabler-hash'" class="size-4" />
              </span>
              <div class="min-w-0">
                <p class="line-clamp-1 text-sm font-medium text-highlighted">{{ item.name }}</p>
                <p class="line-clamp-1 font-mono text-xs text-muted">/{{ item.slug }}</p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <UBadge :label="`${item.count} 个`" color="neutral" variant="subtle" />
              <UIcon name="i-tabler-chevron-right" class="size-4 text-dimmed transition group-hover:translate-x-0.5" />
            </div>
          </button>
        </div>
      </section>

      <ManagePageFooter>
        <template #left>
          <span>{{ title }} {{ filteredItems.length }} 个</span>
          <span v-if="pagedItems.length">当前页 {{ pagedItems.length }} 个</span>
        </template>
        <template #right>
          <ManagePagination v-model="page" :total-pages="totalPages" />
        </template>
      </ManagePageFooter>
    </template>

    <USlideover v-model:open="panelOpen" :title="current ? `编辑${title}` : `新建${title}`" side="right">
      <template #body>
        <div class="space-y-5">
          <UFormField label="名称" required>
            <UInput v-model="form.name" class="w-full" :placeholder="kind === 'category' ? '例如:开发工具' : '例如:DevOps'" />
          </UFormField>
          <UFormField v-if="current" label="Slug" help="URL 标识；改动会影响分类/标签链接。">
            <UInput v-model="form.slug" class="w-full" />
          </UFormField>
          <UFormField v-if="kind === 'category'" label="父分类">
            <USelectMenu
              v-model="form.parentId"
              :items="parentItems"
              value-key="value"
              placeholder="选择父分类"
              :search-input="{ placeholder: '搜索分类' }"
              class="w-full"
            />
          </UFormField>
          <UFormField label="描述">
            <UTextarea v-model="form.description" :rows="3" class="w-full" />
          </UFormField>

          <template v-if="current">
            <USeparator />

            <section class="space-y-3">
              <div>
                <h3 class="text-sm font-medium text-highlighted">合并到其它{{ title }}</h3>
                <p class="mt-1 text-xs text-muted">关联资源会改挂到目标项，当前项随后删除。</p>
              </div>
              <div class="flex gap-2">
                <USelectMenu
                  v-model="mergeTarget"
                  :items="mergeTargets"
                  value-key="value"
                  placeholder="选择目标"
                  :search-input="{ placeholder: '搜索目标' }"
                  class="min-w-0 flex-1"
                />
                <UButton icon="i-tabler-arrows-join" label="合并" color="warning" variant="soft" :disabled="!mergeTarget" :loading="mergingBusy" @click="doMerge" />
              </div>
            </section>

            <section class="rounded-lg border border-error/30 bg-error/5 p-3">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <h3 class="text-sm font-medium text-error">删除{{ title }}</h3>
                  <p class="mt-1 text-xs text-muted">删除不可撤销。有关联或子分类时后端会拒绝。</p>
                </div>
                <UButton v-if="!deleteArmed" label="删除" color="error" variant="soft" size="sm" @click="armDelete" />
              </div>
              <div v-if="deleteArmed" class="mt-3 flex justify-end gap-2">
                <UButton label="取消" color="neutral" variant="ghost" size="sm" @click="cancelDelete" />
                <UButton label="确认删除" color="error" size="sm" :loading="deleting" @click="doDelete" />
              </div>
            </section>
          </template>
        </div>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton label="取消" color="neutral" variant="outline" @click="closePanel" />
          <UButton :label="current ? '保存' : '创建'" icon="i-tabler-check" :loading="saving" :disabled="!form.name.trim()" @click="save" />
        </div>
      </template>
    </USlideover>
  </div>
</template>
