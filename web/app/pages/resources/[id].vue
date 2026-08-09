<script setup lang="ts">
import ShareBar from "~/components/ShareBar.vue";
import type {
  DeliveryItemView,
  ListResources,
  ResourceAssetView,
  ResourceDetail,
  ResourceView,
} from "~/types";

definePageMeta({ middleware: "url-lifecycle" });
const route = useRoute();
const id = route.params.id as string;
const { call } = useApi();
const { siteSettings } = useResourceSettings();

const { data, error } = await useAsyncData(`resource-${id}`, () =>
  call<ResourceDetail>(`/api/v1/resources/${id}`),
);
if (error.value || !data.value?.resource) {
  throw createError({
    statusCode: 404,
    statusMessage: "资源不存在",
    fatal: true,
  });
}

const r = computed(() => data.value!.resource);
const assets = computed(() => data.value?.assets ?? []);
const assetById = computed(
  () => new Map(assets.value.map((asset) => [asset.assetId, asset])),
);
const taxonomies = computed(() => data.value?.taxonomies ?? []);
const categories = computed(() =>
  taxonomies.value.filter((t) => t.taxonomy === "category"),
);
const tags = computed(() =>
  taxonomies.value.filter((t) => t.taxonomy === "tag"),
);
const relatedTaxonomy = computed(
  () => categories.value[0]?.slug || tags.value[0]?.slug || "",
);

const { data: related } = await useAsyncData(`resource-related-${id}`, () =>
  call<ListResources>("/api/v1/resources", {
    query: {
      taxonomy: relatedTaxonomy.value || undefined,
      type: relatedTaxonomy.value ? undefined : r.value.type,
      size: 5,
    },
  }).catch(() => ({ items: [] as ResourceView[], total: 0, page: 1, size: 5 })),
);

const relatedItems = computed(() =>
  (related.value?.items ?? [])
    .filter((item) => item.id !== r.value.id)
    .slice(0, 4),
);
const fileDeliveryAssets = computed(() => {
  const items =
    r.value.deliveryPayload?.items?.filter(
      (item) =>
        item.kind === "asset_file" && item.enabled !== false && item.assetId,
    ) || [];
  if (!items.length) return assets.value;
  return items
    .map((item) => assetById.value.get(item.assetId || ""))
    .filter(Boolean) as ResourceAssetView[];
});
const primaryAsset = computed(() => fileDeliveryAssets.value[0]);
const netdiskItems = computed<DeliveryItemView[]>(() =>
  (r.value.deliveryPayload?.items || []).filter(
    (item) =>
      item.kind === "netdisk" && item.enabled !== false && !!item.netdisk?.url,
  ),
);
const primaryNetdisk = computed(() => netdiskItems.value[0]);
const deliveryCount = computed(
  () => fileDeliveryAssets.value.length + netdiskItems.value.length,
);
const totalSize = computed(() =>
  fileDeliveryAssets.value.reduce((sum, asset) => sum + (asset.size || 0), 0),
);
const typeLabel = computed(
  () =>
    ({
      software: "软件工具",
      design: "设计素材",
      script: "脚本",
    })[r.value.type] ||
    r.value.type ||
    "资源",
);

useDiscoveryPage(() => data.value?.discovery);

onMounted(() => {
  const viewEvent = {
    // identifier-gate: allow Traffic replay key owned by the view-event contract
    eventId: crypto.randomUUID(),
    occurredAt: new Date().toISOString(),
  };
  const recordView = () =>
    call(`/api/v1/resources/${id}/view`, {
      method: "POST",
      body: viewEvent,
    });
  recordView().catch(() => recordView().catch(() => {}));
});

const downloading = ref("");
async function download(assetId: string) {
  downloading.value = assetId;
  try {
    const res = await call<{ deliveryUrl: string }>(
      `/api/v1/resources/${id}/download/${assetId}`,
    );
    if (res.deliveryUrl) window.location.href = res.deliveryUrl;
  } finally {
    downloading.value = "";
  }
}

function fmtSize(n: number) {
  if (!n) return "0 B";
  if (n >= 1 << 30) return `${(n / (1 << 30)).toFixed(2)} GB`;
  if (n >= 1 << 20) return `${(n / (1 << 20)).toFixed(1)} MB`;
  if (n >= 1 << 10) return `${(n / (1 << 10)).toFixed(1)} KB`;
  return `${n} B`;
}

function fmtDate(value: string) {
  if (!value) return "";
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date(value));
}

