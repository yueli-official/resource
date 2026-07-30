<script setup lang="ts">
const { label = "配置项" } = defineProps<{ label?: string }>();
</script>

<template>
  <article class="grid min-w-0 gap-3 rounded-lg border border-default bg-elevated/25 p-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-start" :aria-label="label">
    <div class="min-w-0"><slot /></div>
    <div v-if="$slots.actions" class="flex shrink-0 items-center justify-end gap-1"><slot name="actions" /></div>
  </article>
</template>
