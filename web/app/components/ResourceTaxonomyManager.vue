<script setup lang="ts">
import { PageHeader } from "@yueli/ui/dashboard/pattern";
import { useActionFeedback, useMinimumLoading } from "@yueli/ui/feedback";
import { ActionFeedbackButton } from "@yueli/ui/feedback/pattern";
import { createCollectionRouteQueryCodec } from "@yueli/ui/collection";
import { useVueRouterCollectionQuery } from "@yueli/ui/collection/vue-router";
import {
  CollectionDock,
  CollectionPagination,
  CollectionSortDirectionButton,
  CollectionToolbar,
} from "@yueli/ui/collection/pattern";
import { ManageEmpty, SkeletonList } from "@platform/manage/components";
import type { ListTaxonomies, TaxonomyView } from "~/types";

const { kind } = defineProps<{ kind: "category" | "tag" }>();
const { isAdmin } = useAuth();
const { call } = useApi();
const router = useRouter();
const ROOT = "__root__";
type TaxonomySort = "count" | "name" | "slug";
type Direction = "asc" | "desc";
interface TaxonomyQuery {
  q: string;
  sort: TaxonomySort;
  direction: Direction;
  page: number;
  size: number;
}
const pageSizes = [15, 30, 60, 100] as const;
const defaultSort: TaxonomySort = kind === "category" ? "name" : "count";
const defaultDirection: Direction = kind === "category" ? "asc" : "desc";
const { query, replace: replaceQuery } =
  useVueRouterCollectionQuery<TaxonomyQuery>({
    router,
    codec: createCollectionRouteQueryCodec({
      q: { kind: "string", default: "" },
      sort: {
        kind: "enum",
        values: ["count", "name", "slug"],
        default: defaultSort,
      },
      direction: {
        kind: "enum",
        values: ["asc", "desc"],
        default: defaultDirection,
      },
      page: { kind: "positive-integer", default: 1 },
      size: { kind: "positive-integer", values: pageSizes, default: 30 },
    }),
  });
function updateQuery(patch: Partial<TaxonomyQuery>, resetPage = true) {
  void replaceQuery({
    ...query.value,
    ...patch,
    ...(resetPage ? { page: 1 } : {}),
  } as TaxonomyQuery);
}
const q = computed(() => query.value.q);
const sort = computed({
  get: () => query.value.sort,
  set: (value: TaxonomySort) => updateQuery({ sort: value }),
});
const direction = computed({
  get: () => query.value.direction,
  set: (value: Direction) => updateQuery({ direction: value }),
});
const page = computed({
  get: () => query.value.page,
  set: (value: number) => updateQuery({ page: value }, false),
});
const size = computed({
  get: () => query.value.size,
  set: (value: number) => updateQuery({ size: value }),
});
const searchInput = ref(q.value);
let searchTimer: ReturnType<typeof setTimeout> | undefined;
watch(searchInput, (value) => {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(() => updateQuery({ q: value.trim() }), 300);
});
watch(q, (value) => {
  if (searchInput.value !== value) searchInput.value = value;
});
onScopeDispose(() => {
  if (searchTimer) clearTimeout(searchTimer);
});

const mounted = ref(false);
onMounted(() => {
  mounted.value = true;
});
const flat = computed(() => kind === "tag" || Boolean(q.value.trim()));
const { data, pending, refresh, error } = await useAsyncData(
  `resource-taxonomies-${kind}`,
  () =>
    call<ListTaxonomies>("/api/v1/taxonomies", {
      query: {
        taxonomy: kind,
        q: q.value.trim() || undefined,
        sort: sort.value,
        direction: direction.value,
        page: flat.value ? page.value : undefined,
        size: flat.value ? size.value : undefined,
      },
    }),
  {
    server: false,
    watch: [q, sort, direction, page, size],
    default: () => ({
      items: [] as TaxonomyView[],
      total: 0,
      page: 1,
      size: 0,
    }),
  },
);

