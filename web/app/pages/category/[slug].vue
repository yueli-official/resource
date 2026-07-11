<script setup lang="ts">
import { ManagePagination } from '@platform/ui/components'
import type { ListResources, ListTaxonomies, TaxonomyView } from '~/types'

// Category archive: multi-level — breadcrumb + sub-category entries + a card grid.
const route = useRoute()
const slug = computed(() => route.params.slug as string)
const { call } = useApi()
const page = ref(1)
const size = 12

const { data: catsData } = await useAsyncData(
  'all-categories',
  () => call<ListTaxonomies>('/api/v1/taxonomies', { query: { taxonomy: 'category' } })
)
const cats = computed(() => catsData.value?.items ?? [])
const byId = computed(() => new Map(cats.value.map(c => [c.id, c])))
const current = computed<TaxonomyView | undefined>(() => cats.value.find(c => c.slug === slug.value))
const children = computed(() => current.value ? cats.value.filter(c => c.parentId === current.value!.id) : [])
const crumbs = computed(() => {
  const chain: TaxonomyView[] = []
  let node = current.value
  while (node) {
    chain.unshift(node)
    node = node.parentId ? byId.value.get(node.parentId) : undefined
  }
  return chain
})

const { data, pending } = await useAsyncData(
  `cat-${slug.value}`,
  () => call<ListResources>('/api/v1/resources', { query: { taxonomy: slug.value, page: page.value, size } }),
  { watch: [page, slug] }
)
const totalPages = computed(() => Math.max(1, Math.ceil((data.value?.total ?? 0) / size)))
watch(slug, () => { page.value = 1 })
useSeoMeta({ title: () => `${current.value?.name || slug.value} · 分类` })
</script>

<template>
  <div>
    <!-- breadcrumb -->
    <nav class="mb-5 flex flex-wrap items-center gap-1.5 text-sm text-muted">
      <NuxtLink to="/" class="transition hover:text-primary">首页</NuxtLink>
      <template v-for="c in crumbs" :key="c.id">
        <UIcon name="i-tabler-chevron-right" class="size-3.5 opacity-60" />
        <NuxtLink
          :to="`/category/${c.slug}`"
          class="transition hover:text-primary"
          :class="c.id === current?.id ? 'text-highlighted font-medium' : ''"
        >{{ c.name }}</NuxtLink>
      </template>
    </nav>

    <div class="mb-6 flex items-start gap-2">
      <UIcon name="i-tabler-folder" class="mt-1 size-6 shrink-0 text-primary" />
      <div>
        <h1 class="font-display text-2xl font-semibold text-highlighted">{{ current?.name || slug }}</h1>
        <p v-if="current?.description" class="mt-1 text-sm text-muted">{{ current.description }}</p>
        <p class="mt-1 text-xs text-dimmed">{{ data?.total ?? 0 }} 个资源</p>
      </div>
    </div>

    <!-- sub-category entries -->
    <div v-if="children.length" class="mb-8 flex flex-wrap gap-2">
      <NuxtLink
        v-for="c in children"
        :key="c.id"
        :to="`/category/${c.slug}`"
        class="inline-flex items-center gap-1.5 rounded-full border border-default px-3 py-1 text-sm text-default transition hover:border-primary/40 hover:text-primary"
      >
        <UIcon name="i-tabler-folder" class="size-3.5" />{{ c.name }}
        <span class="text-xs text-dimmed">{{ c.count }}</span>
      </NuxtLink>
    </div>

    <div v-if="pending" class="py-16 text-center text-muted">
      <UIcon name="i-tabler-loader-2" class="size-6 animate-spin" />
    </div>
    <div v-else-if="!data?.items?.length" class="py-16 text-center text-muted">
      <p class="text-sm">该分类下还没有资源</p>
    </div>
    <div v-else class="grid gap-4 [grid-template-columns:repeat(auto-fill,minmax(min(16rem,100%),1fr))]">
      <ResourceCard v-for="r in data.items" :key="r.id" :resource="r" />
    </div>

    <div class="mt-10 flex items-center justify-center">
      <ManagePagination v-model="page" :total-pages="totalPages" />
    </div>
  </div>
</template>
