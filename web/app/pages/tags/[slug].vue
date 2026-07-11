<script setup lang="ts">
import { ManagePagination } from '@platform/manage/components'
import type { ListResources, ListTaxonomies, TaxonomyView } from '~/types'

// Tag archive: flat (tags have no hierarchy), distinct route from categories.
const route = useRoute()
const slug = computed(() => route.params.slug as string)
const { call } = useApi()
const page = ref(1)
const size = 12

const { data: tagsData } = await useAsyncData(
  'all-tags',
  () => call<ListTaxonomies>('/api/v1/taxonomies', { query: { taxonomy: 'tag' } })
)
const current = computed<TaxonomyView | undefined>(() => (tagsData.value?.items ?? []).find(t => t.slug === slug.value))

const { data, pending } = await useAsyncData(
  `tag-${slug.value}`,
  () => call<ListResources>('/api/v1/resources', { query: { taxonomy: slug.value, page: page.value, size } }),
  { watch: [page, slug] }
)
const totalPages = computed(() => Math.max(1, Math.ceil((data.value?.total ?? 0) / size)))
watch(slug, () => { page.value = 1 })
useSeoMeta({ title: () => `#${current.value?.name || slug.value} · 标签` })
</script>

<template>
  <div>
    <nav class="mb-5 flex items-center gap-1.5 text-sm text-muted">
      <NuxtLink to="/" class="transition hover:text-primary">首页</NuxtLink>
      <UIcon name="i-tabler-chevron-right" class="size-3.5 opacity-60" />
      <span class="text-muted">标签</span>
    </nav>

    <div class="mb-6 flex items-center gap-2">
      <UIcon name="i-tabler-hash" class="size-6 text-primary" />
      <h1 class="font-display text-2xl font-semibold text-highlighted">{{ current?.name || slug }}</h1>
      <span class="text-sm text-dimmed">{{ data?.total ?? 0 }} 个</span>
    </div>

    <div v-if="pending" class="py-16 text-center text-muted">
      <UIcon name="i-tabler-loader-2" class="size-6 animate-spin" />
    </div>
    <div v-else-if="!data?.items?.length" class="py-16 text-center text-muted">
      <p class="text-sm">该标签下还没有资源</p>
    </div>
    <div v-else class="grid gap-4 [grid-template-columns:repeat(auto-fill,minmax(min(16rem,100%),1fr))]">
      <ResourceCard v-for="r in data.items" :key="r.id" :resource="r" />
    </div>

    <div class="mt-10 flex items-center justify-center">
      <ManagePagination v-model="page" :total-pages="totalPages" />
    </div>
  </div>
</template>
