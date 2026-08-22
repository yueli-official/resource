<script setup lang="ts">
import { AssetRegistrationSummary } from '@yueli/asset-nuxt/components'
import { PageHeader } from '@yueli/ui/dashboard/pattern'

definePageMeta({ layout: 'manage', middleware: 'auth' })

useSeoMeta({ title: '资源策略 · 控制台' })
</script>

<template>
  <div class="space-y-5">
    <PageHeader title="资源策略" />
    <ClientOnly>
      <AssetRegistrationSummary expected-namespace="resource" :profile-order="['resource-cover', 'resource-content', 'resource']" />
      <template #fallback><USkeleton class="h-56 w-full rounded-xl" /></template>
    </ClientOnly>
  </div>
</template>
