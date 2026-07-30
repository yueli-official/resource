<script setup lang="ts">
import { CollectionPagination } from "@yueli/ui/collection/pattern";
import SkeletonCards from "~/components/SkeletonCards.vue";
import { useMinimumLoading } from "@yueli/ui/feedback";
import type { ListResources, ListTaxonomies, ResourceView } from "~/types";

const { call } = useApi();
const { homeSettings, siteSettings } = useResourceSettings();
const router = useRouter();
const type = ref("");
const page = ref(1);
const searchQ = ref("");
const size = computed(() => siteSettings.value.resource.resourcesPerPage || 12);

const typeFilters = [
  { label: "全部", value: "" },
  { label: "软件/工具", value: "software" },
  { label: "设计素材", value: "design" },
  { label: "脚本", value: "script" },
];

const { data: catsData } = await useAsyncData("home-categories", () =>
  call<ListTaxonomies>("/api/v1/taxonomies", {
    query: { taxonomy: "category" },
  }),
);
const { data: tagsData } = await useAsyncData("home-tags", () =>
  call<ListTaxonomies>("/api/v1/taxonomies", { query: { taxonomy: "tag" } }),
);
const categories = computed(() => catsData.value?.items ?? []);
const tags = computed(() =>
  [...(tagsData.value?.items ?? [])]
    .sort((a, b) => b.count - a.count)
    .slice(0, 18),
);
const visibleCategories = computed(() => {
  const slugs = homeSettings.value.categorySlugs || [];
  if (!slugs.length) return categories.value;
  const set = new Set(slugs);
  return categories.value.filter((category) => set.has(category.slug));
});

const { data, pending } = await useAsyncData(
  "resources-list",
  () =>
    call<ListResources>("/api/v1/resources", {
      query: {
        type: type.value || undefined,
        page: page.value,
        size: size.value,
      },
    }),
  { watch: [type, page, size] },
);
const showSkeleton = useMinimumLoading(pending);
const resources = computed<ResourceView[]>(() => data.value?.items ?? []);
const featuredResources = computed(() => {
  const byId = new Map(resources.value.map((item) => [item.id, item]));
  const configured = (homeSettings.value.featuredResourceIds || [])
    .map((id) => byId.get(id))
    .filter(Boolean) as ResourceView[];
  if (configured.length) return configured.slice(0, 3);
  return [...resources.value]
    .sort((a, b) => b.downloadCount - a.downloadCount)
    .slice(0, 3);
});
const lead = computed(() => featuredResources.value[0] || resources.value[0]);
const popular = computed(() =>
  [...resources.value]
    .sort((a, b) => b.downloadCount - a.downloadCount)
    .slice(0, 5),
);
const recommended = computed(() => {
  const seen = new Set<string>();
  const items: ResourceView[] = [];
  for (const item of [...featuredResources.value, ...popular.value]) {
    if (item.id === lead.value?.id || seen.has(item.id)) continue;
    seen.add(item.id);
    items.push(item);
  }
  return items.slice(0, 5);
});

function setType(v: string) {
  type.value = v;
  page.value = 1;
}
const totalPages = computed(() =>
  Math.max(1, Math.ceil((data.value?.total ?? 0) / size.value)),
);

function goSearch() {
  const q = searchQ.value.trim();
  if (q) void router.push({ path: "/search", query: { q } });
}
</script>

