<script setup lang="ts">
// Shared social-share rail for consumer-site detail pages. URLs are built
// client-side from window.location, so it renders only after mount (a plain
// `v-if="mounted"` rather than <ClientOnly>, to avoid depending on a Nuxt
// built-in resolving inside this workspace-package SFC). No backend — pure
// outbound intent links + clipboard + Web Share API.
const props = defineProps<{ title: string }>()

const mounted = ref(false)
const url = ref('')
const copied = ref(false)
const canNativeShare = ref(false)

onMounted(() => {
  mounted.value = true
  url.value = window.location.href
  canNativeShare.value = typeof navigator !== 'undefined' && typeof navigator.share === 'function'
})

const targets = computed(() => {
  const u = encodeURIComponent(url.value)
  const t = encodeURIComponent(props.title)
  return [
    { key: 'weibo', label: '分享到微博', icon: 'i-tabler-brand-weibo', href: `https://service.weibo.com/share/share.php?url=${u}&title=${t}` },
    { key: 'x', label: '分享到 X', icon: 'i-tabler-brand-x', href: `https://twitter.com/intent/tweet?url=${u}&text=${t}` },
    { key: 'qq', label: '分享到 QQ 空间', icon: 'i-tabler-brand-qq', href: `https://sns.qzone.qq.com/cgi-bin/qzshare/cgi_qzshare_onekey?url=${u}&title=${t}` }
  ]
})

async function copyLink() {
  try {
    await navigator.clipboard.writeText(url.value)
    copied.value = true
    setTimeout(() => (copied.value = false), 1800)
  } catch {
    // clipboard blocked (insecure context / permission) — leave the icon as-is
  }
}

async function nativeShare() {
  try {
    await navigator.share({ title: props.title, url: url.value })
  } catch {
    // user dismissed the share sheet — nothing to do
  }
}
</script>

<template>
  <div v-if="mounted" class="flex items-center gap-1">
    <span class="mr-1 hidden text-xs font-medium text-muted sm:inline">分享</span>
    <UButton
      v-for="tgt in targets"
      :key="tgt.key"
      :icon="tgt.icon"
      :to="tgt.href"
      target="_blank"
      rel="noopener noreferrer"
      color="neutral"
      variant="ghost"
      size="sm"
      square
      :aria-label="tgt.label"
      :title="tgt.label"
    />
    <UButton
      v-if="canNativeShare"
      icon="i-tabler-share-2"
      color="neutral"
      variant="ghost"
      size="sm"
      square
      aria-label="系统分享"
      title="系统分享"
      @click="nativeShare"
    />
    <UButton
      :icon="copied ? 'i-tabler-check' : 'i-tabler-link'"
      :color="copied ? 'primary' : 'neutral'"
      variant="ghost"
      size="sm"
      square
      :aria-label="copied ? '已复制链接' : '复制链接'"
      :title="copied ? '已复制链接' : '复制链接'"
      @click="copyLink"
    />
  </div>
</template>
