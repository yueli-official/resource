<script setup lang="ts">
import { PageHeader } from "@yueli/ui/dashboard/pattern";
import { ManageTaxonomyChips } from "~/utils/manageComponents";
import {
  CollectionLifecycleTabs,
  CollectionPanel,
  CollectionViewToggle,
} from "@yueli/ui/collection/pattern";
import {
  createCollectionRouteQueryCodec,
  createJsonCollectionQueryPolicy,
  type CollectionControl,
  type CollectionControlValue,
  type CollectionPanelMessages,
  type CollectionWorkflow,
} from "@yueli/ui/collection";
import { useVueCollectionWorkflow } from "@yueli/ui/collection/vue";
import { createVueRouterCollectionQuerySync } from "@yueli/ui/collection/vue-router";
import { abs } from "~/utils/date";
import { useMinimumLoading } from "@yueli/ui/feedback";
import type {
  MyResources,
  ResourceLifecycleCounts,
  ResourceView,
} from "~/types";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "资源管理 · 控制台" });

const { user } = useAuth();
const { can } = useResourceMe();
const canCreate = computed(() => can("resource.item.create"));
const { call } = useApi();
const router = useRouter();
const mounted = ref(false);
const emptyCounts: ResourceLifecycleCounts = {
  all: 0,
  published: 0,
  draft: 0,
  archived: 0,
  issues: 0,
};
type ResourceStatus = "" | "published" | "draft" | "archived" | "issues";
type ResourceSort =
  | "title"
  | "createdAt"
  | "updatedAt"
  | "publishedAt"
  | "viewCount"
  | "downloadCount";
type ResourceDirection = "asc" | "desc";
type ResourceCollectionView = "list" | "grid";
interface ResourceCollectionQuery {
  q: string;
  status: ResourceStatus;
  sort: ResourceSort;
  direction: ResourceDirection;
  page: number;
  size: number;
  view: ResourceCollectionView;
}

const defaultQuery: ResourceCollectionQuery = {
  q: "",
  status: "",
  sort: "updatedAt",
  direction: "desc",
  page: 1,
  size: 24,
  view: "list",
};
const statuses = ["", "published", "draft", "archived", "issues"] as const;
const sorts = [
  "title",
  "createdAt",
  "updatedAt",
  "publishedAt",
  "viewCount",
  "downloadCount",
] as const;
const pageSizes = [12, 24, 48, 96] as const;
const views = ["list", "grid"] as const;
const queryPolicy = createJsonCollectionQueryPolicy<ResourceCollectionQuery>();
const counts = ref<ResourceLifecycleCounts>(emptyCounts);
const searchInput = ref("");

const sync = createVueRouterCollectionQuerySync({
  router,
  codec: createCollectionRouteQueryCodec({
    q: { kind: "string", default: defaultQuery.q, maxLength: 200 },
    status: { kind: "enum", values: statuses, default: defaultQuery.status },
    sort: { kind: "enum", values: sorts, default: defaultQuery.sort },
    direction: {
      kind: "enum",
      values: ["asc", "desc"] as const,
      default: defaultQuery.direction,
    },
    page: { kind: "positive-integer", default: defaultQuery.page },
    size: {
      kind: "positive-integer",
      values: pageSizes,
      default: defaultQuery.size,
    },
    view: { kind: "enum", values: views, default: defaultQuery.view },
  }),
});
const {
  snapshot: collection,
  workflow,
  reload,
} = useVueCollectionWorkflow({
  initialQuery: defaultQuery,
  queryPolicy,
  keyOf: (resource: ResourceView) => resource.id,
  querySync: sync,
  dataQueryKey,
  load,
});

