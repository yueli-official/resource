<script setup lang="ts">
import { AdminOverview, AdminMetricCard } from "@yueli/ui/admin";
import { PageHeader } from "@yueli/ui/dashboard/pattern";
import { useMinimumLoading } from "@yueli/ui/feedback";
import { abs } from "~/utils/date";
import type { MyResources, ResourceLifecycleCounts } from "~/types";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "控制台 · 资源后台" });

const requestFetch = useRequestFetch();
const emptyCounts: ResourceLifecycleCounts = {
  all: 0,
  published: 0,
  draft: 0,
  archived: 0,
  issues: 0,
};
const { data, pending, error } = await useAsyncData(
  "resource-dashboard",
  () =>
    requestFetch<MyResources>("/api/v1/resources/mine", {
      query: { page: 1, size: 6, sortBy: "updatedAt", sortOrder: "desc" },
    }),
  {
    default: () => ({
      items: [],
      total: 0,
      page: 1,
      size: 6,
      counts: emptyCounts,
    }),
  },
);
const showSkeleton = useMinimumLoading(pending, { delayMs: 120 });
const resources = computed(() => data.value?.items ?? []);
const counts = computed(() => data.value?.counts ?? emptyCounts);
const metrics = computed(() => [
  {
    label: "全部资源",
    value: counts.value.all,
    icon: "i-tabler-package",
    to: "/manage",
  },
  {
    label: "已发布",
    value: counts.value.published,
    icon: "i-tabler-circle-check",
    to: "/manage?status=published",
  },
  {
    label: "草稿",
    value: counts.value.draft,
    icon: "i-tabler-pencil",
    to: "/manage?status=draft",
  },
  {
    label: "待完善",
    value: counts.value.issues,
    icon: "i-tabler-alert-circle",
    to: "/manage?status=issues",
  },
]);
const needsAttention = computed(
  () => counts.value.issues > 0 || counts.value.draft > 0,
);
</script>

<template>
  <div class="space-y-5">
    <PageHeader title="控制台" icon="i-tabler-dashboard">
      <template #tools><AdminOverview>
        <template #artwork><ManageOverviewArtwork /></template>
        <div v-if="showSkeleton" data-admin-metrics><USkeleton v-for="n in 4" :key="n" class="h-24 rounded-xl" /></div>
        <div v-else data-admin-metrics><AdminMetricCard v-for="metric in metrics" :key="metric.label" v-bind="metric" /></div>
      </AdminOverview></template>
      <template #actions>
        <UButton to="/manage" icon="i-tabler-package" label="管理资源" />
      </template>
    </PageHeader>

    <UAlert
      v-if="error"
      color="error"
      variant="subtle"
      icon="i-tabler-alert-circle"
      title="资源数据加载失败"
      description="请刷新后重试。"
    />



    <UCard v-if="!showSkeleton && needsAttention" class="yueli-card">
      <template #header>
        <h2 class="text-sm font-semibold text-highlighted">需要处理</h2>
      </template>
      <div class="grid gap-3 sm:grid-cols-2">
        <NuxtLink
          v-if="counts.issues"
          to="/manage?status=issues"
          class="group flex min-h-16 items-center gap-3 rounded-lg bg-elevated/50 px-3 py-2.5 transition-colors hover:bg-elevated"
        >
          <span class="grid size-9 shrink-0 place-items-center rounded-lg bg-warning/10 text-warning">
            <UIcon name="i-tabler-alert-circle" class="size-5" />
          </span>
          <span class="min-w-0 flex-1 text-sm font-medium text-highlighted">
            {{ counts.issues }} 个资源待完善
          </span>
          <UIcon name="i-tabler-chevron-right" class="size-4 text-dimmed" />
        </NuxtLink>
        <NuxtLink
          v-if="counts.draft"
          to="/manage?status=draft"
          class="group flex min-h-16 items-center gap-3 rounded-lg bg-elevated/50 px-3 py-2.5 transition-colors hover:bg-elevated"
        >
          <span class="grid size-9 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
            <UIcon name="i-tabler-pencil" class="size-5" />
          </span>
          <span class="min-w-0 flex-1 text-sm font-medium text-highlighted">
            {{ counts.draft }} 个草稿资源
          </span>
          <UIcon name="i-tabler-chevron-right" class="size-4 text-dimmed" />
        </NuxtLink>
      </div>
    </UCard>

    <UCard class="yueli-card" :ui="{ body: 'p-0 sm:p-0' }">
      <template #header>
        <div class="flex items-center justify-between gap-3">
          <h2 class="text-sm font-semibold text-highlighted">最近更新</h2>
          <UButton
            to="/manage"
            label="全部资源"
            trailing-icon="i-tabler-chevron-right"
            color="neutral"
            variant="ghost"
            size="sm"
          />
        </div>
      </template>
      <div v-if="showSkeleton" class="divide-y divide-default p-4">
        <div v-for="item in 6" :key="item" class="flex min-h-16 items-center gap-3">
          <USkeleton class="size-10 rounded-lg" />
          <div class="min-w-0 flex-1 space-y-2">
            <USkeleton class="h-3 w-36" />
            <USkeleton class="h-3 w-24" />
          </div>
        </div>
      </div>
      <div v-else-if="resources.length" class="divide-y divide-default">
        <NuxtLink
          v-for="resource in resources"
          :key="resource.id"
          :to="`/manage/${resource.id}`"
          class="group grid min-h-16 grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-3 px-4 py-3 transition-colors hover:bg-elevated/50"
        >
          <span class="grid size-10 place-items-center overflow-hidden rounded-lg bg-elevated">
            <img
              v-if="resource.coverUrl"
              :src="resource.coverUrl"
              :alt="resource.title"
              class="size-full object-cover"
            />
            <UIcon v-else name="i-tabler-package" class="size-5 text-muted" />
          </span>
          <span class="min-w-0">
            <span class="block truncate text-sm font-medium text-highlighted group-hover:text-primary">
              {{ resource.title || "(无标题)" }}
            </span>
            <span class="block truncate text-xs text-muted">
              {{ resource.type }} ·
              {{ resource.status === "published" ? "已发布" : resource.status === "archived" ? "已归档" : "草稿" }}
            </span>
          </span>
          <span class="hidden text-xs text-muted sm:block">{{ abs(resource.updatedAt) }}</span>
        </NuxtLink>
      </div>
      <div v-else class="p-10 text-center text-sm text-muted">
        还没有资源
      </div>
    </UCard>
  </div>
</template>