const items = computed(() => data.value?.items ?? []);
const total = computed(() => data.value?.total ?? items.value.length);
const comparator = (a: TaxonomyView, b: TaxonomyView) => {
  const multiplier = direction.value === "asc" ? 1 : -1;
  if (sort.value === "name")
    return a.name.localeCompare(b.name, "zh-CN") * multiplier;
  if (sort.value === "slug") return a.slug.localeCompare(b.slug) * multiplier;
  return (
    ((a.count || 0) - (b.count || 0) || a.name.localeCompare(b.name, "zh-CN")) *
    multiplier
  );
};
const rows = computed<{ tax: TaxonomyView; depth: number }[]>(() => {
  if (flat.value) {
    return items.value.map((tax) => ({ tax, depth: 0 }));
  }

  const byParent = new Map<string, TaxonomyView[]>();
  for (const item of items.value) {
    const parent = item.parentId || "";
    if (!byParent.has(parent)) byParent.set(parent, []);
    byParent.get(parent)!.push(item);
  }
  const result: { tax: TaxonomyView; depth: number }[] = [];
  const walk = (parent: string, depth: number) => {
    for (const item of [...(byParent.get(parent) || [])].sort(comparator)) {
      result.push({ tax: item, depth });
      walk(item.id, depth + 1);
    }
  };
  walk("", 0);
  return result;
});
const totalPages = computed(() =>
  flat.value ? Math.max(1, Math.ceil(total.value / size.value)) : 1,
);
const pagedRows = computed(() => rows.value);
const showSkeleton = useMinimumLoading(
  computed(() => !mounted.value || pending.value),
);
const sortItems = [
  { label: "按资源数", value: "count" },
  { label: "按名称", value: "name" },
  { label: "按 Slug", value: "slug" },
];
const pageSizeItems = [15, 30, 60, 100].map((value) => ({
  label: `${value}/页`,
  value,
}));

watch(
  totalPages,
  (lastPage) => {
    if (page.value > lastPage) page.value = lastPage;
  },
  { flush: "sync" },
);

const panelOpen = ref(false);
const current = ref<TaxonomyView | null>(null);
const mergeTarget = ref("");
const operationBusy = ref<"" | "merge" | "delete">("");
const operationError = ref("");
const saveError = ref("");
const deleteArmed = ref(false);
const form = reactive({ name: "", slug: "", description: "", parentId: ROOT });
const {
  status: saveStatus,
  pending: markSaving,
  success: markSaved,
  reset: resetSave,
} = useActionFeedback();
const options = ref<TaxonomyView[]>([]);
const optionsLoading = ref(false);
watch(
  [items, flat],
  ([values, isFlat]) => {
    if (!isFlat) options.value = [...values];
  },
  { immediate: true },
);
const optionSource = computed(() =>
  options.value.length ? options.value : items.value,
);

const title = computed(() => (kind === "category" ? "分类" : "标签"));
const headerTitle = computed(() =>
  kind === "category" ? "资源分类" : "资源标签",
);
const headerSubtitle = computed(() =>
  kind === "category"
    ? "维护资源站目录层级，帮助用户按用途浏览资源。"
    : "维护资源站标签，合并重复词，保持搜索和筛选清晰。",
);
const parentItems = computed(() => [
  { label: "顶级分类", value: ROOT },
  ...optionSource.value
    .filter((category) => category.id !== current.value?.id)
    .map((category) => ({ label: category.name, value: category.id })),
]);
const mergeTargets = computed(() =>
  optionSource.value
    .filter((item) => item.id !== current.value?.id)
    .map((item) => ({
      label: `${item.name} · ${item.count || 0} 个资源`,
      value: item.id,
    })),
);

function resetPanelState() {
  resetSave();
  mergeTarget.value = "";
  operationBusy.value = "";
  operationError.value = "";
  saveError.value = "";
  deleteArmed.value = false;
}

