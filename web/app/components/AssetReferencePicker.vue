<script setup lang="ts">
import { PlatformImageCropper } from '@platform/asset/components'
import type { AssetView } from '~/types'
import { assetExtension, assetFileIcon, assetFileTone, formatAssetSize } from '~/utils/asset-display.mjs'

const props = withDefaults(defineProps<{
  modelValue?: string
  label: string
  profileKey?: string
  imageOnly?: boolean
  visibility?: 'public' | 'private'
  uploadEnabled?: boolean
  cropAspect?: number
  cropOutputWidth?: number
  cropOutputHeight?: number
}>(), {
  modelValue: '',
  profileKey: '',
  imageOnly: false,
  visibility: undefined,
  uploadEnabled: true,
  cropAspect: 3 / 2,
  cropOutputWidth: 1200,
  cropOutputHeight: 800
})

const emit = defineEmits<{ 'update:modelValue': [value: string], selected: [asset: AssetView] }>()

const open = ref(false)
const q = ref('')
const selected = ref<AssetView | null>(null)
const { call } = useAssetApi()
const { slug: siteSlug } = useSiteRuntime()
const { uploadAsset } = useAssetUpload()
const toast = useToast()
const fileInput = ref<HTMLInputElement>()
const uploading = ref(false)
const uploadPct = ref(0)
const cropOpen = ref(false)
const cropFile = ref<File | null>(null)

const { data, pending, refresh } = await useAsyncData(
  () => `resource-assets-${props.profileKey}-${props.imageOnly}`,
  () => call<{ items: AssetView[], total: number }>('/api/v1/assets', {
    query: {
      siteKey: siteSlug.value,
      profileKey: props.profileKey || undefined,
      mime: props.imageOnly ? 'image/' : undefined,
      page: 1,
      size: 48
    }
  }),
  { server: false, immediate: false, default: () => ({ items: [], total: 0 }) }
)

const assets = computed(() => data.value?.items ?? [])
const filtered = computed(() => {
  const kw = q.value.trim().toLowerCase()
  if (!kw) return assets.value
  return assets.value.filter(asset =>
    asset.filename.toLowerCase().includes(kw)
    || asset.title.toLowerCase().includes(kw)
    || asset.id.toLowerCase().includes(kw)
  )
})

watch(open, async (value) => {
  if (value) await refresh()
})

watch([() => props.modelValue, assets], () => {
  selected.value = assets.value.find(asset => asset.id === props.modelValue) || selected.value
}, { immediate: true })

function pick(asset: AssetView) {
  selected.value = asset
  emit('update:modelValue', asset.id)
  emit('selected', asset)
  open.value = false
}

function clear() {
  selected.value = null
  emit('update:modelValue', '')
}

function pickFile() {
  fileInput.value?.click()
}

async function onFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  if (props.imageOnly && !file.type.startsWith('image/')) {
    toast.add({ title: '请选择图片文件', color: 'warning', icon: 'i-tabler-alert-circle' })
    return
  }
  if (props.imageOnly) {
    cropFile.value = file
    open.value = false
    cropOpen.value = true
    return
  }
  await uploadChosenFile(file)
}

async function uploadChosenFile(file: File) {
  uploading.value = true
  uploadPct.value = 1
  try {
    const asset = await uploadAsset(file, {
      profileKey: props.profileKey || 'resource-content',
      visibility: props.visibility || 'public',
      optimizeImage: props.imageOnly,
      imageMaxSide: props.imageOnly ? Math.max(props.cropOutputWidth, props.cropOutputHeight) : 1920,
      imageQuality: props.imageOnly ? 0.9 : 0.86,
      onProgress: (pct) => { uploadPct.value = pct }
    })
    selected.value = asset
    emit('update:modelValue', asset.id)
    emit('selected', asset)
    toast.add({ title: `${props.label}已上传`, color: 'success', icon: 'i-tabler-check' })
    await refresh()
  } catch (err) {
    toast.add({ title: '上传失败', description: (err as Error).message, color: 'error', icon: 'i-tabler-alert-circle' })
  } finally {
    uploading.value = false
  }
}

async function onCropped(payload: { file: File }) {
  cropOpen.value = false
  cropFile.value = null
  await uploadChosenFile(payload.file)
}

function cancelCrop() {
  cropOpen.value = false
  cropFile.value = null
}

function onCropError(message: string) {
  toast.add({ title: '裁剪失败', description: message, color: 'error', icon: 'i-tabler-alert-circle' })
}

watch(cropOpen, (open) => {
  if (!open) cropFile.value = null
})

async function loadAssets() {
  try {
    await refresh()
  } catch (err) {
    toast.add({ title: '素材库加载失败', description: (err as Error).message, color: 'error', icon: 'i-tabler-alert-circle' })
  }
}

const selectedLabel = computed(() => selected.value?.filename || (props.modelValue ? '已选择素材' : '未选择'))
const selectedMeta = computed(() => selected.value ? assetMeta(selected.value) : (props.modelValue ? '来自素材库' : '从素材库选择或上传'))
const canUsePublicImage = (asset?: AssetView | null) => Boolean(asset?.mime.startsWith('image/') && asset.visibility === 'public')
const uploadAccept = computed(() => props.imageOnly ? 'image/*' : undefined)

function assetMeta(asset: AssetView) {
  const ext = assetExtension(asset.filename)
  return [
    ext ? ext.toUpperCase() : asset.mime || '文件',
    formatAssetSize(asset.size),
    asset.visibility === 'private' ? '签名交付' : '公开'
  ].filter(Boolean).join(' · ')
}
</script>

