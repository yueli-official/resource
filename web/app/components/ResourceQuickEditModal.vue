<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import { z } from 'zod'
import type { ResourceView } from '~/types'

const open = defineModel<boolean>('open', { required: true })
const { resource, types } = defineProps<{
  resource?: ResourceView
  types: readonly { label: string, value: string }[]
}>()
const emit = defineEmits<{ saved: [resource: ResourceView] }>()

const { call } = useApi()
const schema = z.object({
  title: z.string().trim().min(1, '标题不能为空').max(200, '标题不能超过 200 个字符'),
  slug: z.string().trim().min(1, 'Slug 不能为空').max(200, 'Slug 不能超过 200 个字符'),
  summary: z.string().max(500, '摘要不能超过 500 个字符'),
  type: z.string().min(1, '请选择资源类型'),
  status: z.enum(['draft', 'published', 'archived']),
  tagsText: z.string()
})
type Schema = z.output<typeof schema>

const state = reactive<Schema>({ title: '', slug: '', summary: '', type: 'default', status: 'draft', tagsText: '' })
const saving = ref(false)
const submitError = ref('')
const statusItems = [
  { label: '草稿', value: 'draft' },
  { label: '已发布', value: 'published' },
  { label: '归档', value: 'archived' }
]
const typeItems = computed(() => types.map(item => ({ ...item })))

function tagsFromText(value: string) {
  return Array.from(new Set(value.split(',').map(item => item.trim()).filter(Boolean)))
}

const dirty = computed(() => Boolean(resource) && (
  state.title.trim() !== resource!.title
  || state.slug.trim() !== resource!.slug
  || state.summary !== resource!.summary
  || state.type !== resource!.type
  || state.status !== resource!.status
  || tagsFromText(state.tagsText).join(',') !== (resource!.tags ?? []).join(',')
))

function reset() {
  if (!resource) return
  state.title = resource.title
  state.slug = resource.slug
  state.summary = resource.summary
  state.type = resource.type
  state.status = schema.shape.status.safeParse(resource.status).success ? resource.status as Schema['status'] : 'draft'
  state.tagsText = (resource.tags ?? []).join(', ')
  submitError.value = ''
}

function close() {
  open.value = false
}

watch([() => open.value, () => resource?.id], ([isOpen]) => {
  if (isOpen) reset()
}, { immediate: true })

async function save(event: FormSubmitEvent<Schema>) {
  if (!resource || saving.value) return
  saving.value = true
  submitError.value = ''
  try {
    const response = await call<{ resource: ResourceView }>(`/api/v1/resources/${resource.id}`, {
      method: 'PATCH',
      body: {
        title: event.data.title,
        slug: event.data.slug,
        summary: event.data.summary,
        type: event.data.type,
        status: event.data.status,
        tags: tagsFromText(event.data.tagsText)
      }
    })
    emit('saved', response.resource)
    open.value = false
  } catch (error) {
    const apiError = error as { data?: { code?: string, message?: string } }
    if (apiError.data?.code === 'resource.slug_taken') {
      submitError.value = '这个 Slug 已被使用，请换一个。'
    } else if (apiError.data?.code === 'resource.invalid_state' && state.status === 'published') {
      submitError.value = '发布前需要至少一个文件或可用网盘交付，请先打开完整编辑器。'
    } else {
      submitError.value = resourceFailureMessage(error, '保存失败，请检查输入后重试。')
    }
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <UModal
    v-model:open="open"
    title="快速编辑资源"
    description="修改列表中的高频信息；封面、交付文件、详细说明和 SEO 请使用完整编辑器。"
    scrollable
    :ui="{ content: 'sm:max-w-2xl', footer: 'p-0' }"
  >
    <template #body>
      <UForm v-if="resource" id="resource-quick-edit-form" :schema="schema" :state="state" class="space-y-5" @submit="save">
        <UAlert v-if="submitError" title="暂时无法保存" :description="submitError" icon="i-tabler-alert-circle" color="error" variant="soft" />

        <UFormField name="title" label="标题" required>
          <UInput v-model="state.title" class="w-full" placeholder="资源标题" autofocus />
        </UFormField>

        <div class="grid items-start gap-5 sm:grid-cols-[minmax(0,2fr)_minmax(10rem,1fr)]">
          <UFormField name="slug" label="Slug" description="公开地址会使用规范化后的 Slug。" required :ui="{ description: 'min-h-5' }">
            <UInput v-model="state.slug" class="w-full font-mono" icon="i-tabler-link" placeholder="resource-slug" />
          </UFormField>
          <UFormField name="status" label="生命周期" description="控制资源是否公开。" required :ui="{ description: 'min-h-5' }">
            <USelect v-model="state.status" :items="statusItems" value-key="value" class="w-full" />
          </UFormField>
        </div>

        <div class="grid items-start gap-5 sm:grid-cols-2">
          <UFormField name="type" label="资源类型" description="用于目录和类型筛选。" :ui="{ description: 'min-h-5' }">
            <USelect v-model="state.type" :items="typeItems" value-key="value" class="w-full" />
          </UFormField>
          <UFormField name="tagsText" label="标签" description="使用逗号分隔，重复标签会自动合并。" :ui="{ description: 'min-h-5' }">
            <UInput v-model="state.tagsText" class="w-full" placeholder="cli, windows, 开源" />
          </UFormField>
        </div>

        <UFormField name="summary" label="摘要" description="用于列表与搜索结果，最多 500 个字符。">
          <UTextarea v-model="state.summary" class="w-full" :rows="3" autoresize :maxrows="6" placeholder="一句话介绍这个资源…" />
        </UFormField>
      </UForm>
    </template>

    <template #footer>
      <div class="flex w-full flex-col-reverse gap-2 p-4 sm:flex-row sm:items-center">
        <UButton v-if="resource" :to="`/manage/${resource.id}`" label="打开完整编辑器" icon="i-tabler-file-pencil" color="neutral" variant="ghost" class="justify-center sm:justify-start" />
        <div class="flex gap-2 sm:ml-auto">
          <UButton label="取消" color="neutral" variant="outline" class="flex-1 justify-center sm:flex-none" :disabled="saving" @click="close" />
          <UButton form="resource-quick-edit-form" type="submit" label="保存更改" icon="i-tabler-check" class="flex-1 justify-center sm:flex-none" :loading="saving" :disabled="!dirty" />
        </div>
      </div>
    </template>
  </UModal>
</template>