const query = computed(() => collection.value.query);
function updateQuery(
  patch: Partial<ResourceCollectionQuery>,
  resetPage = true,
) {
  workflow.setQuery({
    ...query.value,
    ...patch,
    ...(resetPage ? { page: 1 } : {}),
  });
}
const status = computed({
  get: () => query.value.status,
  set: (value: ResourceStatus) => updateQuery({ status: value }),
});
const q = computed(() => query.value.q);
const sort = computed({
  get: () => query.value.sort,
  set: (value: ResourceSort) => updateQuery({ sort: value }),
});
const direction = computed({
  get: () => query.value.direction,
  set: (value: ResourceDirection) => updateQuery({ direction: value }),
});
const page = computed({
  get: () => query.value.page,
  set: (value: number) => updateQuery({ page: value }, false),
});
const size = computed({
  get: () => query.value.size,
  set: (value: number) => updateQuery({ size: value }),
});
const viewMode = computed({
  get: () => query.value.view,
  set: (value: ResourceCollectionView) => updateQuery({ view: value }, false),
});

let searchTimer: ReturnType<typeof setTimeout> | undefined;
watch(searchInput, (value) => {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(() => updateQuery({ q: value.trim() }), 300);
});

async function load(
  nextQuery: Readonly<ResourceCollectionQuery>,
  activeWorkflow: CollectionWorkflow<
    ResourceView,
    string,
    ResourceCollectionQuery
  >,
) {
  const token = activeWorkflow.beginLoad();
  try {
    const result = await call<MyResources>("/api/v1/resources/mine", {
      query: {
        q: nextQuery.q || undefined,
        status: nextQuery.status || undefined,
        sortBy: nextQuery.sort,
        sortOrder: nextQuery.direction,
        page: nextQuery.page,
        size: nextQuery.size,
      },
    });
    const lastPage = Math.max(1, Math.ceil(result.total / nextQuery.size));
    if (nextQuery.page > lastPage) {
      activeWorkflow.setQuery({ ...nextQuery, page: lastPage });
      return;
    }
    if (
      activeWorkflow.resolveLoad(token, {
        items: result.items,
        total: result.total,
      })
    )
      counts.value = result.counts;
  } catch {
    activeWorkflow.rejectLoad(token, {
      key: "resource.collection.load_failed",
    });
  }
}

function dataQueryKey(nextQuery: Readonly<ResourceCollectionQuery>) {
  const { view: _view, ...dataQuery } = nextQuery;
  return JSON.stringify(dataQuery);
}

searchInput.value = collection.value.query.q;
watch(q, (value) => {
  if (searchInput.value !== value) searchInput.value = value;
});
onMounted(() => {
  mounted.value = true;
});
onScopeDispose(() => {
  if (searchTimer) clearTimeout(searchTimer);
});

const resources = computed(() => collection.value.items);
const total = computed(() => collection.value.total);
const pending = computed(
  () =>
    collection.value.loadState === "loading" ||
    collection.value.loadState === "refreshing",
);
const error = computed(() => collection.value.issue);
const showSkeleton = useMinimumLoading(
  computed(() => !mounted.value || pending.value),
);

const types = [
  { label: "软件 / 工具", value: "software" },
  { label: "设计素材 / 模板", value: "design" },
  { label: "脚本", value: "script" },
  { label: "其它", value: "default" },
];
const statusTabs = computed(() => [
  { key: "", label: "全部", count: counts.value.all },
  { key: "published", label: "已发布", count: counts.value.published },
  { key: "draft", label: "草稿", count: counts.value.draft },
  { key: "archived", label: "归档", count: counts.value.archived },
  { key: "issues", label: "待完善", count: counts.value.issues },
]);
const sortItems = [
  { label: "按名称", value: "title" },
  { label: "按创建日期", value: "createdAt" },
  { label: "按更新日期", value: "updatedAt" },
  { label: "按发布时间", value: "publishedAt" },
  { label: "按浏览量", value: "viewCount" },
  { label: "按下载量", value: "downloadCount" },
];
const collectionControls = computed<CollectionControl[]>(() => [
  {
    kind: "select",
    id: "sort",
    label: "资源排序",
    value: sort.value,
    options: sortItems,
    icon: "i-tabler-arrows-sort",
    class: "w-36",
  },
  {
    kind: "direction",
    id: "direction",
    label: "排序方向",
    value: direction.value,
    ascendingLabel: "切换为倒序",
    descendingLabel: "切换为正序",
  },
]);
const collectionMessages: CollectionPanelMessages = {
  searchPlaceholder: "搜索标题、摘要或描述…",
  searchAction: "搜索",
  filtersAction: "筛选",
  activeFilters: (count) => `筛选（${count}）`,
  clearFilters: "清除筛选",
  selectPage: "选择当前页资源",
  selectItem: (label) => `选择资源：${label}`,
  bulkRegion: "资源批量操作",
  selected: (count) => `已选择 ${count} 个资源`,
  selectAllResults: "选择全部结果",
  clearSelection: "取消选择",
  emptyTitle: "没有匹配的资源",
  emptyDescription: "请调整搜索条件或状态后重试。",
  errorTitle: "资源加载失败",
  retry: "重新加载",
  showing: (first, last, count) => `显示 ${first}–${last}，共 ${count} 个资源`,
  pageSize: "每页",
  pageSizeControl: "每页资源数量",
  pageSizeOption: (value) => `${value} 个`,
};
function changeCollectionControl(id: string, value: CollectionControlValue) {
  if (id === "sort" && sorts.includes(value as ResourceSort))
    sort.value = value as ResourceSort;
  if (id === "direction" && (value === "asc" || value === "desc"))
    direction.value = value;
}
function submitCollectionSearch(value: string) {
  if (searchTimer) clearTimeout(searchTimer);
  searchInput.value = value;
  updateQuery({ q: value.trim() });
}
const resourceKey = (resource: ResourceView) => resource.id;
const resourceLabel = (resource: ResourceView) => resource.title;

