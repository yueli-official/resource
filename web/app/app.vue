<script setup lang="ts">
import { FeedbackToastRegion } from "@yueli/ui/feedback/pattern";

const { brand: siteBrand } = useSiteRuntime()
useHead({
  htmlAttrs: { lang: "zh-CN" },
  titleTemplate: title => (title ? `${title} · ${siteBrand.value}` : siteBrand.value),
})
</script>

<template>
  <UApp :toaster="null">
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
    <FeedbackToastRegion />
  </UApp>
</template>
