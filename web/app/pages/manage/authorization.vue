<script setup lang="ts">
import { PageHeader } from "@yueli/ui/dashboard/pattern";

interface RoleView {
  key: string;
  displayName: string;
  kind: string;
  protected: boolean;
  capabilities: string[];
  assignmentSources: string[];
}
interface ApplicationView {
  id: string;
  subject: string;
  role: string;
  reason: string;
}
interface GrantView {
  id: string;
  subject: string;
  role: string;
  source: string;
}
interface ConsoleView {
  activeRevision: number;
  policy: { number: number; state: string };
  roles: RoleView[];
  automaticRules: { key: string; enabled: boolean }[];
  applications: ApplicationView[];
  grants: GrantView[];
  capabilities: { key: string; displayName: string }[];
}

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "权限与申请" });

const { call } = useApi();
const { isAdministrator } = useResourceMe();
const toast = useToast();
const busy = ref(false);
const createRoleOpen = ref(false);
const roleForm = reactive({
  key: "",
  displayName: "",
  capabilities: [] as string[],
  assignmentSources: ["application", "invitation", "direct"] as string[],
});
const grantForm = reactive({ subject: "", role: "contributor" });
const { data, pending, error, refresh } = await useAsyncData(
  "resource-authorization-console",
  () => call<ConsoleView>("/api/v1/authorization/manage/console"),
  { server: false },
);
const state = computed(() => data.value);
const draft = computed(() => state.value?.policy.state === "draft");

async function mutate(task: () => Promise<unknown>, _success: string) {
  if (busy.value) return;
  busy.value = true;
  try {
    const result = await task();
    if (result === false) return;
    await refresh();
  } catch (failure) {
    const message = failure instanceof Error
      ? failure.message
      : (failure as { data?: { message?: string } }).data?.message;
    toast.add({
      title: "操作失败",
      description: message || "请刷新后重试。",
      color: "error",
      icon: "i-tabler-alert-circle",
    });
  } finally {
    busy.value = false;
  }
}

function createDraft() {
  if (!state.value?.activeRevision) return;
  return mutate(
    () => call("/api/v1/authorization/manage/policies/drafts", {
      method: "POST",
      body: { expectedActiveRevision: state.value!.activeRevision },
    }),
    "策略草稿已创建",
  );
}

function toggleRoleCapability(role: RoleView, capability: string) {
  if (!draft.value || role.protected) return;
  const capabilities = role.capabilities.includes(capability)
    ? role.capabilities.filter((item) => item !== capability)
    : [...role.capabilities, capability];
  return mutate(
    () => call(
      `/api/v1/authorization/manage/policies/${state.value!.policy.number}/roles/${role.key}/capabilities`,
      { method: "PUT", body: { capabilities } },
    ),
    "角色能力已更新到草稿",
  );
}

function toggleAutomatic(enabled: boolean) {
  const rule = state.value?.automaticRules[0];
  if (!draft.value || !rule) return;
  return mutate(
    () => call(
      `/api/v1/authorization/manage/policies/${state.value!.policy.number}/automatic/${rule.key}`,
      { method: "PUT", body: { enabled } },
    ),
    enabled ? "已启用注册自动授权" : "已关闭注册自动授权",
  );
}

async function validateAndActivate() {
  const current = state.value;
  if (!draft.value || !current) return;
  await mutate(async () => {
    const validation = await call<{ valid: boolean; violations: string[] }>(
      `/api/v1/authorization/manage/policies/${current.policy.number}/validate`,
      { method: "POST" },
    );
    if (!validation.valid) throw new Error(validation.violations.join("；"));
    const impact = await call<{ removedBindings: number }>(
      `/api/v1/authorization/manage/policies/${current.policy.number}/preview`,
      { method: "POST" },
    );
    if (
      impact.removedBindings > 0
      && !window.confirm(`本次发布会移除 ${impact.removedBindings} 项能力绑定，是否继续？`)
    ) return false;
    await call(
      `/api/v1/authorization/manage/policies/${current.policy.number}/activate`,
      {
        method: "POST",
        body: { expectedActiveRevision: current.activeRevision },
      },
    );
  }, "权限策略已发布");
}

function review(application: ApplicationView, decision: "approve" | "reject") {
  return mutate(
    () => call(
      `/api/v1/authorization/manage/applications/${application.id}/review`,
      {
        method: "POST",
        body: {
          decision,
          reason: decision === "approve" ? "管理员批准" : "管理员拒绝",
        },
      },
    ),
    decision === "approve" ? "申请已批准" : "申请已拒绝",
  );
}

function createRole() {
  if (!draft.value || !state.value) return;
  return mutate(async () => {
    await call(
      `/api/v1/authorization/manage/policies/${state.value!.policy.number}/roles`,
      { method: "POST", body: roleForm },
    );
    createRoleOpen.value = false;
    Object.assign(roleForm, {
      key: "",
      displayName: "",
      capabilities: [],
      assignmentSources: ["application", "invitation", "direct"],
    });
  }, "自定义角色已创建");
}