async function ensureOptions() {
  if (optionsLoading.value || (!flat.value && options.value.length)) return;
  optionsLoading.value = true;
  try {
    const response = await call<ListTaxonomies>("/api/v1/taxonomies", {
      query: { taxonomy: kind, sort: "name", direction: "asc" },
    });
    options.value = response.items;
  } finally {
    optionsLoading.value = false;
  }
}

function openCreate() {
  resetPanelState();
  current.value = null;
  Object.assign(form, { name: "", slug: "", description: "", parentId: ROOT });
  panelOpen.value = true;
  void ensureOptions();
}

function openEdit(item: TaxonomyView) {
  resetPanelState();
  current.value = item;
  Object.assign(form, {
    name: item.name,
    slug: item.slug,
    description: item.description || "",
    parentId: item.parentId || ROOT,
  });
  panelOpen.value = true;
  void ensureOptions();
}

async function save() {
  if (!form.name.trim()) return;
  markSaving();
  saveError.value = "";
  try {
    const body: Record<string, unknown> = {
      name: form.name.trim(),
      description: form.description.trim(),
    };
    if (current.value) body.slug = form.slug.trim();
    else body.taxonomy = kind;
    if (kind === "category")
      body.parentId = form.parentId === ROOT ? "" : form.parentId;

    const response = current.value
      ? await call<{ taxonomy: TaxonomyView }>(
          `/api/v1/taxonomies/${current.value.id}`,
          { method: "PATCH", body },
        )
      : await call<{ taxonomy: TaxonomyView }>("/api/v1/taxonomies", {
          method: "POST",
          body,
        });
    current.value = response.taxonomy;
    Object.assign(form, {
      name: response.taxonomy.name,
      slug: response.taxonomy.slug,
      description: response.taxonomy.description || "",
      parentId: response.taxonomy.parentId || ROOT,
    });
    options.value = [];
    await refresh();
    markSaved();
  } catch (err: any) {
    resetSave();
    saveError.value =
      err?.data?.message || "保存失败，请检查名称与 slug 后重试";
  }
}

async function mergeCurrent() {
  if (!current.value || !mergeTarget.value || operationBusy.value) return;
  operationBusy.value = "merge";
  operationError.value = "";
  try {
    await call(`/api/v1/taxonomies/${current.value.id}/merge`, {
      method: "POST",
      body: { targetId: mergeTarget.value },
    });
    panelOpen.value = false;
    options.value = [];
    await refresh();
  } catch (err: any) {
    operationError.value = err?.data?.message || "合并失败，请重试";
  } finally {
    operationBusy.value = "";
  }
}

async function deleteCurrent() {
  if (!current.value || operationBusy.value) return;
  operationBusy.value = "delete";
  operationError.value = "";
  try {
    await call(`/api/v1/taxonomies/${current.value.id}`, { method: "DELETE" });
    panelOpen.value = false;
    options.value = [];
    await refresh();
  } catch (err: any) {
    operationError.value = err?.data?.message || "可能仍有子分类或关联资源";
  } finally {
    operationBusy.value = "";
  }
}

function armDelete() {
  deleteArmed.value = true;
}
function cancelDelete() {
  deleteArmed.value = false;
}
</script>

