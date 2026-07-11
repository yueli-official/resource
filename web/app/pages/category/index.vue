<script setup lang="ts">
import type { ListTaxonomies, TaxonomyView } from '~/types'

const { call } = useApi()
const q = ref('')
const sort = ref<'count' | 'name'>('count')

const { data } = await useAsyncData(
  'resource-categories-index',
  () => call<ListTaxonomies>('/api/v1/taxonomies', { query: { taxonomy: 'category' } })
)
const all = computed<TaxonomyView[]>(() => data.value?.items ?? [])
const roots = computed(() => all.value.filter(item => !item.parentId))
const childrenByParent = computed(() => {
  const map = new Map<string, TaxonomyView[]>()
  for (const item of all.value) {
    if (!item.parentId) continue
    if (!map.has(item.parentId)) map.set(item.parentId, [])
    map.get(item.parentId)!.push(item)
  }
  return map
})
const visible = computed(() => {
  const kw = q.value.trim().toLowerCase()
  const filtered = kw
    ? roots.value.filter(item => matches(item, kw) || (childrenByParent.value.get(item.id) || []).some(child => matches(child, kw)))
    : roots.value
  return [...filtered].sort((a, b) => sort.value === 'name'
    ? a.name.localeCompare(b.name, 'zh-CN')
    : b.count - a.count || a.name.localeCompare(b.name, 'zh-CN'))
})

function matches(item: TaxonomyView, keyword: string) {
  return `${item.name} ${item.slug} ${item.description || ''}`.toLowerCase().includes(keyword)
}

function visibleChildren(item: TaxonomyView) {
  const kw = q.value.trim().toLowerCase()
  const children = childrenByParent.value.get(item.id) || []
  const filtered = kw ? children.filter(child => matches(child, kw)) : children
  return [...filtered].sort((a, b) => b.count - a.count || a.name.localeCompare(b.name, 'zh-CN')).slice(0, 8)
}

useSeoMeta({ title: '资源分类' })
</script>

<template>
  <div class="mx-auto max-w-5xl">
    <header class="border-b border-default pb-6">
      <div class="flex items-center gap-2 text-primary">
        <span class="h-px w-6 bg-primary/40" />
        <span class="text-xs font-semibold uppercase tracking-[0.16em]">Categories</span>
      </div>
      <h1 class="font-display mt-3 text-3xl font-bold text-highlighted">资源分类</h1>
      <p class="mt-2 max-w-2xl text-sm leading-6 text-muted">按用途、类型和内容体系浏览资源。</p>
      <div class="mt-5 flex flex-wrap items-center gap-2">
        <UInput v-model="q" icon="i-tabler-search" placeholder="搜索分类名称或 slug" class="w-full sm:w-72" />
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
      <UIcon name="i-tabler-folder-off" class="mx-auto size-8" />
      <p class="mt-2 text-sm">没有匹配的分类</p>
    </div>

    <div v-else class="mt-8 grid gap-4 [grid-template-columns:repeat(auto-fill,minmax(min(18rem,100%),1fr))]">
      <NuxtLink
        v-for="category in visible"
        :key="category.id"
        :to="`/category/${category.slug}`"
        class="group rounded-lg border border-default bg-default p-4 transition hover:border-primary/40 hover:shadow-md"
      >
        <div class="flex items-start justify-between gap-3">
          <span class="grid size-10 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
            <UIcon name="i-tabler-folder" class="size-5" />
          </span>
          <UBadge :label="`${category.count} 个`" color="neutral" variant="subtle" />
        </div>
        <h2 class="font-display mt-4 line-clamp-1 text-lg font-semibold text-highlighted transition group-hover:text-primary">{{ category.name }}</h2>
        <p class="mt-1 line-clamp-2 min-h-[2.5rem] text-sm text-muted">{{ category.description || `/${category.slug}` }}</p>
        <div v-if="visibleChildren(category).length" class="mt-4 flex flex-wrap gap-1.5 border-t border-default pt-4">
          <span v-for="child in visibleChildren(category)" :key="child.id" class="rounded-full bg-elevated px-2 py-0.5 text-xs text-muted">
            {{ child.name }} <span class="text-dimmed">{{ child.count }}</span>
          </span>
        </div>
      </NuxtLink>
    </div>
  </div>
</template>
