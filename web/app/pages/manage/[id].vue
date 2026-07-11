<script setup lang="ts">
import { ActionFeedbackButton } from '@platform/manage/components'
import { useActionFeedback } from '@platform/manage/use-action-feedback'
import type { AssetView, DeliveryItemView, ListTaxonomies, ResourceAssetView, ResourceDetail, TaxonomyView } from '~/types'

definePageMeta({ layout: 'manage', middleware: 'auth' })

const route = useRoute()
const id = route.params.id as string
const { call } = useApi()
const { uploadFile } = useUpload()
const { uploadPublicImage } = useAssetUpload()
const toast = useToast()

const { data, pending, refresh } = await useAsyncData(
  `manage-${id}`,
  () => call<ResourceDetail>(`/api/v1/resources/${id}`),
  { server: false }
)

const r = computed(() => data.value?.resource)
const assets = computed<ResourceAssetView[]>(() => data.value?.assets || [])

const mounted = ref(false)
onMounted(() => { mounted.value = true })

const types = [
  { label: '软件 / 工具', value: 'software' },
  { label: '设计素材 / 模板', value: 'design' },
  { label: '脚本', value: 'script' },
  { label: '其它', value: 'default' }
]
const statusMeta: Record<string, { label: string, color: 'neutral' | 'success' | 'warning', icon: string }> = {
  draft: { label: '草稿', color: 'neutral', icon: 'i-tabler-pencil' },
  published: { label: '已发布', color: 'success', icon: 'i-tabler-circle-check' },
  archived: { label: '已归档', color: 'warning', icon: 'i-tabler-archive' }
}
const deliveryKindItems = [
  { label: '文件', value: 'asset_file' },
  { label: '网盘', value: 'netdisk' }
]

const form = reactive({
  slug: '',
  title: '',
  summary: '',
  description: '',
  type: 'software',
  tagsText: '',
  coverAssetId: '',
  coverUrl: '',
  status: 'draft',
  publishedAt: '',
  viewCount: 0,
  downloadCount: 0,
  deliveryItems: [] as DeliveryItemView[]
})

watch(data, (d) => {
  if (!d?.resource) return
  const x = d.resource
  form.slug = x.slug
  form.title = x.title
  form.summary = x.summary
  form.description = x.description
  form.type = x.type
  form.tagsText = (x.tags || []).join(', ')
  form.coverAssetId = x.coverAssetId || ''
  form.coverUrl = x.coverUrl || ''
  form.status = x.status || 'draft'
  form.publishedAt = toDateTimeInput(x.publishedAt)
  form.viewCount = x.viewCount || 0
  form.downloadCount = x.downloadCount || 0
  form.deliveryItems = normalizeDeliveryItems(x.deliveryPayload?.items || [], d.assets || [])
}, { immediate: true })

watch(() => form.deliveryItems.map(item => item.kind).join('|'), () => {
  for (const item of form.deliveryItems) {
    if (item.kind === 'netdisk' && !item.netdisk) item.netdisk = {}
    if (item.kind === 'asset_file') item.netdisk = undefined
  }
})

watch(() => form.coverAssetId, (assetId) => {
  if (!assetId) form.coverUrl = ''
})

const sm = computed(() => statusMeta[form.status || 'draft'] || statusMeta.draft!)
const title = computed(() => form.title || '未命名资源')
const previewTo = computed(() => `/resources/${id}`)
const assetItems = computed(() => assets.value.map(asset => ({
  label: asset.label || asset.filename,
  value: asset.assetId,
  description: `${asset.filename} · ${fmtSize(asset.size)}`
})))
const enabledDeliveryCount = computed(() => form.deliveryItems.filter(item => item.enabled !== false && deliveryItemReady(item)).length)
const readiness = computed(() => {
  const missing: { key: string, label: string, description: string, icon: string }[] = []
  if (!form.coverUrl) missing.push({ key: 'cover', label: '缺少封面', description: '列表和详情页会缺少主要视觉信息。', icon: 'i-tabler-photo' })
  if (!enabledDeliveryCount.value) missing.push({ key: 'delivery', label: '缺少交付配置', description: '需要至少一个文件或网盘交付项。', icon: 'i-tabler-package' })
  if (!form.description.trim()) missing.push({ key: 'description', label: '缺少资源详情', description: '用户无法了解使用方式、授权范围和注意事项。', icon: 'i-tabler-file-text' })
  return { complete: missing.length === 0, missing }
})
const lifecycleItems = [
  { label: '草稿', value: 'draft' },
  { label: '已发布', value: 'published' },
  { label: '已归档', value: 'archived' }
]