function typeLabel(value: string) {
  return types.find((item) => item.value === value)?.label || value;
}

function resourceMeta(resource: ResourceView) {
  return [
    resource.createdAt ? `创建 ${abs(resource.createdAt)}` : "",
    resource.updatedAt ? `更新 ${abs(resource.updatedAt)}` : "",
  ]
    .filter(Boolean)
    .join(" · ");
}

const selectedIds = computed(() =>
  collection.value.selection.mode === "keys"
    ? collection.value.selection.keys
    : [],
);
const selectionCount = computed(() => collection.value.selection.count);
const isPageSelected = computed(() => collection.value.isPageSelected);
const isPageIndeterminate = computed(
  () => collection.value.isPageIndeterminate,
);
const isSelected = (id: string) => workflow.isSelected(id);
const toggleOne = (id: string, selected?: boolean) => {
  if (selected === undefined || selected !== workflow.isSelected(id))
    workflow.toggleKey(id);
};
const togglePage = (selected?: boolean | "indeterminate") =>
  workflow.togglePage(selected === true);
const clearSelection = () => workflow.clearSelection();
function replaceSelection(ids: readonly string[]) {
  workflow.clearSelection();
  for (const id of ids) workflow.toggleKey(id);
}

type BatchFailure = { id: string; code: string; message: string };
type BatchResult = {
  changed: number;
  failures: BatchFailure[];
  interrupted?: boolean;
  message?: string;
};
const batchItems = [
  { label: "发布", value: "publish" },
  { label: "转为草稿", value: "draft" },
  { label: "归档", value: "archive" },
];
const batchAction = ref<string>();
const batchBusy = ref(false);
const batchResult = ref<BatchResult>();

async function applyBatch() {
  if (!batchAction.value || !selectedIds.value.length) return;
  const requestIds = [...selectedIds.value];
  batchBusy.value = true;
  batchResult.value = undefined;
  try {
    const result = await call<BatchResult>("/api/v1/resources/mine/batch", {
      method: "POST",
      body: { ids: requestIds, action: batchAction.value },
    });
    batchResult.value = result;
    const failedIds = result.failures.map((item) => item.id);
    if (failedIds.length) replaceSelection(failedIds);
    else clearSelection();
    batchAction.value = undefined;
    await reload();
  } catch (batchError) {
    replaceSelection(requestIds);
    batchResult.value = {
      changed: 0,
      failures: [],
      interrupted: true,
      message: resourceFailureMessage(
        batchError,
        "批量请求中断，已保留选择，请核对当前状态后重试。",
      ),
    };
    await reload();
  } finally {
    batchBusy.value = false;
  }
}

