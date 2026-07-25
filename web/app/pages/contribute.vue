<script setup lang="ts">
interface RoleView {
  key: string;
  displayName: string;
  capabilities: string[];
}
interface ApplicationView {
  id: string;
  role: string;
  reason: string;
  state: string;
  createdAt: string;
}

definePageMeta({ middleware: "auth" });
useSeoMeta({ title: "申请贡献资源" });

const { call } = useApi();
const reason = ref("");
const selectedRole = ref("");
const submitting = ref(false);
const message = ref("");
const { data, pending, error, refresh } = await useAsyncData(
  "resource-contributor-application",
  async () => {
    const [roles, mine] = await Promise.all([
      call<{ items: RoleView[] }>("/api/v1/authorization/requestable-roles"),
      call<{ items: ApplicationView[] }>("/api/v1/authorization/applications/mine"),
    ]);
    return { roles: roles.items, applications: mine.items };
  },
  { server: false },
);
watchEffect(() => {
  if (!selectedRole.value && data.value?.roles[0]) {
    selectedRole.value = data.value.roles[0].key;
  }
});
const pendingApplication = computed(() =>
  data.value?.applications.find((item) => item.state === "pending"),
);

async function apply() {
  if (!selectedRole.value || submitting.value || pendingApplication.value) return;
  submitting.value = true;
  message.value = "";
  try {
    await call("/api/v1/authorization/applications", {
      method: "POST",
      body: { role: selectedRole.value, reason: reason.value.trim() },
    });
    message.value = "申请已提交，管理员处理后会在本站生效。";
    reason.value = "";
    await refresh();
  } catch (failure) {
    const apiError = failure as { data?: { message?: string } };
    message.value = apiError.data?.message || "申请提交失败，请稍后重试。";
  } finally {
    submitting.value = false;
  }
}

async function withdraw() {
  if (!pendingApplication.value || submitting.value) return;
  submitting.value = true;
  message.value = "";
  try {
    await call(
      `/api/v1/authorization/applications/${pendingApplication.value.id}/withdraw`,
      { method: "POST" },
    );
    message.value = "申请已撤回。";
    await refresh();
  } catch (failure) {
    const apiError = failure as { data?: { message?: string } };
    message.value = apiError.data?.message || "撤回失败，请稍后重试。";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="mx-auto w-full max-w-3xl px-4 py-10 sm:px-6 lg:py-16">
    <UButton
      to="/"
      label="返回资源站"
      icon="i-tabler-arrow-left"
      color="neutral"
      variant="ghost"
    />
    <div class="mt-6 rounded-2xl border border-default bg-default p-5 sm:p-8">
      <div class="flex items-start gap-4">
        <span class="grid size-11 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary">
          <UIcon name="i-tabler-user-edit" class="size-6" />
        </span>
        <div>
          <h1 class="text-xl font-semibold text-highlighted">申请成为贡献者</h1>
          <p class="mt-2 text-sm leading-6 text-muted">
            普通用户可以申请本站贡献者或其他开放角色。批准后只能使用管理员为该角色配置的能力，
            不会获得其他站点或用户中心的管理员身份。
          </p>
        </div>
      </div>

      <SkeletonList v-if="pending" class="mt-8" :rows="4" />
      <UAlert
        v-else-if="error"
        class="mt-8"
        color="error"
        icon="i-tabler-alert-circle"
        title="申请信息加载失败"
        description="请确认已经登录并稍后重试。"
      />
      <div v-else class="mt-8 space-y-5">
        <UAlert
          v-if="pendingApplication"
          color="warning"
          icon="i-tabler-clock"
          title="已有待处理申请"
          :description="`申请角色：${pendingApplication.role}。管理员处理前无需重复提交。`"
        />
        <UButton
          v-if="pendingApplication"
          label="撤回申请"
          color="neutral"
          variant="outline"
          :loading="submitting"
          @click="withdraw"
        />
        <template v-else-if="data?.roles.length">
          <UFormField label="申请角色" required>
            <USelect
              v-model="selectedRole"
              :items="data.roles.map((role) => ({ label: role.displayName, value: role.key }))"
              value-key="value"
              class="w-full"
            />
          </UFormField>
          <UFormField
            label="申请说明"
            description="说明你希望贡献的资源类型或范围，方便管理员判断。"
          >
            <UTextarea v-model="reason" :rows="5" class="w-full" maxlength="2000" />
          </UFormField>
          <UButton
            label="提交申请"
            icon="i-tabler-send"
            :loading="submitting"
            :disabled="!selectedRole"
            @click="apply"
          />
        </template>
        <UAlert
          v-else
          color="neutral"
          icon="i-tabler-user-off"
          title="当前没有可申请角色"
          description="管理员可以在本站权限设置中开放角色申请来源。"
        />
        <UAlert
          v-if="message"
          color="neutral"
          icon="i-tabler-info-circle"
          :description="message"
        />
      </div>
    </div>
  </div>
</template>

