<script setup lang="ts">
import {
  ManageIconPicker,
  ManageRepeaterRow,
  SkeletonList,
} from "@platform/manage/components";
import {
  platformSettingsSaveMessages,
  usePlatformSettingsProtection,
} from "@platform/manage/settings";
import { createPlatformNotifier } from "@platform/ui/feedback";
import { SiteProfileEditor } from "@yueli/site-profile";
import type {
  SiteProfile,
  SiteProfileReplaceResult,
} from "@yueli/site-profile/types";
import { useActionFeedback, useMinimumLoading } from "@yueli/ui/feedback";
import { SettingsLayout, SettingsSaveDock } from "@yueli/ui/settings/pattern";
import { useVueSettingsWorkflow } from "@yueli/ui/settings/vue";
import type {
  AdminSiteSettingsView,
  HomeSettingsView,
  ResourceView,
  TaxonomyView,
} from "~/types";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "站点设置 · 控制台" });

const { can } = useResourceMe();
const canManageSettings = computed(() => can("resource.site_settings.manage"));
const { call } = useApi();
const toast = createPlatformNotifier(useToast());
const route = useRoute();
const router = useRouter();
const {
  status: saveStatus,
  pending: markSaving,
  success: markSaved,
  reset: resetSave,
} = useActionFeedback();
const saveError = ref("");

const mounted = ref(false);
onMounted(() => {
  mounted.value = true;
});

const sections = [
  {
    key: "home",
    label: "首页设置",
    icon: "i-tabler-home-cog",
    description: "首屏文案、介绍亮点、精选资源和分类入口",
  },
  {
    key: "footer",
    label: "页脚设置",
    icon: "i-tabler-layout-bottombar",
    description: "页脚文案、链接分组、社交入口和备案信息",
  },
  {
    key: "resource",
    label: "资源参数",
    icon: "i-tabler-adjustments-horizontal",
    description: "资源分页、下载开关和大文件提示",
  },
] as const;
const sectionKeys = sections.map((item) => item.key);
const section = ref<"home" | "footer" | "resource">("home");
const activeIconPicker = ref("");
const iconOptions = [
  { label: "资源", value: "i-tabler-package" },
  { label: "文件", value: "i-tabler-file-description" },
  { label: "分类", value: "i-tabler-folder" },
  { label: "标签", value: "i-tabler-tags" },
  { label: "下载", value: "i-tabler-cloud-download" },
  { label: "搜索", value: "i-tabler-search" },
  { label: "链接", value: "i-tabler-link" },
  { label: "外链", value: "i-tabler-arrow-up-right" },
  { label: "星标", value: "i-tabler-star" },
  { label: "闪光", value: "i-tabler-sparkles" },
  { label: "工具", value: "i-tabler-tool" },
  { label: "代码", value: "i-tabler-code" },
  { label: "GitHub", value: "i-tabler-brand-github" },
  { label: "微博", value: "i-tabler-brand-weibo" },
  { label: "微信", value: "i-tabler-brand-wechat" },
];

const homeForm = reactive<HomeSettingsView>({
  heroTitle: "",
  heroSubtitle: "",
  introTitle: "",
  introBody: "",
  introHighlights: [],
  featuredResourceIds: [],
  categorySlugs: [],
  quickLinks: [],
});

const profileForm = reactive<SiteProfile>({
  identity: { name: "", tagline: "", description: "" },
  branding: {},
  announcement: {
    enabled: false,
    text: "",
    tone: "info",
    dismissible: true,
  },
  support: { contacts: [] },
  footer: {
    tagline: "",
    copyright: "",
    linkGroups: [],
    social: [],
    legal: [],
    compliance: { records: [], extraText: "" },
  },
});
const profileEditor = shallowRef<SiteProfileEditor>();
const settingsETag = ref("");
const settingsForm = reactive({
  resource: {
    resourcesPerPage: 0,
    downloadsEnabled: false,
    largeFileThresholdMB: 0,
    largeFileHint: "",
  },
});
const settingsState = useVueSettingsWorkflow({
  snapshot: () => ({
    home: homeForm,
    profile: profileForm,
    resource: settingsForm.resource,
  }),
  restore: (snapshot) => {
    Object.assign(homeForm, snapshot.home);
    applyProfile(snapshot.profile);
    Object.assign(settingsForm.resource, snapshot.resource);
  },
});
usePlatformSettingsProtection(() => settingsState.dirty.value);

