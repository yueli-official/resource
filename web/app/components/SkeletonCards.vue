<script setup lang="ts">
// Shared post-card loading skeleton: a responsive grid of placeholder cards
// shaped like the public cover cards (16:9 cover + title + excerpt + meta).
// Mirrors the cover-card grid used by the home / category / tag lists so the
// layout doesn't jump when data lands. USkeleton resolves globally in the apps.
withDefaults(defineProps<{ count?: number }>(), { count: 6 })
</script>

<template>
  <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 sm:gap-6 lg:grid-cols-3">
    <div v-for="i in count" :key="i" class="flex flex-col overflow-hidden rounded-2xl border border-default">
      <USkeleton class="aspect-[16/9] w-full rounded-none" />
      <div class="flex flex-1 flex-col gap-2.5 p-5">
        <USkeleton class="h-4 w-3/4" />
        <USkeleton class="h-3 w-full" />
        <div class="flex gap-3 pt-1">
          <USkeleton class="h-3 w-16" />
          <USkeleton class="h-3 w-12" />
        </div>
      </div>
    </div>
  </div>
</template>
