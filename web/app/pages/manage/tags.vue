<script setup lang="ts">
definePageMeta({ layout: 'manage', middleware: 'auth' })
useSeoMeta({ title: '资源标签 · 控制台' })
</script>

<template>
  <ResourceTaxonomyManager kind="tag" />
</template>