function normalizeDeliveryItems(items: DeliveryItemView[], currentAssets: ResourceAssetView[]): DeliveryItemView[] {
  if (items.length) {
    return items.map((item, index) => ({
      id: item.id || newItemId(),
      kind: item.kind || 'asset_file',
      title: item.title || (item.kind === 'netdisk' ? '网盘交付' : '下载文件'),
      assetId: item.assetId || '',
      netdisk: item.kind === 'netdisk' ? { ...(item.netdisk || {}) } : undefined,
      sort: item.sort ?? index,
      enabled: item.enabled !== false,
      required: item.required !== false
    }))
  }
  return currentAssets.map((asset, index) => ({
    id: newItemId(),
    kind: 'asset_file',
    title: asset.label || asset.filename || '下载文件',
    assetId: asset.assetId,
    sort: index,
    enabled: true,
    required: true
  }))
}

function newItemId() {
  return `delivery-${Math.random().toString(36).slice(2, 8)}`
}

function toSlug(value: string) {
  return value.toLowerCase().trim().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '')
}

function toDateTimeInput(value?: string | null) {
  if (!value) return ''
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return ''
  const local = new Date(d.getTime() - d.getTimezoneOffset() * 60000)
  return local.toISOString().slice(0, 16)
}

function fromDateTimeInput(value: string) {
  if (!value) return ''
  const d = new Date(value)
  return Number.isNaN(d.getTime()) ? '' : d.toISOString()
}

function deliveryItemReady(item: DeliveryItemView) {
  if (item.kind === 'asset_file') return !!item.assetId
  return !!item.netdisk?.url
}

function deliveryPayload() {
  return {
    items: form.deliveryItems
      .filter(item => item.enabled !== false)
      .map((item, index) => ({
        id: item.id,
        kind: item.kind,
        title: item.title || (item.kind === 'netdisk' ? '网盘交付' : '下载文件'),
        assetId: item.kind === 'asset_file' ? (item.assetId || '') : '',
        netdisk: item.kind === 'netdisk' ? (item.netdisk || {}) : undefined,
        sort: index,
        enabled: true,
        required: item.required !== false
      }))
  }
}

function buildSaveBody(extra: Record<string, unknown> = {}) {
  const tags = form.tagsText.split(',').map(s => s.trim()).filter(Boolean)
  const hasFile = form.deliveryItems.some(item => item.enabled !== false && item.kind === 'asset_file' && item.assetId)
  const hasNetdisk = form.deliveryItems.some(item => item.enabled !== false && item.kind === 'netdisk' && item.netdisk?.url)
  return {
    slug: form.slug,
    title: form.title,
    summary: form.summary,
    description: form.description,
    type: form.type,
    coverAssetId: form.coverAssetId,
    coverUrl: form.coverUrl,
    tags,
    status: form.status,
    publishedAt: fromDateTimeInput(form.publishedAt),
    viewCount: Number(form.viewCount || 0),
    downloadCount: Number(form.downloadCount || 0),
    deliveryKind: hasFile && hasNetdisk ? 'bundle' : hasNetdisk ? 'netdisk' : 'asset_file',
    deliveryPayload: deliveryPayload(),
    ...extra
  }
}

