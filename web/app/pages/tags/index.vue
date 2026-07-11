<script setup lang="ts">
import type { ListTaxonomies, TaxonomyView } from '~/types'

const { call } = useApi()
const q = ref('')
const sort = ref<'count' | 'name'>('count')

const { data } = await useAsyncData(
  'resource-tags-index',
  () => call<ListTaxonomies>('/api/v1/taxonomies', { query: { taxonomy: 'tag' } })
)
const all = computed<TaxonomyView[]>(() => data.value?.items ?? [])
const visible = computed(() => {
  const kw = q.value.trim().toLowerCase()
  const filtered = kw
    ? all.value.filter(item => `${item.name} ${item.slug} ${item.description || ''}`.toLowerCase().includes(kw))
    : all.value
  return [...filtered].sort((a, b) => sort.value === 'name'
    ? a.name.localeCompare(b.name, 'zh-CN')
    : b.count - a.count || a.name.localeCompare(b.name, 'zh-CN'))
})

useSeoMeta({ title: '资源标签' })
</script>

<template>
  <div class="mx-auto max-w-5xl">
    <header class="border-b border-default pb-6">
      <div class="flex items-center gap-2 text-primary">
        <span class="h-px w-6 bg-primary/40" />
        <span class="text-xs font-semibold uppercase tracking-[0.16em]">Tags</span>
      </div>
      <h1 class="font-display mt-3 text-3xl font-bold text-highlighted">资源标签</h1>
      <p class="mt-2 max-w-2xl text-sm leading-6 text-muted">用关键词快速定位资源集合。</p>
      <div class="mt-5 flex flex-wrap items-center gap-2">
        <UInput v-model="q" icon="i-tabler-search" placeholder="搜索标签名称或 slug" class="w-full sm:w-72" />
        <USelectMenu
          v-model="sort"
          :items="[{ label: '按资源数', value: 'count' }, { label: '按名称', value: 'name' }]"
          value-key="value"
          icon="i-tabler-arrows-sort"
          class="w-full sm:w-36"
        />
      </div>
    </header>

    <div v-if="!visible.length" class="mt-8 rounded-lg border border-dashed border-default py-16 text-center text-muted">
      <UIcon name="i-tabler-tags-off" class="mx-auto size-8" />
      <p class="mt-2 text-sm">没有匹配的标签</p>
    </div>

    <div v-else class="mt-8 flex flex-wrap gap-2">
      <NuxtLink
        v-for="tag in visible"
        :key="tag.id"
        :to="`/tags/${tag.slug}`"
        class="group inline-flex items-center gap-2 rounded-full border border-default bg-default px-3.5 py-2 text-sm text-muted transition hover:border-primary/40 hover:text-primary"
      >
        <UIcon name="i-tabler-hash" class="size-4 text-primary/70" />
        <span>{{ tag.name }}</span>
        <span class="rounded-full bg-elevated px-2 py-0.5 text-xs text-dimmed group-hover:text-primary">{{ tag.count }}</span>
      </NuxtLink>
    </div>
  </div>
</template>