<template>
  <div class="space-y-2">
    <div class="flex items-center justify-between gap-2">
      <p class="text-sm font-medium text-highlighted">{{ label }}</p>
      <div class="flex items-center gap-1">
        <UButton v-if="uploadEnabled" size="xs" color="neutral" variant="ghost" icon="i-tabler-upload" label="上传" :loading="uploading" @click="pickFile" />
        <UButton v-if="modelValue" size="xs" color="neutral" variant="ghost" icon="i-tabler-x" label="清除" @click="clear" />
      </div>
    </div>

    <button
      type="button"
      class="resource-surface-card resource-surface-interactive flex w-full items-center gap-3 overflow-hidden rounded-lg p-3 text-left"
      @click="open = true; loadAssets()"
    >
      <div
        class="grid size-12 shrink-0 place-items-center overflow-hidden rounded-md"
        :class="selected ? assetFileTone(selected.filename, selected.mime) : 'bg-[var(--resource-surface-inset)] text-muted'"
      >
        <img
          v-if="modelValue && imageOnly && (!selected || canUsePublicImage(selected))"
          :src="`/asset-api/api/v1/assets/${encodeURIComponent(modelValue)}/image/@600x400_mode=fit_type=webp_q=72.webp`"
          :alt="selectedLabel"
          class="size-full object-cover"
        >
        <UIcon v-else :name="selected ? assetFileIcon(selected.filename, selected.mime) : imageOnly ? 'i-tabler-photo' : 'i-tabler-package'" class="size-5" />
      </div>
      <div class="min-w-0 flex-1">
        <p class="truncate text-sm font-medium text-highlighted">{{ selectedLabel }}</p>
        <p class="mt-1 truncate text-xs text-muted">{{ selectedMeta }}</p>
      </div>
      <UIcon name="i-tabler-chevron-right" class="size-4 shrink-0 text-muted" />
    </button>
    <div v-if="uploading" class="h-1.5 overflow-hidden rounded-full bg-accented">
      <div class="h-full rounded-full bg-primary transition-all" :style="{ width: `${uploadPct}%` }" />
    </div>

    <input ref="fileInput" type="file" class="hidden" :accept="uploadAccept" @change="onFile">

    <PlatformImageCropper
      v-model:open="cropOpen"
      :file="cropFile"
      title="裁剪资源封面"
      mode="fixed"
      :aspect-ratio="cropAspect"
      :preview-aspect-ratio="cropAspect"
      :output-width="cropOutputWidth"
      :output-height="cropOutputHeight"
      output-type="image/webp"
      :quality="0.9"
      @cropped="onCropped"
      @cancel="cancelCrop"
      @error="onCropError"
    />

    <USlideover v-model:open="open" :title="`选择${label}`">
      <template #body>
        <div class="space-y-4">
          <div class="flex gap-2">
            <UInput v-model="q" icon="i-tabler-search" placeholder="搜索文件名 / 标题" class="min-w-0 flex-1" />
            <UButton v-if="uploadEnabled" icon="i-tabler-upload" color="neutral" variant="outline" label="上传" :loading="uploading" @click="pickFile" />
          </div>
          <div v-if="uploading" class="resource-surface-inset rounded-lg p-3">
            <div class="flex items-center justify-between text-xs text-muted">
              <span>正在上传 {{ label }}</span>
              <span>{{ uploadPct }}%</span>
            </div>
            <div class="mt-2 h-1.5 overflow-hidden rounded-full bg-accented">
              <div class="h-full rounded-full bg-primary transition-all" :style="{ width: `${uploadPct}%` }" />
            </div>
          </div>

          <div v-if="pending" class="grid gap-3 sm:grid-cols-2">
            <USkeleton v-for="i in 6" :key="i" class="h-28 rounded-lg" />
          </div>

          <div v-else-if="!filtered.length" class="resource-surface-card rounded-lg border-dashed py-12 text-center text-sm text-muted">
            素材库暂无可选素材
          </div>

          <div v-else class="grid gap-3 sm:grid-cols-2">
            <button
              v-for="asset in filtered"
              :key="asset.id"
              type="button"
              class="resource-surface-card resource-surface-interactive overflow-hidden rounded-lg text-left"
              @click="pick(asset)"
            >
              <div
                class="grid aspect-[3/2] place-items-center"
                :class="canUsePublicImage(asset) ? 'bg-[var(--resource-surface-inset)]' : assetFileTone(asset.filename, asset.mime)"
              >
                <img
                  v-if="canUsePublicImage(asset)"
                  :src="`/asset-api/api/v1/assets/${encodeURIComponent(asset.id)}/image/@600x400_mode=fit_type=webp_q=72.webp`"
                  :alt="asset.title || asset.filename"
                  class="size-full object-cover"
                  loading="lazy"
                >
                <div v-else class="grid place-items-center gap-2 text-center">
                  <UIcon :name="assetFileIcon(asset.filename, asset.mime)" class="size-9" />
                  <span class="rounded bg-black/5 px-2 py-0.5 text-xs font-semibold uppercase text-current dark:bg-white/10">
                    {{ assetExtension(asset.filename) || 'file' }}
                  </span>
                </div>
              </div>
              <div class="space-y-1 p-3">
                <p class="truncate text-sm font-medium text-highlighted">{{ asset.title || asset.filename }}</p>
                <p class="truncate text-xs text-muted">{{ assetMeta(asset) }}</p>
              </div>
            </button>
          </div>
        </div>
      </template>
    </USlideover>
  </div>
</template>