const { status: saveStatus, pending: markSaving, success: markSaved, reset: resetSave } = useActionFeedback()
async function save() {
  const slug = toSlug(form.slug)
  if (!slug) {
    toast.add({ title: 'Slug 不能为空', color: 'warning', icon: 'i-tabler-alert-circle' })
    return
  }
  form.slug = slug
  markSaving()
  try {
    await call(`/api/v1/resources/${id}`, { method: 'PATCH', body: buildSaveBody() })
    await saveTaxonomies(false)
    markSaved()
    await refresh()
  } catch (e: any) {
    resetSave()
    toast.add({ title: '保存失败', description: e?.data?.message || e?.message || '请重试', color: 'error' })
  }
}

const { data: allTaxData, refresh: refreshTax } = await useAsyncData(
  `all-tax-${id}`,
  () => call<ListTaxonomies>('/api/v1/taxonomies'),
  { server: false, default: () => ({ items: [] as TaxonomyView[] }) }
)
const allCategories = computed(() => (allTaxData.value?.items || []).filter(t => t.taxonomy === 'category'))
const allTags = computed(() => (allTaxData.value?.items || []).filter(t => t.taxonomy === 'tag'))
const selected = ref<Set<string>>(new Set())
watch(data, (d) => {
  if (d?.taxonomies) selected.value = new Set(d.taxonomies.map(t => t.id))
}, { immediate: true })

const categoryItems = computed(() => {
  const byParent = new Map<string, TaxonomyView[]>()
  for (const c of allCategories.value) {
    const k = c.parentId || ''
    if (!byParent.has(k)) byParent.set(k, [])
    byParent.get(k)!.push(c)
  }
  const out: { label: string, value: string }[] = []
  const walk = (parent: string, depth: number) => {
    for (const c of byParent.get(parent) || []) {
      out.push({ label: `${'  '.repeat(depth)}${c.name}`, value: c.id })
      walk(c.id, depth + 1)
    }
  }
  walk('', 0)
  return out
})
const tagItems = computed(() => allTags.value.map(t => ({ label: t.name, value: t.id })))
const selectedCategoryIds = computed<string[]>({
  get: () => allCategories.value.filter(t => selected.value.has(t.id)).map(t => t.id),
  set: (ids) => {
    const s = new Set(selected.value)
    for (const c of allCategories.value) s.delete(c.id)
    for (const id of ids) s.add(id)
    selected.value = s
  }
})
const selectedTagIds = computed<string[]>({
  get: () => allTags.value.filter(t => selected.value.has(t.id)).map(t => t.id),
  set: (ids) => {
    const s = new Set(selected.value)
    for (const t of allTags.value) s.delete(t.id)
    for (const id of ids) s.add(id)
    selected.value = s
  }
})

const { pending: markTaxSaving, success: markTaxSaved, reset: resetTaxSave } = useActionFeedback()
async function saveTaxonomies(withFeedback = true) {
  if (withFeedback) markTaxSaving()
  try {
    await call(`/api/v1/resources/${id}/taxonomies`, { method: 'PUT', body: { taxonomyIds: [...selected.value] } })
    if (withFeedback) {
      markTaxSaved()
      await refresh()
    }
  } catch (e: any) {
    if (withFeedback) {
      resetTaxSave()
      toast.add({ title: '保存归类失败', description: e?.data?.message || '请重试', color: 'error' })
    }
    else {
      throw e
    }
  }
}

const taxonomyOpen = ref(false)
const taxonomySaving = ref(false)
const taxonomyKind = ref<'category' | 'tag'>('category')
const taxonomyForm = reactive({ name: '', slug: '' })
function openTaxonomy(kind: 'category' | 'tag') {
  taxonomyKind.value = kind
  taxonomyForm.name = ''
  taxonomyForm.slug = ''
  taxonomyOpen.value = true
}
watch(() => taxonomyForm.name, (name) => {
  if (!taxonomyOpen.value || taxonomyForm.slug) return
  taxonomyForm.slug = toSlug(name)
})
async function createTaxonomy() {
  const name = taxonomyForm.name.trim()
  const slug = toSlug(taxonomyForm.slug || name)
  if (!name || !slug) {
    toast.add({ title: '请填写名称和 Slug', color: 'warning', icon: 'i-tabler-alert-circle' })
    return
  }
  taxonomySaving.value = true
  try {
    const res = await call<{ taxonomy: TaxonomyView }>('/api/v1/taxonomies', {
      method: 'POST',
      body: { name, slug, taxonomy: taxonomyKind.value }
    })
    await refreshTax()
    if (taxonomyKind.value === 'category') selectedCategoryIds.value = [...new Set([...selectedCategoryIds.value, res.taxonomy.id])]
    else selectedTagIds.value = [...new Set([...selectedTagIds.value, res.taxonomy.id])]
    taxonomyOpen.value = false
  } catch (e: any) {
    toast.add({ title: '创建失败', description: e?.data?.message || '请重试', color: 'error' })
  } finally {
    taxonomySaving.value = false
  }
}