function grantRole() {
  if (!grantForm.subject.trim() || !grantForm.role) return;
  return mutate(async () => {
    await call("/api/v1/authorization/manage/grants", {
      method: "POST",
      body: { subject: grantForm.subject.trim(), role: grantForm.role },
    });
    grantForm.subject = "";
  }, "角色已直接授予");
}

function revokeGrant(grant: GrantView) {
  if (!window.confirm(`确定撤销 ${grant.subject} 的 ${grant.role} 角色吗？`)) return;
  return mutate(
    () => call(`/api/v1/authorization/manage/grants/${grant.id}`, { method: "DELETE" }),
    "角色授权已撤销",
  );
}
</script>

<template>
  <div
    id="authorization"
    class="mx-auto w-full max-w-screen-2xl space-y-4"
  >
    <PageHeader
      title="权限与申请"
      description="配置本站角色能力、贡献者申请与自动授权；用户中心只提供登录身份。"
    >
      <template #actions>
        <UButton
          v-if="state && !draft"
          label="创建策略草稿"
          icon="i-tabler-file-plus"
          :loading="busy"
          @click="createDraft"
        />
        <UButton
          v-else-if="draft"
          label="验证并发布"
          icon="i-tabler-rocket"
          :loading="busy"
          @click="validateAndActivate"
        />
      </template>
    </PageHeader>

    <ManageClientBoundary :rows="8">
      <UAlert
        v-if="!isAdministrator"
        color="error"
        icon="i-tabler-lock"
        title="只有管理员可以管理本站权限"
        description="资源站权限独立存储，不继承用户中心或其他站点的管理员角色。"
      />
      <SkeletonList v-else-if="pending" :rows="8" />
      <UAlert
        v-else-if="error || !state"
        color="error"
        icon="i-tabler-alert-circle"
        title="权限配置加载失败"
        description="请检查 Resource API 与本站数据库状态。"
      />

      <template v-else>
        <section class="rounded-xl border border-default bg-default p-4 sm:p-5">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h2 class="text-sm font-semibold text-highlighted">策略发布</h2>
              <p class="mt-1 text-sm text-muted">
                生效修订 {{ state.activeRevision }}，当前查看修订 {{ state.policy.number }}。
              </p>
            </div>
            <UBadge
              :label="draft ? '草稿' : '已生效'"
              :color="draft ? 'warning' : 'success'"
              variant="subtle"
            />
          </div>
          <p class="mt-3 text-xs leading-5 text-muted">
            角色与自动规则先进入草稿，验证影响后一次发布，不会逐项改变线上权限。
          </p>
        </section>

        <section class="space-y-3">
          <div>
            <h2 class="text-sm font-semibold text-highlighted">成员授权</h2>
            <p class="mt-1 text-xs text-muted">
              可直接授予管理员、贡献者或自定义角色；申请审批与自动授权也会出现在这里。
            </p>
          </div>
          <div class="grid gap-3 rounded-xl border border-default bg-default p-4 sm:grid-cols-[minmax(0,1fr)_14rem_auto]">
            <UFormField label="用户标识">
              <UInput v-model="grantForm.subject" placeholder="Identity subject" class="w-full" />
            </UFormField>
            <UFormField label="角色">
              <USelect
                v-model="grantForm.role"
                value-key="value"
                :items="state.roles
                  .filter((role) => role.kind !== 'custom' || role.capabilities.length)
                  .map((role) => ({ label: role.displayName, value: role.key }))"
                class="w-full"
              />
            </UFormField>
            <div class="flex items-end">
              <UButton
                label="直接授予"
                icon="i-tabler-user-plus"
                :disabled="!grantForm.subject.trim() || busy"
                @click="grantRole"
              />
            </div>
          </div>
          <ManageEmpty
            v-if="!state.grants.length"
            icon="i-tabler-users"
            text="当前没有角色授权"
          />
          <div v-else class="divide-y divide-default overflow-hidden rounded-xl border border-default bg-default">
            <article
              v-for="grant in state.grants"
              :key="grant.id"
              class="grid gap-3 p-4 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center"
            >
              <div class="min-w-0">
                <p class="truncate text-sm font-medium text-highlighted">{{ grant.subject }}</p>
                <p class="mt-1 text-xs text-muted">
                  {{ state.roles.find((role) => role.key === grant.role)?.displayName || grant.role }}
                  · {{ grant.source }}
                </p>
              </div>
              <UButton
                label="撤销"
                color="error"
                variant="ghost"
                :disabled="busy"
                @click="revokeGrant(grant)"
              />
            </article>
          </div>
        </section>

        <section class="space-y-3">
          <div class="flex flex-wrap items-end justify-between gap-3">
            <div>
              <h2 class="text-sm font-semibold text-highlighted">角色与能力</h2>
              <p class="mt-1 text-xs text-muted">
                贡献者可以独立发布自己的资源；管理员角色受保护，自定义角色可按需组合能力。
              </p>
            </div>
            <UButton
              v-if="draft"
              label="新建自定义角色"
              icon="i-tabler-user-plus"
              color="neutral"
              variant="soft"
              @click="() => { createRoleOpen = true }"
            />
          </div>
          <div class="grid gap-3 lg:grid-cols-2">
            <article
              v-for="role in state.roles"
              :key="role.key"
              class="rounded-xl border border-default bg-default p-4"
            >
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="font-medium text-highlighted">{{ role.displayName }}</p>
                  <p class="mt-0.5 text-xs text-muted">{{ role.key }}</p>
                </div>
                <UBadge
                  :label="role.protected ? '受保护' : role.kind === 'custom' ? '自定义' : '内置'"
                  color="neutral"
                  variant="soft"
                />
              </div>
              <div class="mt-4 grid gap-2 sm:grid-cols-2">
                <UCheckbox
                  v-for="capability in state.capabilities"
                  :key="capability.key"
                  :model-value="role.capabilities.includes(capability.key)"
                  :label="capability.displayName"
                  :disabled="role.protected || !draft || busy"
                  @update:model-value="toggleRoleCapability(role, capability.key)"
                />
              </div>
            </article>
          </div>
        </section>

        <section class="rounded-xl border border-default bg-default p-4 sm:p-5">
          <div class="flex items-start justify-between gap-4">
            <div>
              <h2 class="text-sm font-semibold text-highlighted">注册用户自动成为贡献者</h2>
              <p class="mt-1 text-xs leading-5 text-muted">
                默认开启以保留登录即可贡献的体验。关闭后，新用户需要申请或由管理员直接授权。
              </p>
            </div>
            <USwitch
              :model-value="state.automaticRules[0]?.enabled ?? false"
              :disabled="!draft || busy"
              @update:model-value="toggleAutomatic(Boolean($event))"
            />
          </div>
        </section>

        <section class="space-y-3">
          <div>
            <h2 class="text-sm font-semibold text-highlighted">待处理申请</h2>
            <p class="mt-1 text-xs text-muted">批准后本站授权立即生效，不修改用户中心角色。</p>
          </div>
          <ManageEmpty
            v-if="!state.applications.length"
            icon="i-tabler-user-check"
            text="当前没有待处理申请"
          />
          <div v-else class="divide-y divide-default overflow-hidden rounded-xl border border-default bg-default">
            <article
              v-for="application in state.applications"
              :key="application.id"
              class="grid gap-3 p-4 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center"
            >
              <div class="min-w-0">
                <p class="truncate text-sm font-medium text-highlighted">{{ application.subject }}</p>
                <p class="mt-1 text-xs text-muted">
                  申请 {{ application.role }} · {{ application.reason || "未填写原因" }}
                </p>
              </div>
              <div class="flex gap-2">
                <UButton
                  label="拒绝"
                  color="neutral"
                  variant="outline"
                  :disabled="busy"
                  @click="review(application, 'reject')"
                />
                <UButton label="批准" :disabled="busy" @click="review(application, 'approve')" />
              </div>
            </article>
          </div>
        </section>
      </template>
    </ManageClientBoundary>

    <UModal
      v-model:open="createRoleOpen"
      title="新建自定义角色"
      description="角色能力写入当前策略草稿，发布后才生效。"
    >
      <template #body>
        <div class="space-y-4">
          <UFormField label="角色标识" required>
            <UInput v-model="roleForm.key" placeholder="editor" class="w-full" />
          </UFormField>
          <UFormField label="显示名称" required>
            <UInput v-model="roleForm.displayName" placeholder="编辑" class="w-full" />
          </UFormField>
          <UFormField label="能力">
            <div class="grid gap-2 sm:grid-cols-2">
              <UCheckbox
                v-for="capability in state?.capabilities ?? []"
                :key="capability.key"
                :model-value="roleForm.capabilities.includes(capability.key)"
                :label="capability.displayName"
                @update:model-value="
                  roleForm.capabilities = $event
                    ? [...roleForm.capabilities, capability.key]
                    : roleForm.capabilities.filter((item) => item !== capability.key)
                "
              />
            </div>
          </UFormField>
        </div>
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton
            label="取消"
            color="neutral"
            variant="outline"
            @click="() => { createRoleOpen = false }"
          />
          <UButton
            label="创建角色"
            :disabled="!roleForm.key.trim() || !roleForm.displayName.trim()"
            :loading="busy"
            @click="createRole"
          />
        </div>
      </template>
    </UModal>
  </div>
</template>