<template>
  <div class="space-y-5">
    <PageHeader :title="headerTitle">
      <template #subtitle>{{ headerSubtitle }}</template>
      <template #actions>
        <UButton
          v-if="isAdmin"
          :icon="kind === 'category' ? 'i-tabler-folder-plus' : 'i-tabler-hash'"
          :label="`新建${title}`"
          @click="openCreate"
        />
      </template>
    </PageHeader>

    <SkeletonList v-if="showSkeleton" :rows="8" />
    <UAlert
      v-else-if="error"
      color="error"
      icon="i-tabler-alert-circle"
      title="加载失败"
      :description="error.message"
    />
    <UAlert
      v-else-if="!isAdmin"
      color="warning"
      icon="i-tabler-shield-lock"
      title="需要管理员权限"
      :description="`${title}治理会影响全站资源目录，请使用管理员账户操作。`"
    />

    <template v-else>
      <CollectionToolbar
        v-model:search="searchInput"
        :search-placeholder="`搜索${title}名称、slug 或描述…`"
        compact-filters
      >
        <template #filters>
          <USelectMenu
            v-model="sort"
            :items="sortItems"
            value-key="value"
            icon="i-tabler-arrows-sort"
            size="sm"
          />
          <CollectionSortDirectionButton v-model="direction" />
        </template>
      </CollectionToolbar>

      <ManageEmpty
        v-if="!rows.length"
        :icon="
          kind === 'category' ? 'i-tabler-folder-off' : 'i-tabler-hash-off'
        "
        :text="q ? `没有匹配的${title}` : `还没有${title}`"
      />

      <section
        v-else
        class="overflow-hidden rounded-xl border border-default bg-default"
      >
        <div class="divide-y divide-default">
          <button
            v-for="row in pagedRows"
            :key="row.tax.id"
            type="button"
            class="group grid w-full grid-cols-[minmax(0,1fr)_2.75rem] items-center gap-3 px-3 py-3 text-left transition hover:bg-elevated/50 focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-primary sm:px-4 lg:grid-cols-[minmax(14rem,1fr)_8rem_2.75rem]"
            @click="openEdit(row.tax)"
          >
            <span
              class="flex min-w-0 items-center gap-3"
              :style="
                kind === 'category'
                  ? { paddingLeft: `${Math.min(row.depth, 4) * 22}px` }
                  : undefined
              "
            >
              <UIcon
                v-if="kind === 'category' && row.depth > 0"
                name="i-tabler-corner-down-right"
                class="size-4 shrink-0 text-dimmed"
              />
              <span
                class="grid size-10 shrink-0 place-items-center rounded-lg bg-elevated text-muted"
              >
                <UIcon
                  :name="
                    kind === 'category' ? 'i-tabler-folder' : 'i-tabler-hash'
                  "
                  class="size-4"
                />
              </span>
              <span class="min-w-0">
                <span
                  class="line-clamp-1 text-sm font-medium text-highlighted"
                  >{{ row.tax.name }}</span
                >
                <span
                  class="mt-0.5 block line-clamp-1 font-mono text-xs text-muted"
                  >/{{ row.tax.slug }}</span
                >
                <span
                  v-if="row.tax.description"
                  class="mt-1 block line-clamp-1 text-xs text-muted"
                  >{{ row.tax.description }}</span
                >
              </span>
            </span>
            <span
              class="col-start-1 pl-13 text-xs text-muted lg:col-start-auto lg:pl-0 lg:text-right"
            >
              <span class="font-semibold text-highlighted">{{
                row.tax.count || 0
              }}</span>
              个资源
            </span>
            <span
              class="row-start-1 col-start-2 grid size-11 place-items-center text-muted lg:row-auto lg:col-start-auto"
              aria-hidden="true"
            >
              <UIcon name="i-tabler-pencil" class="size-4" />
            </span>
          </button>
        </div>
      </section>

      <CollectionDock v-if="pagedRows.length" :label="`${title}统计与分页`">
        <template #selection>
          <span>共 {{ total }} 个{{ title }}</span>
          <span v-if="flat && total > rows.length" class="text-xs text-muted"
            >本页 {{ rows.length }} 个</span
          >
        </template>
        <template #pagination>
          <template v-if="flat">
            <USelect
              v-model="size"
              :items="pageSizeItems"
              value-key="value"
              size="sm"
              class="w-24"
            />
            <CollectionPagination v-model="page" :total-pages="totalPages" />
          </template>
        </template>
      </CollectionDock>
    </template>

    <USlideover
      v-model:open="panelOpen"
      :title="current ? `编辑${title}` : `新建${title}`"
      side="right"
    >
      <template #body>
        <div class="space-y-5">
          <div class="space-y-4">
            <UAlert
              v-if="saveError"
              color="error"
              variant="subtle"
              icon="i-tabler-alert-circle"
              title="保存失败"
              :description="saveError"
            />
            <UFormField label="名称" required>
              <UInput
                v-model="form.name"
                class="w-full"
                :placeholder="
                  kind === 'category' ? '例如：开发工具' : '例如：DevOps'
                "
              />
            </UFormField>
            <UFormField
              v-if="current"
              label="Slug"
              help="URL 标识；改动会影响分类/标签链接。"
            >
              <UInput v-model="form.slug" class="w-full" />
            </UFormField>
            <UFormField v-if="kind === 'category'" label="父分类">
              <USelectMenu
                v-model="form.parentId"
                :items="parentItems"
                value-key="value"
                placeholder="选择父分类"
                :search-input="{ placeholder: '搜索分类…' }"
                :loading="optionsLoading"
                class="w-full"
              />
            </UFormField>
            <UFormField label="描述">
              <UTextarea v-model="form.description" :rows="3" class="w-full" />
            </UFormField>
            <ActionFeedbackButton
              block
              :status="saveStatus"
              idle-label="保存"
              pending-label="保存中"
              success-label="已保存"
              :disabled="!form.name.trim()"
              @click="save"
            />
          </div>

          <template v-if="current">
            <USeparator />
            <section
              aria-labelledby="resource-taxonomy-management"
              class="space-y-4"
            >
              <div>
                <h3
                  id="resource-taxonomy-management"
                  class="text-sm font-medium text-highlighted"
                >
                  管理{{ title }}
                </h3>
                <p class="mt-1 text-xs text-muted">
                  合并会迁移关联资源；删除有关联或子分类时会被后端拒绝。
                </p>
              </div>
              <UAlert
                v-if="operationError"
                color="error"
                variant="subtle"
                icon="i-tabler-alert-circle"
                title="操作失败"
                :description="operationError"
              />
              <div class="rounded-lg border border-default bg-elevated/35 p-3">
                <UFormField
                  :label="`合并到其他${title}`"
                  :help="`当前关联 ${current.count || 0} 个资源`"
                >
                  <div class="grid gap-2 sm:grid-cols-[minmax(0,1fr)_auto]">
                    <USelectMenu
                      v-model="mergeTarget"
                      :items="mergeTargets"
                      value-key="value"
                      placeholder="选择目标…"
                      :search-input="{ placeholder: `搜索目标${title}…` }"
                      :loading="optionsLoading"
                      class="w-full"
                    />
                    <UButton
                      icon="i-tabler-arrows-join"
                      label="合并"
                      color="warning"
                      variant="soft"
                      :disabled="!mergeTarget"
                      :loading="operationBusy === 'merge'"
                      @click="mergeCurrent"
                    />
                  </div>
                </UFormField>
              </div>
              <div class="rounded-lg border border-error/30 bg-error/5 p-3">
                <div
                  v-if="!deleteArmed"
                  class="flex items-start justify-between gap-3"
                >
                  <div>
                    <p class="text-sm font-medium text-highlighted">
                      删除{{ title }}
                    </p>
                    <p class="mt-1 text-xs text-muted">
                      删除不可撤销；有关联或子分类时后端会拒绝。
                    </p>
                  </div>
                  <UButton
                    label="删除"
                    icon="i-tabler-trash"
                    color="error"
                    variant="soft"
                    size="sm"
                    @click="armDelete"
                  />
                </div>
                <div v-else>
                  <p class="text-sm text-highlighted">
                    确定删除「{{ current.name }}」？此操作不可恢复。
                  </p>
                  <div class="mt-3 flex justify-end gap-2">
                    <UButton
                      label="取消"
                      color="neutral"
                      variant="ghost"
                      size="sm"
                      @click="cancelDelete"
                    />
                    <UButton
                      label="确认删除"
                      color="error"
                      size="sm"
                      :loading="operationBusy === 'delete'"
                      @click="deleteCurrent"
                    />
                  </div>
                </div>
              </div>
            </section>
          </template>
        </div>
      </template>
    </USlideover>
  </div>
</template>
