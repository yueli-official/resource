<script setup lang="ts">
const { brand: siteBrand } = useSiteRuntime()
useHead({ titleTemplate: title => (title ? `${title} · ${siteBrand.value}` : siteBrand.value) })
</script>

<template>
  <UApp>
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </UApp>
</template>