interface Job { name: string, pct: number, status: 'uploading' | 'done' | 'error', error?: string }
const jobs = ref<Job[]>([])
const activeJobs = computed(() => jobs.value.filter(j => j.status === 'uploading' || j.status === 'error'))
const fileInput = ref<HTMLInputElement>()
const uploadTargetIndex = ref<number | null>(null)

function pickFileForItem(index: number) {
  uploadTargetIndex.value = index
  fileInput.value?.click()
}

async function onPickFiles(e: Event) {
  const input = e.target as HTMLInputElement
  const files = Array.from(input.files || [])
  input.value = ''
  for (const file of files) {
    const job = reactive<Job>({ name: file.name, pct: 0, status: 'uploading' })
    jobs.value.push(job)
    try {
      const asset = await uploadFile(id, file, '', (pct) => { job.pct = pct })
      job.status = 'done'
      job.pct = 100
      if (uploadTargetIndex.value != null && form.deliveryItems[uploadTargetIndex.value]) {
        const item = form.deliveryItems[uploadTargetIndex.value]!
        item.assetId = asset.assetId
        if (!item.title || item.title === '下载文件') item.title = asset.label || asset.filename
      }
      else {
        form.deliveryItems.push({
          id: newItemId(),
          kind: 'asset_file',
          title: asset.label || asset.filename || '下载文件',
          assetId: asset.assetId,
          sort: form.deliveryItems.length,
          enabled: true,
          required: true
        })
      }
      uploadTargetIndex.value = null
      await refresh()
    } catch (err: any) {
      job.status = 'error'
      job.error = err?.message || err?.data?.message || '上传失败'
      toast.add({ title: `「${file.name}」上传失败`, description: job.error, color: 'error' })
    }
  }
}

function addDeliveryItem(kind: 'asset_file' | 'netdisk') {
  form.deliveryItems.push({
    id: newItemId(),
    kind,
    title: kind === 'netdisk' ? '网盘交付' : '下载文件',
    assetId: '',
    netdisk: kind === 'netdisk' ? {} : undefined,
    sort: form.deliveryItems.length,
    enabled: true,
    required: true
  })
}

function removeDeliveryItem(index: number) {
  form.deliveryItems.splice(index, 1)
}

function moveDeliveryItem(index: number, offset: -1 | 1) {
  const next = index + offset
  if (next < 0 || next >= form.deliveryItems.length) return
  const [item] = form.deliveryItems.splice(index, 1)
  if (!item) return
  form.deliveryItems.splice(next, 0, item)
}

function coverUrlForAsset(assetId: string) {
  return `/asset-api/api/v1/assets/${encodeURIComponent(assetId)}/image/@1200x800_mode=fit_type=webp_q=72.webp`
}

function onCoverSelected(asset: AssetView) {
  form.coverAssetId = asset.id
  form.coverUrl = coverUrlForAsset(asset.id)
}

async function uploadEditorImage(file: File) {
  const res = await uploadPublicImage(file, undefined, 'resource-content')
  return res.url
}

