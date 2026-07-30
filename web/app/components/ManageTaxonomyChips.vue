<script setup lang="ts">
import { computed } from 'vue'

export interface ManageTaxonomyChip {
  key: string
  label: string
  kind?: 'category' | 'tag'
}

const props = withDefaults(defineProps<{
  items: readonly ManageTaxonomyChip[]
  max?: number
}>(), { max: 2 })

const limit = computed(() => Math.max(1, props.max))
const visible = computed(() => props.items.slice(0, limit.value))
const overflow = computed(() => Math.max(0, props.items.length - visible.value.length))
const accessibleLabel = computed(() => props.items.map(item => item.kind === 'tag' ? `标签 ${item.label}` : `分类 ${item.label}`).join('，'))

function displayLabel(item: ManageTaxonomyChip) {
  return item.kind === 'tag' ? `#${item.label.replace(/^#/, '')}` : item.label
}
</script>

<template>
  <div v-if="items.length" class="flex min-w-0 flex-wrap items-center gap-1" :aria-label="accessibleLabel">
    <UBadge
      v-for="item in visible"
      :key="item.key"
      :label="displayLabel(item)"
      color="neutral"
      variant="subtle"
      size="sm"
    />
    <UBadge v-if="overflow" :label="`+${overflow}`" color="neutral" variant="soft" size="sm" :aria-label="`另有 ${overflow} 项分类或标签`" />
  </div>
</template>
