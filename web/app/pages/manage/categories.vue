<script setup lang="ts">
definePageMeta({ layout: 'manage', middleware: 'auth' })
useSeoMeta({ title: '资源分类 · 控制台' })
</script>

<template>
  <ResourceTaxonomyManager kind="category" />
</template>
