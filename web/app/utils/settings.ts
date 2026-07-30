import { onMounted, onScopeDispose } from "vue";
import { bindSettingsBeforeUnload } from "@yueli/ui/settings/browser";
import type { SettingsSaveDockMessages } from "@yueli/ui/settings/pattern";
import { useSettingsLeaveGuard } from "@yueli/ui/settings/vue-router";

export const resourceSettingsSaveMessages: SettingsSaveDockMessages = {
  region: "设置保存操作",
  unsaved: "有未保存的更改",
  saving: "正在保存更改",
  saved: "更改已保存",
  failed: "保存失败",
  discard: "放弃",
  save: "保存",
  savePending: "保存中",
  saveSuccess: "已保存",
};

export function useResourceSettingsProtection(isDirty: () => boolean) {
  let unbindBeforeUnload: (() => void) | undefined;
  onMounted(() => {
    unbindBeforeUnload = bindSettingsBeforeUnload({ isDirty });
  });
  onScopeDispose(() => unbindBeforeUnload?.());
  useSettingsLeaveGuard({
    isDirty,
    confirm: () => window.confirm("有未保存的更改，确定离开当前页面吗？"),
  });
}