const quickEditTarget = ref<ResourceView>();
const showQuickEdit = ref(false);
function openQuickEdit(resource: ResourceView) {
  quickEditTarget.value = resource;
  showQuickEdit.value = true;
}
async function onQuickEditSaved(resource: ResourceView) {
  quickEditTarget.value = resource;
  await reload();
}

const showCreate = ref(false);
const form = reactive({ title: "", type: "software", summary: "" });
const creating = ref(false);
const createFailureFeedback = ref<ReturnType<typeof resourceFailureFeedback>>();

function openCreateModal() {
  createFailureFeedback.value = undefined;
  showCreate.value = true;
}

async function create() {
  if (!form.title.trim() || creating.value) return;
  creating.value = true;
  createFailureFeedback.value = undefined;
  try {
    const response = await call<{ resource: ResourceView }>(
      "/api/v1/resources",
      {
        method: "POST",
        body: { title: form.title, type: form.type, summary: form.summary },
      },
    );
    showCreate.value = false;
    await navigateTo(`/manage/${response.resource.id}`);
  } catch (createFailure) {
    createFailureFeedback.value = resourceFailureFeedback(
      createFailure,
      "创建失败，请检查输入后重试。",
      { "/title": "title", "/type": "type", "/summary": "summary" },
    );
  } finally {
    creating.value = false;
  }
}
</script>