const {
  data,
  pending,
  error: loadError,
  refresh,
} = await useAsyncData(
  "resource-settings-editor",
  async () => {
    const [home, settings, resources, categories] = await Promise.all([
      call<{ settings: HomeSettingsView }>("/api/v1/resource/home"),
      call<{ settings: AdminSiteSettingsView }>(
        "/api/v1/admin/resource/settings",
      ),
      call<{ items: ResourceView[] }>("/api/v1/resources/mine", {
        query: { page: 1, size: 100, status: "published" },
      }),
      call<{ items: TaxonomyView[] }>("/api/v1/taxonomies", {
        query: { taxonomy: "category" },
      }),
    ]);
    return {
      home: home.settings,
      settings: settings.settings,
      resources: resources.items,
      categories: categories.items,
    };
  },
  {
    server: false,
    default: () => ({
      home: null as unknown as HomeSettingsView,
      settings: null as unknown as AdminSiteSettingsView,
      resources: [],
      categories: [],
    }),
  },
);

watch(
  () => data.value?.home,
  (settings) => {
    if (!settings) return;
    Object.assign(homeForm, {
      ...settings,
      introHighlights: settings.introHighlights?.length
        ? [...settings.introHighlights]
        : [],
      featuredResourceIds: settings.featuredResourceIds,
      categorySlugs: settings.categorySlugs,
      quickLinks: settings.quickLinks?.length ? [...settings.quickLinks] : [],
    });
  },
  { immediate: true },
);

watch(
  () => data.value?.settings,
  (settings) => {
    if (!settings) return;
    profileEditor.value = new SiteProfileEditor(
      settings.schema,
      settings.snapshot,
    );
    settingsETag.value = settings.etag;
    applyProfile(profileEditor.value.draft);
    Object.assign(settingsForm.resource, settings.resource);
    nextTick(settingsState.capture);
  },
  { immediate: true },
);

watch(
  () => route.query.section,
  (value) => {
    section.value =
      typeof value === "string" && sectionKeys.includes(value as any)
        ? (value as typeof section.value)
        : "home";
  },
  { immediate: true },
);
watch(section, (value) => {
  if (route.query.section === value) return;
  router.replace({ query: { ...route.query, section: value } });
});

const showSkeleton = useMinimumLoading(
  computed(() => !mounted.value || pending.value),
);
const activeSection = computed(
  () => sections.find((item) => item.key === section.value) || sections[0],
);
const resourceItems = computed(() =>
  (data.value?.resources ?? []).map((resource) => ({
    label: resource.title,
    value: resource.id,
  })),
);
const categoryItems = computed(() =>
  (data.value?.categories ?? []).map((category) => ({
    label: category.name,
    value: category.slug,
  })),
);
function isIconPickerOpen(key: string) {
  return activeIconPicker.value === key;
}

function setIconPickerOpen(key: string, open: boolean) {
  activeIconPicker.value = open ? key : "";
}

function chooseIcon(target: { icon?: string }, value: string) {
  target.icon = value;
  activeIconPicker.value = "";
}

function chooseSiteIcon(value: string) {
  profileForm.branding.logo = {
    kind: "icon",
    ref: value,
    alt: profileForm.identity.name,
  };
  activeIconPicker.value = "";
}

function addHighlight() {
  homeForm.introHighlights.push({
    icon: "i-tabler-sparkles",
    title: "",
    text: "",
  });
}

function removeHighlight(index: number) {
  homeForm.introHighlights.splice(index, 1);
}