<template>
  <div class="space-y-14 sm:space-y-20">
    <header class="border-b border-default pb-7">
      <div class="max-w-3xl">
        <div class="flex items-center gap-2 text-primary">
          <span class="h-px w-6 bg-primary/40" />
          <span class="text-xs font-semibold uppercase tracking-[0.16em]"
            >Resource Library</span
          >
        </div>
        <h1
          class="font-display mt-3 text-4xl font-bold leading-[1.05] tracking-tight text-highlighted sm:text-5xl"
        >
          {{ homeSettings.heroTitle }}
        </h1>
        <p class="mt-4 max-w-2xl text-base leading-7 text-muted">
          {{ homeSettings.heroSubtitle }}
        </p>
        <form class="mt-6 flex max-w-2xl gap-2" @submit.prevent="goSearch">
          <UInput
            v-model="searchQ"
            icon="i-tabler-search"
            placeholder="搜索软件、素材、脚本…"
            size="lg"
            class="min-w-0 flex-1"
          />
          <UButton
            type="submit"
            icon="i-tabler-search"
            label="搜索"
            size="lg"
            :disabled="!searchQ.trim()"
          />
        </form>
      </div>
    </header>

    <section
      v-if="lead"
      class="grid gap-6 md:h-[450px] md:grid-cols-[minmax(0,1.7fr)_minmax(0,1fr)] md:gap-8"
    >
      <NuxtLink
        :to="`/resources/${lead.id}`"
        class="resource-surface-card resource-surface-interactive group flex min-w-0 flex-col overflow-hidden rounded-lg md:min-h-0"
      >
        <div
          class="relative aspect-[16/9] overflow-hidden bg-elevated md:aspect-auto md:min-h-0 md:flex-1"
        >
          <img
            v-if="lead.coverUrl"
            :src="lead.coverUrl"
            :alt="lead.title"
            class="size-full object-cover transition duration-700 group-hover:scale-[1.03]"
          />
          <div
            v-else
            class="grid size-full place-items-center bg-gradient-to-br from-primary/15 to-elevated text-primary/40"
          >
            <UIcon name="i-tabler-package" class="size-20 md:size-24" />
          </div>
          <span
            class="absolute left-4 top-4 inline-flex items-center gap-1 rounded-full bg-default/90 px-3 py-1 text-xs font-semibold text-primary shadow-sm backdrop-blur"
          >
            <UIcon name="i-tabler-sparkles" class="size-3.5" />精选资源
          </span>
        </div>
        <div class="flex flex-col p-5 sm:p-6 md:flex-none">
          <div class="flex flex-wrap items-center gap-2 text-xs text-muted">
            <span class="flex items-center gap-1"
              ><UIcon name="i-tabler-download" class="size-3.5" />{{
                lead.downloadCount
              }}
              次下载</span
            >
            <span>/{{ lead.slug }}</span>
          </div>
          <h2
            class="font-display mt-3 line-clamp-1 text-xl font-bold leading-tight text-highlighted transition group-hover:text-primary sm:text-2xl"
          >
            {{ lead.title }}
          </h2>
          <p class="mt-2 line-clamp-2 min-h-[2lh] text-sm leading-6 text-muted">
            {{ lead.summary || "未填写摘要" }}
          </p>
        </div>
      </NuxtLink>

      <div v-if="recommended.length" class="flex flex-col">
        <h2
          class="font-display mb-1 flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.14em] text-muted"
        >
          <span class="h-px w-5 bg-primary/50" />推荐下载
        </h2>
        <ul class="flex flex-col divide-y divide-default">
          <li v-for="(item, i) in recommended" :key="item.id">
            <NuxtLink
              :to="`/resources/${item.id}`"
              class="group flex items-center gap-4 py-4"
            >
              <span
                class="font-display w-7 shrink-0 text-lg font-bold tabular-nums text-primary/30"
                >{{ String(i + 1).padStart(2, "0") }}</span
              >
              <div class="min-w-0">
                <h3
                  class="font-display line-clamp-2 font-semibold leading-snug text-highlighted transition group-hover:text-primary"
                >
                  {{ item.title }}
                </h3>
                <div class="mt-1.5 flex items-center gap-3 text-xs text-muted">
                  <span class="flex items-center gap-1"
                    ><UIcon name="i-tabler-download" class="size-3.5" />{{
                      item.downloadCount
                    }}
                    次下载</span
                  >
                  <span>/{{ item.slug }}</span>
                </div>
              </div>
            </NuxtLink>
          </li>
        </ul>
      </div>
    </section>

    <section class="grid gap-10 lg:grid-cols-[minmax(0,1fr)_280px]">
      <div class="min-w-0">
        <div
          class="mb-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
        >
          <h2 class="font-display text-xl font-semibold text-highlighted">
            最新资源
          </h2>
          <div class="flex flex-wrap gap-2">
            <UButton
              v-for="f in typeFilters"
              :key="f.value"
              :color="type === f.value ? 'primary' : 'neutral'"
              :variant="type === f.value ? 'solid' : 'outline'"
              size="sm"
              :label="f.label"
              @click="setType(f.value)"
            />
          </div>
        </div>

        <SkeletonCards v-if="showSkeleton" :count="6" />
        <div
          v-else-if="!resources.length"
          class="rounded-lg border border-dashed border-default py-20 text-center text-muted"
        >
          <div
            class="mx-auto grid size-14 place-items-center rounded-lg bg-primary/10 text-primary"
          >
            <UIcon name="i-tabler-package-off" class="size-7" />
          </div>
          <p class="mt-3 text-sm">还没有资源</p>
        </div>
        <div
          v-else
          class="grid gap-4 [grid-template-columns:repeat(auto-fill,minmax(min(17rem,100%),1fr))]"
        >
          <ResourceCard v-for="r in resources" :key="r.id" :resource="r" />
        </div>

        <div
          v-if="totalPages > 1"
          class="mt-10 flex items-center justify-center"
        >
          <CollectionPagination v-model="page" :total-pages="totalPages" />
        </div>
      </div>

      <aside class="space-y-8 lg:sticky lg:top-20 lg:self-start">
        <div v-if="visibleCategories.length">
          <h3
            class="mb-3 flex items-center gap-2 font-display font-semibold text-highlighted"
          >
            <UIcon name="i-tabler-folders" class="size-4 text-primary" />分类
          </h3>
          <div class="space-y-1.5">
            <NuxtLink
              v-for="c in visibleCategories.slice(0, 10)"
              :key="c.id"
              :to="`/category/${c.slug}`"
              class="flex items-center justify-between gap-3 rounded-md px-2 py-1.5 text-sm text-muted transition hover:bg-elevated/60 hover:text-primary"
            >
              <span class="line-clamp-1">{{ c.name }}</span>
              <span class="text-xs text-dimmed">{{ c.count }}</span>
            </NuxtLink>
          </div>
        </div>
        <div v-if="tags.length">
          <h3
            class="mb-3 flex items-center gap-2 font-display font-semibold text-highlighted"
          >
            <UIcon name="i-tabler-hash" class="size-4 text-primary" />标签
          </h3>
          <div class="flex flex-wrap gap-1.5">
            <NuxtLink
              v-for="tag in tags"
              :key="tag.id"
              :to="`/tags/${tag.slug}`"
              class="rounded-full bg-elevated px-2.5 py-1 text-xs text-muted transition hover:text-primary"
              >#{{ tag.name }}</NuxtLink
            >
          </div>
        </div>
      </aside>
    </section>
  </div>
</template>