<template>
  <div class="space-y-6">
    <PageHeader title="资源管理">
      <template #subtitle>
        已登录：<span class="text-default">{{
          user?.name || user?.email
        }}</span>
        · 管理资源、状态、封面和下载文件。
      </template>
      <template #actions>
        <UButton
          v-if="canCreate"
          icon="i-tabler-plus"
          label="新建资源"
          @click="openCreateModal"
        />
      </template>
    </PageHeader>

    <div class="space-y-6" :inert="batchBusy" :aria-busy="batchBusy">
      <CollectionLifecycleTabs v-model="status" :items="statusTabs" />

      <UAlert
        v-if="batchResult"
        :color="
          batchResult.interrupted || batchResult.failures.length
            ? 'warning'
            : 'success'
        "
        variant="subtle"
        :icon="
          batchResult.interrupted || batchResult.failures.length
            ? 'i-tabler-alert-triangle'
            : 'i-tabler-circle-check'
        "
        :title="
          batchResult.interrupted
            ? '批量请求中断'
            : `已处理 ${batchResult.changed} 个资源`
        "
        :description="
          batchResult.interrupted
            ? batchResult.message
            : batchResult.failures.length
              ? `${batchResult.failures.length} 个资源仍未完成。`
              : '批量操作已完成。'
        "
        :actions="
          batchResult.failures[0]
            ? [
                {
                  label: '查看首个失败项',
                  to: `/manage/${batchResult.failures[0].id}`,
                  color: 'warning',
                  variant: 'link',
                },
              ]
            : undefined
        "
        close
        @update:open="batchResult = undefined"
      />

      <CollectionPanel
        v-model:search="searchInput"
        :items="resources"
        :item-key="resourceKey"
        :item-label="resourceLabel"
        :controls="collectionControls"
        :messages="collectionMessages"
        :state="error ? 'error' : showSkeleton ? 'loading' : 'ready'"
        :error-message="error?.key"
        :total="total"
        :page="page"
        :page-size="size"
        :page-sizes="pageSizes"
        selectable
        :selection-count="selectionCount"
        :page-selected="isPageSelected"
        :page-indeterminate="isPageIndeterminate"
        :is-selected="isSelected"
        :layout="viewMode === 'grid' ? 'grid' : 'rows'"
        label="资源列表"
        @search="submitCollectionSearch"
        @control-change="changeCollectionControl"
        @retry="reload"
        @toggle-page="togglePage"
        @toggle-item="toggleOne"
        @clear-selection="clearSelection"
        @page-change="
          (value) => {
            page = value;
          }
        "
        @page-size-change="
          (value) => {
            size = value;
          }
        "
      >
        <template #view>
          <CollectionViewToggle
            v-model="viewMode"
            :items="[
              { key: 'list', label: '列表', icon: 'i-tabler-list' },
              { key: 'grid', label: '网格', icon: 'i-tabler-layout-grid' },
            ]"
          />
        </template>

        <template #columns>
          <div
            class="grid grid-cols-[minmax(0,1fr)_11rem_5rem] items-center gap-3"
          >
            <span>名称、类型与标签</span>
            <span class="text-right">数据与更新时间</span>
            <span class="text-right">操作</span>
          </div>
        </template>

        <template #bulk-actions>
          <USelect
            v-model="batchAction"
            :items="batchItems"
            value-key="value"
            placeholder="批量操作"
            size="xs"
            class="w-28"
            :disabled="batchBusy"
            aria-label="批量操作"
          />
          <UButton
            size="xs"
            color="primary"
            variant="soft"
            :disabled="!batchAction"
            :loading="batchBusy"
            @click="applyBatch"
            >应用</UButton
          >
        </template>

        <template #item="{ item: resource }">
          <div
            v-if="viewMode === 'list'"
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_11rem_5rem] items-center gap-3"
          >
            <div class="flex min-w-0 items-center gap-3">
              <div
                class="size-16 shrink-0 overflow-hidden rounded-md border border-default bg-elevated"
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
                  <UIcon name="i-tabler-package" class="size-6" />
                </div>
              </div>
              <div class="min-w-0">
                <NuxtLink
                  :to="`/manage/${resource.id}`"
                  class="block truncate text-sm font-semibold text-highlighted hover:text-primary"
                  >{{ resource.title }}</NuxtLink
                >
                <p class="mt-0.5 truncate text-xs text-muted">
                  {{ resource.summary || "未填写摘要" }}
                </p>
                <div class="mt-1.5 flex min-h-5 flex-wrap gap-1">
                  <UBadge
                    :label="typeLabel(resource.type)"
                    color="primary"
                    variant="soft"
                    size="sm"
                  />
                  <ManageTaxonomyChips
                    :items="
                      resource.tags.map((tag) => ({
                        key: tag,
                        label: tag,
                        kind: 'tag',
                      }))
                    "
                  />
                </div>
                <p
                  v-if="resource.issueCount"
                  class="mt-1 inline-flex items-center gap-1 text-xs text-warning"
                >
                  <UIcon name="i-tabler-alert-circle" class="size-3.5" />待完善
                </p>
              </div>
            </div>
            <div class="min-w-0 text-right text-xs">
              <p class="text-sm font-medium text-highlighted">
                {{ resource.downloadCount }} 次下载
              </p>
              <p class="mt-0.5 truncate font-mono text-muted">
                /{{ resource.slug }}
              </p>
              <ClientOnly
                ><p class="mt-1 truncate text-dimmed">
                  {{ resourceMeta(resource) }}
                </p>
                <template #fallback
                  ><p class="mt-1 text-dimmed">…</p></template
                ></ClientOnly
              >
            </div>
            <div class="flex justify-end gap-1">
              <UTooltip text="快速编辑"
                ><UButton
                  icon="i-tabler-pencil"
                  color="neutral"
                  variant="ghost"
                  size="xs"
                  square
                  :aria-label="`快速编辑：${resource.title}`"
                  @click="openQuickEdit(resource)"
              /></UTooltip>
              <UTooltip text="完整编辑"
                ><UButton
                  :to="`/manage/${resource.id}`"
                  icon="i-tabler-file-pencil"
                  color="neutral"
                  variant="ghost"
                  size="xs"
                  square
                  :aria-label="`完整编辑：${resource.title}`"
              /></UTooltip>
            </div>
          </div>

          <div v-else class="group -m-4 overflow-hidden rounded-lg">
            <NuxtLink
              :to="`/manage/${resource.id}`"
              class="block focus:outline-none"
              :aria-label="`查看资源：${resource.title}`"
            >
              <div
                class="relative h-[clamp(7.5rem,13vw,9.5rem)] overflow-hidden border-b border-default bg-elevated"
              >
                <img
                  v-if="resource.coverUrl"
                  :src="resource.coverUrl"
                  :alt="resource.title"
                  class="size-full object-cover transition duration-300 group-hover:scale-105"
                />
                <div
                  v-else
                  class="grid size-full place-items-center bg-primary/10 text-primary"
                >
                  <UIcon name="i-tabler-package" class="size-8" />
                </div>
              </div>
              <div class="min-w-0 p-3">
                <h2 class="truncate text-sm font-semibold text-highlighted">
                  {{ resource.title }}
                </h2>
                <p class="mt-1 truncate text-xs text-muted">
                  {{ resource.summary || "未填写摘要" }}
                </p>
                <p
                  v-if="resource.issueCount"
                  class="mt-1 inline-flex items-center gap-1 text-xs text-warning"
                >
                  <UIcon name="i-tabler-alert-circle" class="size-3.5" />待完善
                </p>
                <div class="mt-2 flex min-h-5 flex-wrap gap-1">
                  <UBadge
                    :label="typeLabel(resource.type)"
                    color="primary"
                    variant="soft"
                    size="sm"
                  />
                  <ManageTaxonomyChips
                    :items="
                      resource.tags.map((tag) => ({
                        key: tag,
                        label: tag,
                        kind: 'tag',
                      }))
                    "
                  />
                </div>
                <div
                  class="mt-3 flex items-end justify-between gap-2 border-t border-default pt-2.5"
                >
                  <div>
                    <p class="text-xs text-muted">下载</p>
                    <p class="text-sm font-semibold text-highlighted">
                      {{ resource.downloadCount }}
                    </p>
                  </div>
                  <ClientOnly
                    ><p class="text-xs text-dimmed">
                      {{ resource.updatedAt ? abs(resource.updatedAt) : "-" }}
                    </p>
                    <template #fallback
                      ><span class="text-xs text-dimmed">…</span></template
                    ></ClientOnly
                  >
                </div>
              </div>
            </NuxtLink>
            <UTooltip text="快速编辑"
              ><UButton
                icon="i-tabler-pencil"
                color="neutral"
                variant="solid"
                size="xs"
                square
                class="absolute right-4 top-4 z-10"
                :aria-label="`快速编辑：${resource.title}`"
                @click.stop="openQuickEdit(resource)"
            /></UTooltip>
          </div>
        </template>
      </CollectionPanel>
    </div>

    <ResourceQuickEditModal
      v-model:open="showQuickEdit"
      :resource="quickEditTarget"
      :types="types"
      @saved="onQuickEditSaved"
    />

    <UModal
      v-model:open="showCreate"
      title="新建资源（草稿）"
      :ui="{ footer: 'justify-end' }"
    >
      <template #body>
        <div class="space-y-4">
          <UAlert
            v-if="createFailureFeedback"
            title="暂时无法创建"
            :description="resourceFailureDescription(createFailureFeedback)"
            icon="i-tabler-alert-circle"
            color="error"
            variant="soft"
          />
          <details v-if="createFailureFeedback" class="text-xs text-muted">
            <summary class="cursor-pointer">技术详情</summary>
            <code class="select-all">{{ resourceFailureTechnical(createFailureFeedback) }}</code>
          </details>
          <UFormField label="标题" required :error="createFailureFeedback?.fieldErrors.title?.[0]">
            <UInput
              v-model="form.title"
              placeholder="例如：My CLI Tool"
              class="w-full"
              autofocus
            />
          </UFormField>
          <UFormField label="类型" :error="createFailureFeedback?.fieldErrors.type?.[0]">
            <USelect
              v-model="form.type"
              :items="types"
              value-key="value"
              class="w-full"
            />
          </UFormField>
          <UFormField label="简介" :error="createFailureFeedback?.fieldErrors.summary?.[0]">
            <UTextarea
              v-model="form.summary"
              :rows="3"
              class="w-full"
              placeholder="一句话介绍这个资源"
            />
          </UFormField>
        </div>
      </template>
      <template #footer="{ close }">
        <UButton
          label="取消"
          color="neutral"
          variant="outline"
          :disabled="creating"
          @click="close"
        />
        <UButton
          label="创建并编辑"
          icon="i-tabler-arrow-right"
          trailing
          :loading="creating"
          :disabled="!form.title.trim()"
          @click="create"
        />
      </template>
    </UModal>
  </div>
</template>
