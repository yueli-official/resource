<script setup lang="ts">
import type { ResourceView } from '~/types'

defineProps<{ resource: ResourceView }>()

const typeLabels: Record<string, string> = {
  software: '软件工具',
  design: '设计素材',
  script: '脚本',
  default: '资源'
}

function fmtCount(n: number) {
  return n >= 1000 ? (n / 1000).toFixed(1) + 'k' : String(n)
}

function typeLabel(type: string) {
  return typeLabels[type] || type
}
</script>

<template>
  <NuxtLink
    :to="`/resources/${resource.id}`"
    class="resource-surface-card resource-surface-interactive group flex min-w-0 flex-col overflow-hidden rounded-lg"
  >
    <div class="relative h-[clamp(9rem,14vw,12rem)] overflow-hidden bg-elevated">
      <img
        v-if="resource.coverUrl"
        :src="resource.coverUrl"
        :alt="resource.title"
        class="size-full object-cover transition duration-300 group-hover:scale-105"
      >
      <div v-else class="grid size-full place-items-center bg-gradient-to-br from-primary/10 to-elevated text-primary/55">
        <UIcon name="i-tabler-package" class="size-10" />
      </div>
      <span class="absolute left-3 top-3 rounded-full bg-default/90 px-2.5 py-1 text-xs font-medium text-primary shadow-sm backdrop-blur">
        {{ typeLabel(resource.type) }}
      </span>
    </div>
    <div class="flex flex-1 flex-col p-4">
      <h3 class="font-display line-clamp-2 text-base font-semibold leading-snug text-highlighted transition group-hover:text-primary">{{ resource.title }}</h3>
      <p class="mt-2 line-clamp-2 flex-1 text-sm leading-6 text-muted">{{ resource.summary || '未填写摘要' }}</p>
      <div class="mt-4 flex items-end justify-between gap-3 text-xs text-muted">
        <div class="flex min-w-0 flex-wrap gap-1">
          <UBadge v-for="t in resource.tags.slice(0, 2)" :key="t" :label="`#${t}`" color="neutral" variant="subtle" size="sm" />
        </div>
        <span class="flex items-center gap-1 whitespace-nowrap">
          <UIcon name="i-tabler-download" class="size-3.5" />{{ fmtCount(resource.downloadCount) }}
        </span>
      </div>
    </div>
  </NuxtLink>
</template>
