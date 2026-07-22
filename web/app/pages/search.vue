<script setup lang="ts">
import { ManageEmpty, SkeletonList } from "@platform/manage/components";
import { CollectionPagination } from "@yueli/ui/collection/pattern";
import type { ListResources } from "~/types";

const route = useRoute();
const router = useRouter();
const { call } = useApi();
const q = ref((route.query.q as string) || "");
const page = ref(Number(route.query.page || 1));
const size = 20;

const query = computed(() => ((route.query.q as string) || "").trim());
const { data, pending } = await useAsyncData(
  () => `resource-search-${query.value}-${page.value}`,
  () => {
    if (!query.value)
      return Promise.resolve({ items: [], total: 0, page: 1, size });
    return call<ListResources>("/api/v1/resources", {
      query: { q: query.value, page: page.value, size },
    });
  },
  {
    watch: [query, page],
    default: () => ({ items: [], total: 0, page: 1, size }),
  },
);
const totalPages = computed(() =>
  Math.max(1, Math.ceil((data.value?.total ?? 0) / size)),
);

watch(
  () => route.query.q,
  () => {
    q.value = (route.query.q as string) || "";
    page.value = 1;
  },
);

watch(page, (value) => {
  if (!query.value) return;
  void router.replace({
    path: "/search",
    query: { q: query.value, page: value > 1 ? value : undefined },
  });
});

function submit() {
  const keyword = q.value.trim();
  page.value = 1;
  void router.push({ path: "/search", query: keyword ? { q: keyword } : {} });
}

useSeoMeta({
  title: () => (route.query.q ? `搜索“${route.query.q}”` : "搜索资源"),
});
</script>

<template>
  <div class="mx-auto max-w-5xl">
    <header class="border-b border-default pb-6">
      <div class="flex items-center gap-2 text-primary">
        <span class="h-px w-6 bg-primary/40" />
        <span class="text-xs font-semibold uppercase tracking-[0.16em]"
          >Search</span
        >
      </div>
      <h1 class="font-display mt-3 text-3xl font-bold text-highlighted">
        搜索资源
      </h1>
      <p class="mt-2 max-w-2xl text-sm leading-6 text-muted">
        按标题、简介和描述检索资源站内容。
      </p>
    </header>

    <form
      class="mt-6 flex gap-2 rounded-lg border border-default bg-default p-2"
      @submit.prevent="submit"
    >
      <UInput
        v-model="q"
        icon="i-tabler-search"
        placeholder="搜索资源标题、简介…"
        size="lg"
        class="min-w-0 flex-1"
        autofocus
      />
      <UButton type="submit" label="搜索" size="lg" :disabled="!q.trim()" />
    </form>

    <SkeletonList v-if="pending" class="mt-6" :rows="6" />

    <template v-else-if="route.query.q">
      <p class="mt-6 text-sm text-muted">
        “{{ route.query.q }}” · 共 {{ data?.total || 0 }} 个资源
      </p>
      <div
        v-if="data?.items?.length"
        class="mt-3 overflow-hidden rounded-lg border border-default bg-default"
      >
        <NuxtLink
          v-for="resource in data.items"
          :key="resource.id"
          :to="`/resources/${resource.id}`"
          class="group grid grid-cols-[72px_minmax(0,1fr)_auto] items-center gap-4 border-b border-default p-3 transition last:border-b-0 hover:bg-elevated/50 sm:grid-cols-[88px_minmax(0,1fr)_8rem_auto] sm:p-4"
        >
          <div
            class="size-[72px] shrink-0 overflow-hidden rounded-lg bg-elevated sm:size-[88px]"
          >
            <img
              v-if="resource.coverUrl"
              :src="resource.coverUrl"
              :alt="resource.title"
              class="size-full object-cover"
            />
            <div
              v-else
              class="grid size-full place-items-center bg-primary/10 text-primary"
            >
              <UIcon name="i-tabler-package" class="size-5" />
            </div>
          </div>
          <div class="min-w-0">
            <h3
              class="line-clamp-1 font-medium text-highlighted transition-colors group-hover:text-primary"
            >
              {{ resource.title }}
            </h3>
            <p class="mt-0.5 line-clamp-2 text-sm text-muted">
              {{ resource.summary || "未填写摘要" }}
            </p>
            <div class="mt-2 flex flex-wrap gap-1">
              <UBadge
                v-for="tag in resource.tags.slice(0, 3)"
                :key="tag"
                :label="`#${tag}`"
                color="neutral"
                variant="subtle"
                size="sm"
              />
            </div>
          </div>
          <div class="hidden text-right text-xs text-muted sm:block">
            <p class="font-medium text-highlighted">
              {{ resource.downloadCount }}
            </p>
            <p>下载</p>
          </div>
          <UIcon
            name="i-tabler-chevron-right"
            class="size-4 shrink-0 text-muted transition group-hover:translate-x-0.5"
          />
        </NuxtLink>
      </div>
      <ManageEmpty
        v-else
        class="mt-8"
        icon="i-tabler-search-off"
        :text="`没有匹配“${route.query.q}”的资源`"
      />

      <div class="mt-8 flex items-center justify-center">
        <CollectionPagination v-model="page" :total-pages="totalPages" />
      </div>
    </template>
  </div>
</template>