function addQuickLink() {
  homeForm.quickLinks.push({
    id: newSettingsID("quick"),
    label: "",
    to: "/",
    icon: "i-tabler-link",
  });
}

function removeQuickLink(index: number) {
  homeForm.quickLinks.splice(index, 1);
}

function addFooterGroup() {
  profileForm.footer.linkGroups.push({
    id: newSettingsID("footer-group"),
    title: "链接分组",
    links: [
      {
        id: newSettingsID("footer-link"),
        label: "",
        href: "/",
        icon: "i-tabler-link",
      },
    ],
  });
}

function removeFooterGroup(index: number) {
  profileForm.footer.linkGroups.splice(index, 1);
}

function addFooterLink(group: SiteProfile["footer"]["linkGroups"][number]) {
  group.links.push({
    id: newSettingsID("footer-link"),
    label: "",
    href: "/",
    icon: "i-tabler-link",
  });
}

function removeFooterLink(
  group: SiteProfile["footer"]["linkGroups"][number],
  index: number,
) {
  group.links.splice(index, 1);
}

function addSocialLink() {
  profileForm.footer.social.push({
    id: newSettingsID("social"),
    platform: "social",
    label: "",
    url: "",
    icon: "i-tabler-brand-github",
  });
}

function removeSocialLink(index: number) {
  profileForm.footer.social.splice(index, 1);
}

function linkKey(
  link: {
    id: string;
    label?: string;
    to?: string;
    href?: string;
    url?: string;
  },
  index: number,
) {
  return (
    link.id ||
    `${link.label ?? ""}-${link.to ?? link.href ?? link.url ?? ""}-${index}`
  );
}

function newSettingsID(prefix: string) {
  return `${prefix}-${crypto.randomUUID()}`;
}

async function save() {
  markSaving();
  saveError.value = "";
  try {
    if (!profileEditor.value) throw new Error("站点资料尚未加载");
    profileEditor.value.replaceDraft(toRaw(profileForm));
    const profileRequest = profileEditor.value.request(settingsETag.value);
    const [home, settings] = await Promise.all([
      call<{ settings: HomeSettingsView }>("/api/v1/admin/resource/home", {
        method: "PUT",
        body: homeForm,
      }),
      call<{ settings: AdminSiteSettingsView }>(
        "/api/v1/admin/resource/settings",
        {
          method: profileRequest.method,
          headers: profileRequest.headers,
          body: {
            profile: profileRequest.body.profile,
            resource: settingsForm.resource,
          },
        },
      ),
    ]);
    Object.assign(homeForm, home.settings);
    profileEditor.value.apply({
      snapshot: settings.settings.snapshot,
      changed: true,
    } satisfies SiteProfileReplaceResult);
    settingsETag.value = settings.settings.etag;
    applyProfile(profileEditor.value.draft);
    Object.assign(settingsForm.resource, settings.settings.resource);
    markSaved();
    await refresh();
    settingsState.capture();
  } catch (err) {
    resetSave();
    saveError.value = (err as Error).message;
    toast.add({
      title: "设置保存失败",
      description: saveError.value,
      color: "error",
    });
  }
}

function applyProfile(profile: SiteProfile) {
  Object.assign(profileForm, structuredClone(profile));
}

const logoIcon = computed(() =>
  profileForm.branding.logo?.kind === "icon"
    ? profileForm.branding.logo.ref
    : "",
);

const supportEmail = computed({
  get: () =>
    profileForm.support.contacts.find((contact) => contact.kind === "email")
      ?.value ?? "",
  set: (value: string) => {
    const contact = profileForm.support.contacts.find(
      (item) => item.kind === "email",
    );
    if (contact) {
      contact.value = value;
      return;
    }
    profileForm.support.contacts.push({
      id: newSettingsID("support-email"),
      kind: "email",
      label: "支持邮箱",
      value,
    });
  },
});