function assetIcon(asset: ResourceAssetView) {
  const file = asset.filename.toLowerCase();
  const mime = asset.mime.toLowerCase();
  if (mime.startsWith("image/")) return "i-tabler-photo";
  if (mime.startsWith("video/")) return "i-tabler-video";
  if (file.endsWith(".zip") || file.endsWith(".7z") || file.endsWith(".rar"))
    return "i-tabler-file-zip";
  if (file.endsWith(".pdf")) return "i-tabler-file-type-pdf";
  if (
    file.endsWith(".psd") ||
    file.endsWith(".fig") ||
    file.endsWith(".sketch")
  )
    return "i-tabler-palette";
  if (
    file.endsWith(".js") ||
    file.endsWith(".ts") ||
    file.endsWith(".py") ||
    file.endsWith(".sh")
  )
    return "i-tabler-code";
  return "i-tabler-file";
}

function netdiskTitle(item: DeliveryItemView) {
  return item.title || item.netdisk?.provider || "网盘交付";
}
</script>

<template>
  <div class="space-y-8 pb-10">
    <nav
      class="flex flex-wrap items-center gap-1.5 text-sm text-muted"
      aria-label="面包屑"
    >
      <NuxtLink to="/" class="transition hover:text-primary">{{
        siteSettings.site.siteName
      }}</NuxtLink>
      <UIcon name="i-tabler-chevron-right" class="size-3.5 opacity-60" />
      <NuxtLink
        v-if="categories[0]"
        :to="`/category/${categories[0].slug}`"
        class="transition hover:text-primary"
        >{{ categories[0].name }}</NuxtLink
      >
      <span v-else>资源详情</span>
    </nav>

    <section class="resource-surface-card overflow-hidden rounded-xl">
      <div class="grid lg:grid-cols-[minmax(0,1.25fr)_minmax(320px,0.75fr)]">
        <div class="relative min-h-[18rem] bg-elevated lg:min-h-[31rem]">
          <img
            v-if="r.coverUrl"
            :src="r.coverUrl"
            :alt="r.title"
            class="size-full object-cover"
          />
          <div
            v-else
            class="grid size-full place-items-center bg-gradient-to-br from-primary/12 to-elevated text-primary/45"
          >
            <UIcon name="i-tabler-package" class="size-20" />
          </div>
          <div class="absolute left-4 top-4 flex flex-wrap gap-2">
            <UBadge
              :label="typeLabel"
              color="primary"
              variant="soft"
              icon="i-tabler-package"
            />
            <UBadge
              :label="`${deliveryCount} 项交付`"
              color="neutral"
              variant="subtle"
              icon="i-tabler-files"
            />
          </div>
        </div>

        <div class="flex min-w-0 flex-col p-5 sm:p-7">
          <div class="flex flex-wrap items-center gap-2">
            <NuxtLink
              v-for="c in categories"
              :key="c.id"
              :to="`/category/${c.slug}`"
              class="inline-flex items-center gap-1 rounded-md border border-default bg-default/75 px-2.5 py-1 text-xs font-medium text-muted transition hover:border-primary/40 hover:text-primary"
            >
              <UIcon name="i-tabler-folder" class="size-3.5" />{{ c.name }}
            </NuxtLink>
          </div>

          <h1
            class="font-display mt-5 text-balance text-3xl font-bold leading-tight text-highlighted sm:text-4xl"
          >
            {{ r.title }}
          </h1>
          <p v-if="r.summary" class="mt-4 text-base leading-7 text-muted">
            {{ r.summary }}
          </p>

          <div class="mt-6 grid gap-3 text-sm text-muted sm:grid-cols-2">
            <div class="resource-surface-inset rounded-lg p-3">
              <span class="flex items-center gap-2"
                ><UIcon
                  name="i-tabler-download"
                  class="size-4 text-primary"
                />下载次数</span
              >
              <strong
                class="mt-1 block text-xl font-semibold text-highlighted"
                >{{ r.downloadCount }}</strong
              >
            </div>
            <div class="resource-surface-inset rounded-lg p-3">
              <span class="flex items-center gap-2"
                ><UIcon
                  name="i-tabler-database"
                  class="size-4 text-primary"
                />文件总量</span
              >
              <strong
                class="mt-1 block text-xl font-semibold text-highlighted"
                >{{ fmtSize(totalSize) }}</strong
              >
            </div>
          </div>

          <div class="mt-auto pt-7">
            <UAlert
              v-if="!siteSettings.resource.downloadsEnabled"
              color="warning"
              variant="subtle"
              icon="i-tabler-lock"
              title="下载暂时关闭"
              description="站点管理员已暂时关闭前台下载。"
              class="mb-4"
            />
            <div class="flex flex-col gap-3 sm:flex-row">
              <UButton
                v-if="siteSettings.resource.downloadsEnabled && primaryAsset"
                icon="i-tabler-download"
                label="下载主文件"
                size="lg"
                class="justify-center"
                :loading="downloading === primaryAsset.assetId"
                @click="download(primaryAsset.assetId)"
              />
              <UButton
                v-else-if="
                  siteSettings.resource.downloadsEnabled &&
                  primaryNetdisk?.netdisk?.url
                "
                :to="primaryNetdisk.netdisk.url"
                target="_blank"
                icon="i-tabler-cloud-download"
                label="打开网盘"
                size="lg"
                class="justify-center"
              />
              <UButton
                to="/"
                icon="i-tabler-arrow-left"
                label="返回列表"
                color="neutral"
                variant="outline"
                size="lg"
                class="justify-center"
              />
            </div>
            <div
              class="mt-5 flex flex-wrap items-center gap-3 text-sm text-muted"
            >
              <span class="flex items-center gap-1"
                ><UIcon name="i-tabler-calendar" class="size-4" />更新于
                {{ fmtDate(r.updatedAt) }}</span
              >
              <ShareBar :title="r.title" />
            </div>
          </div>
        </div>
      </div>
    </section>

    <main class="grid gap-8 lg:grid-cols-[minmax(0,760px)_300px]">
      <article
        class="resource-surface-card min-w-0 rounded-xl px-5 py-6 sm:px-7 sm:py-8"
      >
        <section>
          <h2 class="font-display text-xl font-semibold text-highlighted">
            资源说明
          </h2>
          <ContentProse
            v-if="r.description"
            :content="r.description"
            class="mt-5"
          />
          <p v-else class="mt-4 text-sm leading-6 text-muted">暂无详细说明。</p>
        </section>

        <section v-if="tags.length" class="mt-10 border-t border-default pt-6">
          <h2 class="sr-only">标签</h2>
          <div class="flex flex-wrap gap-2">
            <NuxtLink
              v-for="t in tags"
              :key="t.id"
              :to="`/tags/${t.slug}`"
              class="inline-flex items-center rounded-full bg-elevated px-3 py-1 text-sm text-muted transition hover:text-primary"
              >#{{ t.name }}</NuxtLink
            >
          </div>
        </section>

        <section
          v-if="relatedItems.length"
          class="mt-10 border-t border-default pt-6 lg:hidden"
        >
          <h2 class="font-display text-lg font-semibold text-highlighted">
            相关资源
          </h2>
          <div class="mt-4 grid gap-3 sm:grid-cols-2">
            <ResourceCard
              v-for="item in relatedItems"
              :key="item.id"
              :resource="item"
            />
          </div>
        </section>
      </article>

      <aside class="space-y-5 lg:sticky lg:top-24 lg:self-start">
        <section class="resource-surface-card rounded-lg p-4">
          <h2
            class="mb-3 flex items-center gap-2 text-sm font-semibold text-highlighted"
          >
            <UIcon
              name="i-tabler-download"
              class="size-4 text-primary"
            />交付下载
          </h2>
          <div
            v-if="fileDeliveryAssets.length || netdiskItems.length"
            class="space-y-2"
          >
            <div
              v-for="asset in fileDeliveryAssets"
              :key="asset.assetId"
              class="resource-surface-inset rounded-lg p-3"
            >
              <div class="flex min-w-0 items-start gap-3">
                <span
                  class="grid size-9 shrink-0 place-items-center rounded-lg bg-default text-primary"
                >
                  <UIcon :name="assetIcon(asset)" class="size-5" />
                </span>
                <div class="min-w-0 flex-1">
                  <p class="line-clamp-1 text-sm font-medium text-highlighted">
                    {{ asset.label || asset.filename }}
                  </p>
                  <p class="mt-0.5 line-clamp-1 text-xs text-muted">
                    {{ asset.filename }}
                  </p>
                  <p class="mt-1 text-xs text-dimmed">
                    {{ fmtSize(asset.size) }}
                  </p>
                </div>
              </div>
              <UButton
                v-if="siteSettings.resource.downloadsEnabled"
                icon="i-tabler-download"
                label="下载"
                block
                size="sm"
                class="mt-3"
                :loading="downloading === asset.assetId"
                @click="download(asset.assetId)"
              />
            </div>
            <div
              v-for="item in netdiskItems"
              :key="item.id || item.netdisk?.url"
              class="resource-surface-inset rounded-lg p-3"
            >
              <div class="flex min-w-0 items-start gap-3">
                <span
                  class="grid size-9 shrink-0 place-items-center rounded-lg bg-default text-primary"
                >
                  <UIcon name="i-tabler-cloud" class="size-5" />
                </span>
                <div class="min-w-0 flex-1">
                  <p class="line-clamp-1 text-sm font-medium text-highlighted">
                    {{ netdiskTitle(item) }}
                  </p>
                  <p class="mt-0.5 break-all text-xs text-muted">
                    {{ item.netdisk?.url }}
                  </p>
                </div>
              </div>
              <dl class="mt-3 grid gap-2 text-xs sm:grid-cols-2">
                <div
                  v-if="item.netdisk?.accessCode"
                  class="rounded-md bg-default px-2 py-1.5"
                >
                  <dt class="text-muted">访问码</dt>
                  <dd class="mt-0.5 font-mono text-highlighted">
                    {{ item.netdisk.accessCode }}
                  </dd>
                </div>
                <div
                  v-if="item.netdisk?.extractCode"
                  class="rounded-md bg-default px-2 py-1.5"
                >
                  <dt class="text-muted">提取码</dt>
                  <dd class="mt-0.5 font-mono text-highlighted">
                    {{ item.netdisk.extractCode }}
                  </dd>
                </div>
              </dl>
              <p
                v-if="item.netdisk?.note"
                class="mt-2 text-xs leading-5 text-muted"
              >
                {{ item.netdisk.note }}
              </p>
              <UButton
                v-if="siteSettings.resource.downloadsEnabled"
                :to="item.netdisk?.url"
                target="_blank"
                icon="i-tabler-external-link"
                label="打开网盘"
                block
                size="sm"
                class="mt-3"
              />
            </div>
          </div>
          <p v-else class="text-sm text-muted">暂无交付内容</p>
        </section>

        <section class="resource-surface-card rounded-lg p-4">
          <h2
            class="mb-3 flex items-center gap-2 text-sm font-semibold text-highlighted"
          >
            <UIcon
              name="i-tabler-info-circle"
              class="size-4 text-primary"
            />资源信息
          </h2>
          <dl class="space-y-3 text-sm">
            <div class="flex justify-between gap-3">
              <dt class="text-muted">类型</dt>
              <dd class="text-highlighted">{{ typeLabel }}</dd>
            </div>
            <div class="flex justify-between gap-3">
              <dt class="text-muted">创建</dt>
              <dd class="text-highlighted">{{ fmtDate(r.createdAt) }}</dd>
            </div>
            <div class="flex justify-between gap-3">
              <dt class="text-muted">更新</dt>
              <dd class="text-highlighted">{{ fmtDate(r.updatedAt) }}</dd>
            </div>
          </dl>
        </section>

        <section class="resource-surface-card hidden rounded-lg p-4 lg:block">
          <h2
            class="mb-3 flex items-center gap-2 text-sm font-semibold text-highlighted"
          >
            <UIcon name="i-tabler-share-3" class="size-4 text-primary" />分享
          </h2>
          <ShareBar :title="r.title" />
        </section>

        <section
          v-if="relatedItems.length"
          class="resource-surface-card hidden rounded-lg p-4 lg:block"
        >
          <h2
            class="mb-4 flex items-center gap-2 text-sm font-semibold text-highlighted"
          >
            <UIcon
              name="i-tabler-sparkles"
              class="size-4 text-primary"
            />相关资源
          </h2>
          <div class="space-y-4">
            <NuxtLink
              v-for="item in relatedItems"
              :key="item.id"
              :to="`/resources/${item.id}`"
              class="group flex gap-3"
            >
              <div
                class="size-14 shrink-0 overflow-hidden rounded-lg bg-elevated"
              >
                <img
                  v-if="item.coverUrl"
                  :src="item.coverUrl"
                  :alt="item.title"
                  class="size-full object-cover transition duration-500 group-hover:scale-105"
                />
                <div
                  v-else
                  class="grid size-full place-items-center text-primary/45"
                >
                  <UIcon name="i-tabler-package" class="size-5" />
                </div>
              </div>
              <div class="min-w-0">
                <p
                  class="line-clamp-2 text-sm font-medium leading-snug text-highlighted transition group-hover:text-primary"
                >
                  {{ item.title }}
                </p>
                <p class="mt-1 text-xs text-muted">
                  {{ item.downloadCount }} 次下载
                </p>
              </div>
            </NuxtLink>
          </div>
        </section>
      </aside>
    </main>
  </div>
</template>