const showDelete = ref(false)
const deleting = ref(false)
function openDelete() {
  showDelete.value = true
}
async function del() {
  deleting.value = true
  try {
    await call(`/api/v1/resources/${id}`, { method: 'DELETE' })
    toast.add({ title: '已删除资源', color: 'success', icon: 'i-tabler-check' })
    navigateTo('/manage')
  } catch (e: any) {
    toast.add({ title: '删除失败', description: e?.data?.message || '请重试', color: 'error' })
    deleting.value = false
  }
}

function fmtSize(n: number) {
  if (n >= 1 << 30) return (n / (1 << 30)).toFixed(1) + ' GB'
  if (n >= 1 << 20) return (n / (1 << 20)).toFixed(1) + ' MB'
  if (n >= 1 << 10) return (n / (1 << 10)).toFixed(1) + ' KB'
  return n + ' B'
}
</script>

<template>
  <div class="space-y-5 pb-10">
    <div class="resource-topbar sticky top-16 z-10 -mx-5 border-b px-5 py-3 backdrop-blur lg:-mx-10 lg:px-10">
      <div class="mx-auto flex max-w-7xl flex-wrap items-center gap-3">
        <UButton icon="i-tabler-arrow-left" color="neutral" variant="ghost" aria-label="返回" to="/manage" />
        <div class="min-w-0 flex-1">
          <p class="truncate text-sm font-medium text-highlighted">{{ title }}</p>
          <p class="truncate text-xs text-muted">{{ form.slug ? `/resources/${form.slug}` : '资源编辑' }}</p>
        </div>
        <UPopover mode="hover" :open-delay="120" :close-delay="120" :content="{ side: 'bottom', align: 'end', sideOffset: 10 }" arrow>
          <UButton
            :icon="readiness.complete ? 'i-tabler-circle-check' : 'i-tabler-alert-triangle'"
            :label="readiness.complete ? '配置完整' : `${readiness.missing.length} 项待补`"
            :color="readiness.complete ? 'success' : 'warning'"
            variant="soft"
          />
          <template #content>
            <div class="w-72 p-3">
              <p class="text-sm font-medium text-highlighted">{{ readiness.complete ? '发布配置完整' : '发布前建议补齐' }}</p>
              <div class="mt-3 space-y-2">
                <div v-if="readiness.complete" class="flex items-center gap-2 rounded-md border border-default bg-default/60 px-3 py-2 text-sm text-success">
                  <UIcon name="i-tabler-circle-check" class="size-4" />可以正常发布
                </div>
                <template v-else>
                  <div v-for="item in readiness.missing" :key="item.key" class="flex items-start gap-2 rounded-md border border-default bg-default/60 px-3 py-2">
                    <UIcon :name="item.icon" class="mt-0.5 size-4 shrink-0 text-warning" />
                    <div class="min-w-0">
                      <p class="text-sm font-medium text-highlighted">{{ item.label }}</p>
                      <p class="mt-0.5 text-xs leading-5 text-muted">{{ item.description }}</p>
                    </div>
                  </div>
                </template>
              </div>
            </div>
          </template>
        </UPopover>
        <UButton :to="previewTo" target="_blank" icon="i-tabler-eye" color="neutral" variant="outline" label="预览" />
        <ActionFeedbackButton :status="saveStatus" color="neutral" variant="outline" idle-label="保存" pending-label="保存中" success-label="已保存" @click="save" />
        <UTooltip text="删除资源">
          <UButton icon="i-tabler-trash" color="neutral" variant="ghost" aria-label="删除资源" @click="openDelete" />
        </UTooltip>
      </div>
    </div>

    <USkeleton v-if="!mounted || (pending && !r)" class="mx-auto h-[640px] max-w-5xl rounded-lg" />

    <div v-else-if="r" class="mx-auto grid max-w-7xl gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
      <section class="min-w-0 space-y-5">
        <div class="resource-surface-card rounded-lg p-5">
          <div class="mb-4">
            <div>
              <p class="text-sm font-medium text-primary">Basic</p>
              <h1 class="font-display text-xl font-semibold text-highlighted">基本信息</h1>
            </div>
          </div>
          <div class="space-y-4">
            <UFormField label="资源标题" required>
              <UInput v-model="form.title" class="resource-title-input w-full" placeholder="例如 PostgreSQL 备份工具" />
            </UFormField>
            <div class="grid gap-4 md:grid-cols-[minmax(0,1fr)_14rem]">
              <UFormField label="URL Slug" required>
                <UInput v-model="form.slug" placeholder="resource-slug" icon="i-tabler-link" class="w-full" />
              </UFormField>
              <UFormField label="类型">
                <USelectMenu v-model="form.type" :items="types" value-key="value" class="w-full" />
              </UFormField>
            </div>
            <UFormField label="关键词">
              <UInput v-model="form.tagsText" placeholder="cli, windows, 开源" class="w-full" />
            </UFormField>
            <UFormField label="简介">
              <UTextarea v-model="form.summary" :rows="3" class="w-full" />
            </UFormField>
          </div>
        </div>

        <div class="resource-surface-card rounded-lg p-5">
          <div class="mb-4 flex items-start justify-between gap-3">
            <div>
              <p class="text-sm font-medium text-primary">Delivery</p>
              <h2 class="font-display text-xl font-semibold text-highlighted">交付内容</h2>
            </div>
            <UBadge :label="`${form.deliveryItems.length} 项交付`" color="primary" variant="soft" />
          </div>

          <input ref="fileInput" type="file" multiple class="hidden" @change="onPickFiles">

          <div class="space-y-3">
            <article
              v-for="(item, index) in form.deliveryItems"
              :key="item.id || index"
              class="rounded-lg border border-default bg-default/60 p-4"
            >
              <div class="mb-4 flex items-start justify-between gap-3">
                <div class="flex min-w-0 items-center gap-2">
                  <span class="grid size-9 shrink-0 place-items-center rounded-md bg-primary/10 text-primary">
                    <UIcon :name="item.kind === 'netdisk' ? 'i-tabler-cloud' : 'i-tabler-file-download'" class="size-4" />
                  </span>
                  <div class="min-w-0">
                    <p class="truncate text-sm font-medium text-highlighted">{{ item.title || (item.kind === 'netdisk' ? '网盘交付' : '下载文件') }}</p>
                    <p class="text-xs text-muted">{{ item.kind === 'netdisk' ? '网盘链接' : '资源文件' }}</p>
                  </div>
                </div>
                <div class="flex shrink-0 items-center gap-1">
                  <UTooltip text="上移">
                    <UButton icon="i-tabler-arrow-up" size="xs" color="neutral" variant="ghost" :disabled="index === 0" aria-label="上移交付项" @click="moveDeliveryItem(index, -1)" />
                  </UTooltip>
                  <UTooltip text="下移">
                    <UButton icon="i-tabler-arrow-down" size="xs" color="neutral" variant="ghost" :disabled="index === form.deliveryItems.length - 1" aria-label="下移交付项" @click="moveDeliveryItem(index, 1)" />
                  </UTooltip>
                  <UTooltip text="删除">
                    <UButton icon="i-tabler-trash" size="xs" color="error" variant="ghost" aria-label="删除交付项" @click="removeDeliveryItem(index)" />
                  </UTooltip>
                </div>
              </div>

              <div class="space-y-3">
                <div class="grid gap-3 sm:grid-cols-[minmax(0,1fr)_10rem]">
                  <UFormField label="标题">
                    <UInput v-model="item.title" class="w-full" placeholder="例如 安装包 / 源文件 / 网盘镜像" />
                  </UFormField>
                  <UFormField label="类型">
                    <USelectMenu v-model="item.kind" :items="deliveryKindItems" value-key="value" class="w-full" />
                  </UFormField>
                </div>

                <div v-if="item.kind === 'asset_file'" class="space-y-3">
                  <UFormField label="下载文件">
                    <USelectMenu v-model="item.assetId" :items="assetItems" value-key="value" placeholder="选择已上传文件" class="w-full" />
                  </UFormField>
                  <div class="flex flex-wrap items-center gap-2">
                    <UButton icon="i-tabler-upload" color="neutral" variant="outline" label="上传新文件" @click="pickFileForItem(index)" />
                    <p class="text-xs text-muted">文件上传后会自动选中到当前交付项。</p>
                  </div>
                </div>

                <div v-else class="space-y-3">
                  <div class="grid gap-3 sm:grid-cols-2">
                    <UFormField label="网盘类型" required>
                      <UInput v-model="item.netdisk!.provider" class="w-full" placeholder="百度网盘 / 阿里云盘" />
                    </UFormField>
                    <UFormField label="访问码">
                      <UInput v-model="item.netdisk!.accessCode" class="w-full" placeholder="可选" />
                    </UFormField>
                  </div>
                  <UFormField label="网盘地址" required>
                    <UInput v-model="item.netdisk!.url" class="w-full" placeholder="https://..." />
                  </UFormField>
                  <div class="grid gap-3 sm:grid-cols-2">
                    <UFormField label="提取码 / 解压密码">
                      <UInput v-model="item.netdisk!.extractCode" class="w-full" placeholder="可选" />
                    </UFormField>
                    <UFormField label="状态">
                      <USwitch v-model="item.enabled" label="启用" />
                    </UFormField>
                  </div>
                  <UFormField label="备注">
                    <UTextarea v-model="item.netdisk!.note" :rows="2" class="w-full" placeholder="补充下载说明、失效处理方式等" />
                  </UFormField>
                </div>
              </div>
            </article>
          </div>

          <div v-if="activeJobs.length" class="mt-3 space-y-2">
            <div v-for="(j, i) in activeJobs" :key="'job-' + i" class="rounded-lg border border-default p-3">
              <div class="flex items-center gap-2">
                <UIcon :name="j.status === 'error' ? 'i-tabler-alert-triangle' : 'i-tabler-loader-2'" :class="['size-4 shrink-0', j.status === 'error' ? 'text-error' : 'animate-spin text-muted']" />
                <p class="min-w-0 flex-1 truncate text-sm text-default">{{ j.name }}</p>
                <span class="text-xs text-muted">{{ j.status === 'error' ? '失败' : j.pct + '%' }}</span>
              </div>
              <div v-if="j.status === 'uploading'" class="mt-2 h-1.5 overflow-hidden rounded-full bg-accented">
                <div class="h-full rounded-full bg-primary transition-all" :style="{ width: j.pct + '%' }" />
              </div>
            </div>
          </div>

          <div class="mt-3 flex flex-wrap gap-2">
            <UButton icon="i-tabler-file-plus" color="neutral" variant="outline" label="添加文件" @click="addDeliveryItem('asset_file')" />
            <UButton icon="i-tabler-cloud-plus" color="neutral" variant="outline" label="添加网盘" @click="addDeliveryItem('netdisk')" />
          </div>
        </div>

        <div class="resource-surface-card rounded-lg p-5">
          <div class="mb-4">
            <p class="text-sm font-medium text-primary">Content</p>
            <h2 class="font-display text-xl font-semibold text-highlighted">资源详情</h2>
          </div>
          <ContentEditor
            v-model="form.description"
            :image-uploader="uploadEditorImage"
            draft-key-prefix="resource"
            :draft-entity-id="id"
            :has-initial-content="!!form.description"
          />
        </div>
      </section>

      <aside class="space-y-4 lg:sticky lg:top-32 lg:max-h-[calc(100vh-9rem)] lg:overflow-y-auto lg:pr-1 lg:self-start">
        <div class="resource-surface-card rounded-lg p-4">
          <div class="flex items-center justify-between gap-3">
            <p class="text-sm font-medium text-highlighted">发布与归类</p>
            <UBadge :color="sm.color" :icon="sm.icon" :label="sm.label" variant="soft" />
          </div>

          <div class="mt-4 space-y-3">
            <div class="grid gap-3 sm:grid-cols-2">
              <UFormField label="状态">
                <USelectMenu v-model="form.status" :items="lifecycleItems" value-key="value" class="w-full" />
              </UFormField>
              <UFormField label="发布时间">
                <UInput v-model="form.publishedAt" type="datetime-local" class="w-full" />
              </UFormField>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <UFormField label="浏览量">
                <UInput v-model.number="form.viewCount" type="number" min="0" class="w-full" />
              </UFormField>
              <UFormField label="下载次数">
                <UInput v-model.number="form.downloadCount" type="number" min="0" class="w-full" />
              </UFormField>
            </div>

            <div class="space-y-2">
              <div class="flex items-center justify-between gap-3">
                <p class="text-xs font-medium text-muted">分类</p>
                <UButton size="xs" icon="i-tabler-plus" color="neutral" variant="ghost" label="新建" @click="openTaxonomy('category')" />
              </div>
              <USelectMenu v-model="selectedCategoryIds" :items="categoryItems" value-key="value" multiple placeholder="选择分类" class="w-full" />
            </div>

            <div class="space-y-2">
              <div class="flex items-center justify-between gap-3">
                <p class="text-xs font-medium text-muted">标签</p>
                <UButton size="xs" icon="i-tabler-plus" color="neutral" variant="ghost" label="新建" @click="openTaxonomy('tag')" />
              </div>
              <USelectMenu v-model="selectedTagIds" :items="tagItems" value-key="value" multiple placeholder="选择标签" class="w-full" />
            </div>
          </div>
        </div>

        <div class="resource-surface-card rounded-lg p-4">
          <div class="flex items-center justify-between gap-3">
            <p class="text-sm font-medium text-highlighted">封面</p>
            <UBadge :label="form.coverUrl ? '已配置' : '待上传'" :color="form.coverUrl ? 'success' : 'warning'" variant="soft" />
          </div>
          <div class="relative mt-4 h-[clamp(9rem,22vw,13rem)] overflow-hidden rounded-lg border border-default bg-elevated">
            <img v-if="form.coverUrl" :src="form.coverUrl" alt="cover" class="size-full object-cover">
            <div v-else class="grid size-full place-items-center text-muted">
              <UIcon name="i-tabler-photo" class="size-8" />
            </div>
          </div>
          <AssetReferencePicker
            v-model="form.coverAssetId"
            label="资源封面"
            profile-key="resource-cover"
            visibility="public"
            image-only
            :crop-aspect="3 / 2"
            :crop-output-width="1200"
            :crop-output-height="800"
            class="mt-3"
            @selected="onCoverSelected"
          />
        </div>
      </aside>

      <UModal v-model:open="taxonomyOpen" :title="taxonomyKind === 'category' ? '新建分类' : '新建标签'">
        <template #body>
          <div class="space-y-4">
            <UFormField label="名称" required>
              <UInput v-model="taxonomyForm.name" class="w-full" :placeholder="taxonomyKind === 'category' ? '设计素材' : '模板'" />
            </UFormField>
            <UFormField label="Slug" required>
              <UInput v-model="taxonomyForm.slug" class="w-full" placeholder="design-assets" />
            </UFormField>
          </div>
        </template>
        <template #footer>
          <div class="flex w-full justify-end gap-2">
            <UButton label="取消" color="neutral" variant="outline" @click="void (taxonomyOpen = false)" />
            <UButton icon="i-tabler-plus" label="创建" :loading="taxonomySaving" @click="createTaxonomy" />
          </div>
        </template>
      </UModal>

      <UModal v-model:open="showDelete" title="删除资源" :description="`确定删除「${r.title}」?此操作不可撤销,关联文件也会被清理。`" :ui="{ footer: 'justify-end' }">
        <template #footer="{ close }">
          <UButton label="取消" color="neutral" variant="outline" @click="close" />
          <UButton label="删除" icon="i-tabler-trash" color="error" :loading="deleting" @click="del" />
        </template>
      </UModal>
    </div>
  </div>
</template>

<style scoped>
.resource-title-input :deep(input) {
  min-height: 44px;
  font-size: 22px;
  line-height: 1.25;
  font-weight: 650;
  letter-spacing: 0;
}
</style>
