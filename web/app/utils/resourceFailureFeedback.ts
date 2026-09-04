import { resolveFailureFeedback, type ProblemParams } from "@yueli/http-runtime";
import { resourceFailurePresentation, type ResourceFailureCode } from "../generated/resourceFailure";

const messages: Record<ResourceFailureCode, string> = {
  "resource.asset_not_found": "关联的文件不存在或已被删除。",
  "resource.authorization_unavailable": "权限服务暂时不可用。",
  "resource.forbidden": "你没有执行此操作的权限。",
  "resource.invalid_input": "提交内容不符合要求。",
  "resource.invalid_state": "当前状态不允许执行此操作。",
  "resource.invalid_type": "该文件或资源类型不受支持。",
  "resource.not_found": "请求的资源不存在。",
  "resource.site_profile_precondition_required": "请刷新设置后再提交。",
  "resource.site_profile_revision_conflict": "设置已被其他操作更新，请刷新后重试。",
  "resource.slug_taken": "这个链接地址已被使用。",
  "resource.upstream_failed": "依赖服务暂时不可用。",
};

function resolveText(code: string, _params: ProblemParams) {
  if (!(code in resourceFailurePresentation)) return undefined;
  const typedCode = code as ResourceFailureCode;
  const presentation = resourceFailurePresentation[typedCode];
  const recoveryKey = "recoveryKey" in presentation ? presentation.recoveryKey : undefined;
  return {
    message: messages[typedCode],
    ...(recoveryKey === "recovery.retry_later" ? { recovery: "请稍后重试。" } : {}),
  };
}

export function resourceFailureFeedback(error: unknown, fallback: string) {
  return resolveFailureFeedback(error, { fallback, resolveText });
}

export function resourceFailureMessage(error: unknown, fallback: string) {
  const feedback = resourceFailureFeedback(error, fallback);
  return feedback.recovery ? `${feedback.message}${feedback.recovery}` : feedback.message;
}
