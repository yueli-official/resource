import { requiredEnv } from "./runtime";

export interface BrowserContract {
  version: 1;
  public: BrowserRoute;
  manage: BrowserManageRoute;
  empty: BrowserEmptyState;
  error: BrowserErrorState;
  visual?: {
    managePath?: string;
    scenarios?: Array<{
      key: string;
      path: string;
      readySelector: string;
      authenticated?: boolean;
      status?: number;
      viewports?: Array<"mobile" | "tablet" | "medium" | "desktop">;
      actions?: Array<{
        role: "button" | "link" | "checkbox";
        name: string;
      }>;
    }>;
  };
  features?: { assetSettings?: boolean };
}

export interface BrowserRoute {
  path: string;
  readySelector: string;
}

export interface BrowserManageRoute extends BrowserRoute {
  heading: string;
  settings?: { path: string; fieldLabel: string };
  views?: Array<{ key: string; label: string }>;
}

export interface BrowserEmptyState {
  path: string;
  text: string;
  inputPlaceholder?: string;
  authenticated?: boolean;
}

export interface BrowserErrorState {
  path: string;
  status: number;
  text: string;
}

export interface SiteContract {
  slug: string;
  product: string;
  url: string;
  contract: BrowserContract;
}

let parsedSites: SiteContract[] | undefined;

export function sites(): SiteContract[] {
  if (parsedSites) return parsedSites;
  const value = JSON.parse(
    requiredEnv("RESOURCE_E2E_SITES"),
  ) as SiteContract[];
  if (!Array.isArray(value) || value.length === 0)
    throw new Error("浏览器验收矩阵至少需要一个站点实例");
  for (const site of value) validateSite(site);
  parsedSites = value;
  return value;
}

export function productSites(product: string): SiteContract[] {
  return sites().filter((site) => site.product === product);
}

function validateSite(site: SiteContract) {
  const contract = site.contract;
  if (!site.slug || !site.product || !site.url || contract?.version !== 1) {
    throw new Error(`站点 ${site.slug || "unknown"} 的浏览器合同不完整`);
  }
  for (const route of [contract.public, contract.manage]) {
    if (!route?.path || !route.readySelector)
      throw new Error(`站点 ${site.slug} 缺少路由就绪事实`);
  }
  if (
    !contract.manage.heading ||
    !contract.empty?.path ||
    !contract.empty.text ||
    !contract.error?.path ||
    !contract.error.text
  ) {
    throw new Error(`站点 ${site.slug} 缺少管理、空状态或错误状态事实`);
  }
}