function complianceValue(kind: string, key: "number" | "url") {
  return computed({
    get: () =>
      profileForm.footer.compliance.records.find(
        (record) => record.kind === kind,
      )?.[key] ?? "",
    set: (value: string) => {
      let record = profileForm.footer.compliance.records.find(
        (item) => item.kind === kind,
      );
      if (!record) {
        record = {
          id: kind,
          kind,
          label: kind === "icp" ? "ICP备案" : "公安备案",
          number: "",
          url: "",
        };
        profileForm.footer.compliance.records.push(record);
      }
      record[key] = value;
    },
  });
}

const icpRecord = complianceValue("icp", "number");
const icpUrl = complianceValue("icp", "url");
const policeRecord = complianceValue("police", "number");
const policeUrl = complianceValue("police", "url");

function discardChanges() {
  settingsState.discard();
  saveError.value = "";
  resetSave();
}
</script>

<template>
  <SettingsLayout
    v-model:active-section="section"
    :title="activeSection.label"
    :description="activeSection.description"
    :sections="sections"
    :show-section-navigation="false"
    navigation-label="设置分区"
  >
    <SkeletonList v-if="showSkeleton" :rows="6" />

    <UAlert
      v-else-if="loadError"
      color="error"
      icon="i-tabler-alert-circle"
      title="设置加载失败"
      description="站点配置尚未初始化或服务不可用，请先运行开发环境 provision。"
    />

    <UAlert
      v-else-if="!canManageSettings"
      color="warning"
      icon="i-tabler-shield-lock"
      title="需要管理员权限"
      description="站点设置会影响公开页面展示，请使用管理员账户操作。"
    />

    <div v-else class="grid gap-5">
      <section class="min-w-0 space-y-5">
        <template v-if="section === 'home'">
          <div class="rounded-lg border border-default bg-default p-5">
            <h3 class="font-medium text-highlighted">首屏</h3>
            <div class="mt-4 grid gap-4">
              <UFormField label="标题"
                ><UInput v-model="homeForm.heroTitle" class="w-full"
              /></UFormField>
              <UFormField label="副标题"
                ><UTextarea
                  v-model="homeForm.heroSubtitle"
                  :rows="4"
                  class="w-full"
              /></UFormField>
            </div>
          </div>

          <div class="rounded-lg border border-default bg-default p-5">
            <div
              class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
            >
              <h3 class="font-medium text-highlighted">网站介绍</h3>
              <UButton
                size="sm"
                icon="i-tabler-plus"
                color="neutral"
                variant="outline"
                label="添加亮点"
                @click="addHighlight"
              />
            </div>
            <div class="mt-4 space-y-4">
              <UFormField label="介绍标题"
                ><UInput v-model="homeForm.introTitle" class="w-full"
              /></UFormField>
              <UFormField label="介绍正文"
                ><UTextarea
                  v-model="homeForm.introBody"
                  :rows="4"
                  class="w-full"
              /></UFormField>
              <div
                v-for="(item, index) in homeForm.introHighlights"
                :key="index"
                class="grid gap-2 rounded-lg border border-default bg-elevated/30 p-3 sm:grid-cols-[3rem_1fr_auto]"
              >
                <UPopover
                  :open="isIconPickerOpen(`highlight-${index}`)"
                  :content="{ align: 'start', side: 'bottom' }"
                  :ui="{ content: 'w-auto p-3' }"
                  @update:open="setIconPickerOpen(`highlight-${index}`, $event)"
                >
                  <UButton
                    :icon="item.icon || 'i-tabler-sparkles'"
                    color="primary"
                    variant="soft"
                    square
                    class="size-10 justify-center"
                    aria-label="选择亮点图标"
                  />
                  <template #content
                    ><ManageIconPicker
                      :model-value="item.icon"
                      :icon-options="iconOptions"
                      compact
                      @update:model-value="chooseIcon(item, $event)"
                  /></template>
                </UPopover>
                <div class="grid gap-2 sm:grid-cols-2">
                  <UInput
                    v-model="item.title"
                    placeholder="亮点标题"
                    class="w-full"
                  />
                  <UInput
                    v-model="item.text"
                    placeholder="亮点说明"
                    class="w-full"
                  />
                </div>
                <UButton
                  icon="i-tabler-trash"
                  color="error"
                  variant="ghost"
                  aria-label="删除亮点"
                  @click="removeHighlight(index)"
                />
              </div>
            </div>
          </div>

          <div class="rounded-lg border border-default bg-default p-5">
            <h3 class="font-medium text-highlighted">运营位</h3>
            <div class="mt-4 grid gap-4">
              <UFormField label="精选资源">
                <USelectMenu
                  v-model="homeForm.featuredResourceIds"
                  :items="resourceItems"
                  value-key="value"
                  multiple
                  placeholder="选择首页精选资源"
                  :search-input="{ placeholder: '搜索资源标题' }"
                  class="w-full"
                />
              </UFormField>
              <UFormField label="分类入口">
                <USelectMenu
                  v-model="homeForm.categorySlugs"
                  :items="categoryItems"
                  value-key="value"
                  multiple
                  placeholder="选择首页分类入口"
                  :search-input="{ placeholder: '搜索分类' }"
                  class="w-full"
                />
              </UFormField>
            </div>
          </div>

          <div class="rounded-lg border border-default bg-default p-5">
            <div
              class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
            >
              <h3 class="font-medium text-highlighted">快捷链接</h3>
              <UButton
                size="sm"
                icon="i-tabler-plus"
                color="neutral"
                variant="outline"
                label="添加"
                @click="addQuickLink"
              />
            </div>
            <div class="mt-4 space-y-3">
              <ManageRepeaterRow
                v-for="(link, index) in homeForm.quickLinks"
                :key="linkKey(link, index)"
                :label="`快捷链接 ${index + 1}`"
              >
                <div
                  class="grid min-w-0 gap-2 sm:grid-cols-[3rem_1fr_1fr_auto]"
                >
                  <UPopover
                    :open="isIconPickerOpen(`quick-${index}`)"
                    :content="{ align: 'start', side: 'bottom' }"
                    :ui="{ content: 'w-auto p-3' }"
                    @update:open="setIconPickerOpen(`quick-${index}`, $event)"
                  >
                    <UButton
                      :icon="link.icon || 'i-tabler-link'"
                      color="primary"
                      variant="soft"
                      square
                      class="size-10 justify-center"
                      aria-label="选择快捷链接图标"
                    />
                    <template #content
                      ><ManageIconPicker
                        :model-value="link.icon"
                        :icon-options="iconOptions"
                        compact
                        @update:model-value="chooseIcon(link, $event)"
                    /></template>
                  </UPopover>
                  <UInput
                    v-model="link.label"
                    placeholder="按钮名称"
                    class="w-full"
                  />
                  <UInput
                    v-model="link.to"
                    placeholder="/search"
                    class="w-full"
                  />
                  <UButton
                    icon="i-tabler-trash"
                    color="error"
                    variant="ghost"
                    aria-label="删除快捷链接"
                    @click="removeQuickLink(index)"
                  />
                </div>
              </ManageRepeaterRow>
            </div>
          </div>
        </template>

        <template v-else-if="section === 'footer'">
          <div class="rounded-lg border border-default bg-default p-5">
            <h3 class="font-medium text-highlighted">基础文案</h3>
            <div class="mt-4 grid gap-4 sm:grid-cols-2">
              <UFormField label="页脚标语"
                ><UInput v-model="profileForm.footer.tagline" class="w-full"
              /></UFormField>
              <UFormField label="版权信息"
                ><UInput v-model="profileForm.footer.copyright" class="w-full"
              /></UFormField>
            </div>
          </div>

          <div class="rounded-lg border border-default bg-default p-5">
            <h3 class="font-medium text-highlighted">合规信息</h3>
            <div class="mt-4 grid gap-4 sm:grid-cols-2">
              <UFormField label="ICP备案号"
                ><UInput v-model="icpRecord" class="w-full"
              /></UFormField>
              <UFormField label="ICP备案链接"
                ><UInput v-model="icpUrl" class="w-full"
              /></UFormField>
              <UFormField label="公安备案号"
                ><UInput v-model="policeRecord" class="w-full"
              /></UFormField>
              <UFormField label="公安备案链接"
                ><UInput v-model="policeUrl" class="w-full"
              /></UFormField>
            </div>
            <UFormField class="mt-4" label="其他页脚信息"
              ><UTextarea
                v-model="profileForm.footer.compliance.extraText"
                :rows="3"
                class="w-full"
            /></UFormField>
          </div>

          <div class="rounded-lg border border-default bg-default p-5">
            <div
              class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
            >
              <h3 class="font-medium text-highlighted">链接分组</h3>
              <UButton
                size="sm"
                icon="i-tabler-plus"
                color="neutral"
                variant="outline"
                label="添加分组"
                @click="addFooterGroup"
              />
            </div>
            <div class="mt-4 space-y-4">
              <div
                v-for="(group, groupIndex) in profileForm.footer.linkGroups"
                :key="groupIndex"
                class="rounded-lg border border-default bg-elevated/30 p-4"
              >
                <div class="grid gap-2 sm:grid-cols-[1fr_auto_auto]">
                  <UInput
                    v-model="group.title"
                    placeholder="分组标题"
                    class="w-full"
                  />
                  <UButton
                    size="sm"
                    icon="i-tabler-plus"
                    color="neutral"
                    variant="outline"
                    label="链接"
                    @click="addFooterLink(group)"
                  />
                  <UButton
                    icon="i-tabler-trash"
                    color="error"
                    variant="ghost"
                    aria-label="删除分组"
                    @click="removeFooterGroup(groupIndex)"
                  />
                </div>
                <div class="mt-3 space-y-2">
                  <ManageRepeaterRow
                    v-for="(link, linkIndex) in group.links"
                    :key="linkKey(link, linkIndex)"
                    :label="`页脚链接 ${linkIndex + 1}`"
                  >
                    <div class="grid min-w-0 gap-2 sm:grid-cols-[1fr_1fr_auto]">
                      <UInput
                        v-model="link.label"
                        placeholder="链接名称"
                        class="w-full"
                      />
                      <UInput
                        v-model="link.href"
                        placeholder="/search"
                        class="w-full"
                      />
                      <UButton
                        icon="i-tabler-trash"
                        color="error"
                        variant="ghost"
                        aria-label="删除链接"
                        @click="removeFooterLink(group, linkIndex)"
                      />
                    </div>
                  </ManageRepeaterRow>
                </div>
              </div>
            </div>
          </div>

          <div class="rounded-lg border border-default bg-default p-5">
            <div
              class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
            >
              <h3 class="font-medium text-highlighted">社交入口</h3>
              <UButton
                size="sm"
                icon="i-tabler-plus"
                color="neutral"
                variant="outline"
                label="添加"
                @click="addSocialLink"
              />
            </div>
            <div class="mt-4 space-y-3">
              <ManageRepeaterRow
                v-for="(link, index) in profileForm.footer.social"
                :key="linkKey(link, index)"
                :label="`社交入口 ${index + 1}`"
              >
                <div
                  class="grid min-w-0 gap-2 sm:grid-cols-[3rem_1fr_1fr_auto]"
                >
                  <UPopover
                    :open="isIconPickerOpen(`social-${index}`)"
                    :content="{ align: 'start', side: 'bottom' }"
                    :ui="{ content: 'w-auto p-3' }"
                    @update:open="setIconPickerOpen(`social-${index}`, $event)"
                  >
                    <UButton
                      :icon="link.icon || 'i-tabler-brand-github'"
                      color="primary"
                      variant="soft"
                      square
                      class="size-10 justify-center"
                      aria-label="选择社交入口图标"
                    />
                    <template #content
                      ><ManageIconPicker
                        :model-value="link.icon"
                        :icon-options="iconOptions"
                        compact
                        @update:model-value="chooseIcon(link, $event)"
                    /></template>
                  </UPopover>
                  <UInput
                    v-model="link.label"
                    placeholder="GitHub"
                    class="w-full"
                  />
                  <UInput
                    v-model="link.url"
                    placeholder="https://..."
                    class="w-full"
                  />
                  <UButton
                    icon="i-tabler-trash"
                    color="error"
                    variant="ghost"
                    aria-label="删除社交入口"
                    @click="removeSocialLink(index)"
                  />
                </div>
              </ManageRepeaterRow>
            </div>
          </div>
        </template>

        <template v-else>
          <div class="rounded-lg border border-default bg-default p-5">
            <h3 class="font-medium text-highlighted">站点基础</h3>
            <div class="mt-4 grid gap-4 sm:grid-cols-2">
              <UFormField label="站点名称"
                ><UInput v-model="profileForm.identity.name" class="w-full"
              /></UFormField>
              <UFormField label="站点图标">
                <UPopover
                  :open="isIconPickerOpen('site-logo')"
                  :content="{ align: 'start', side: 'bottom' }"
                  :ui="{ content: 'w-auto p-3' }"
                  @update:open="setIconPickerOpen('site-logo', $event)"
                >
                  <UButton
                    :icon="logoIcon || 'i-tabler-package'"
                    :label="logoIcon || '选择图标'"
                    color="neutral"
                    variant="outline"
                    class="w-full justify-start font-mono"
                  />
                  <template #content
                    ><ManageIconPicker
                      :model-value="logoIcon"
                      :icon-options="iconOptions"
                      compact
                      @update:model-value="chooseSiteIcon"
                  /></template>
                </UPopover>
              </UFormField>
              <UFormField label="一句话描述"
                ><UInput v-model="profileForm.identity.tagline" class="w-full"
              /></UFormField>
              <UFormField label="支持邮箱"
                ><UInput v-model="supportEmail" type="email" class="w-full"
              /></UFormField>
            </div>
            <div
              class="mt-4 grid gap-3 sm:grid-cols-[10rem_minmax(0,1fr)] sm:items-start"
            >
              <UFormField label="公告启用"
                ><USwitch v-model="profileForm.announcement.enabled"
              /></UFormField>
              <UFormField label="公告内容"
                ><UTextarea
                  v-model="profileForm.announcement.text"
                  :rows="3"
                  class="w-full"
              /></UFormField>
            </div>
          </div>

          <div class="rounded-lg border border-default bg-default p-5">
            <h3 class="font-medium text-highlighted">资源运行参数</h3>
            <div class="mt-4 grid gap-4 sm:grid-cols-2">
              <UFormField label="资源每页数量"
                ><UInput
                  v-model.number="settingsForm.resource.resourcesPerPage"
                  type="number"
                  min="1"
                  max="96"
                  class="w-full"
              /></UFormField>
              <UFormField label="大文件提示阈值（MB）"
                ><UInput
                  v-model.number="settingsForm.resource.largeFileThresholdMB"
                  type="number"
                  min="1"
                  max="5120"
                  class="w-full"
              /></UFormField>
            </div>
            <UFormField class="mt-4" label="下载开关">
              <div class="flex min-h-8 items-center gap-3">
                <USwitch v-model="settingsForm.resource.downloadsEnabled" />
                <span class="text-sm text-muted">{{
                  settingsForm.resource.downloadsEnabled
                    ? "允许前台下载"
                    : "隐藏前台下载按钮"
                }}</span>
              </div>
            </UFormField>
            <UFormField class="mt-4" label="大文件提示文案"
              ><UTextarea
                v-model="settingsForm.resource.largeFileHint"
                :rows="3"
                class="w-full"
            /></UFormField>
          </div>
        </template>
      </section>
    </div>
    <SettingsSaveDock
      :dirty="settingsState.dirty.value"
      :status="saveStatus"
      :error="saveError"
      :disabled="!canManageSettings"
      :messages="platformSettingsSaveMessages"
      dock-class="lg:left-60"
      @discard="discardChanges"
      @save="save"
    />
  </SettingsLayout>
</template>
