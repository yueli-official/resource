import { expect, test } from "@playwright/test";
import { productSites } from "./contracts";
import {
  capturePageFailures,
  loginE2E,
  requiredEnv,
  settleNuxt,
} from "./runtime";

const accountURL = requiredEnv("RESOURCE_E2E_ACCOUNT_URL");
const missing = "__resource_e2e_missing__";

export function registerJourneySuite(product: string) {
  for (const site of productSites(product)) {
    const contract = site.contract;
    test.describe(`${site.slug} (${site.product})`, () => {
      test("公开入口完成渲染且没有浏览器错误", async ({ page }) => {
        const errors = capturePageFailures(page);
        const response = await page.goto(
          new URL(contract.public.path, site.url).toString(),
          { waitUntil: "domcontentloaded" },
        );
        expect(response?.ok()).toBeTruthy();
        await expect(
          page.locator(contract.public.readySelector).first(),
        ).toBeVisible();
        await settleNuxt(page);
        expect(errors).toEqual([]);
      });

      test("匿名访问管理入口进入账户登录流程", async ({ page }) => {
        const errors = capturePageFailures(page);
        await page.goto(new URL(contract.manage.path, site.url).toString(), {
          waitUntil: "domcontentloaded",
        });
        await expect(page).toHaveURL(
          new RegExp(
            `^${accountURL.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}/login(?:\\?|$)`,
          ),
        );
        await expect(
          page.getByRole("heading", { name: "欢迎回来" }),
        ).toBeVisible();
        await settleNuxt(page);
        expect(errors).toEqual([]);
      });

      test("管理壳刷新时直接由服务端渲染", async ({ browser }) => {
        const context = await loginE2E(browser);
        const sessionResponse = await context.request.get(
          new URL("/auth/login?return_to=/manage", site.url).toString(),
        );
        const sessionBody = await sessionResponse.text();
        const sessionError = sessionBody
          .replace(/<script[\s\S]*?<\/script>/giu, " ")
          .replace(/<style[\s\S]*?<\/style>/giu, " ")
          .replace(/<[^>]+>/gu, " ")
          .replace(/\s+/gu, " ")
          .trim();
        expect(
          sessionResponse.ok(),
          `site session bootstrap failed with HTTP ${sessionResponse.status()} at ${sessionResponse.url()}: ${sessionError.slice(0, 1_000)}`,
        ).toBeTruthy();
        await context.addInitScript(() => {
          const state = { samples: [] as Array<{ text: string; at: number }> };
          Object.defineProperty(globalThis, "__adminOpeningProbe", {
            value: state,
            configurable: true,
          });
          const startedAt = performance.now();
          const sample = () => {
            const text = document.body?.innerText || "";
            const match = text.match(/正在打开[^\n]{0,16}控制台/u);
            if (!match || state.samples.some((entry) => entry.text === match[0]))
              return;
            state.samples.push({ text: match[0], at: performance.now() - startedAt });
          };
          new MutationObserver(sample).observe(document, {
            childList: true,
            subtree: true,
            characterData: true,
          });
          document.addEventListener("DOMContentLoaded", sample, { once: true });
        });
        const page = await context.newPage();
        try {
          const manageURL = new URL(contract.manage.path, site.url).toString();
          for (let index = 0; index < 3; index += 1) {
            const manageResponse = index === 0
              ? await page.goto(manageURL, { waitUntil: "domcontentloaded" })
              : await page.reload({ waitUntil: "domcontentloaded" });
            const manageHTML = await manageResponse?.text();
            expect(manageHTML).not.toMatch(/正在打开[^\n]{0,16}控制台/u);
            expect(manageHTML).toContain("data-admin-shell");
            await expect(page.locator("[data-admin-shell]")).toBeVisible();
            await page.waitForTimeout(100);
            const openingSamples = await page.evaluate(() =>
              (
                globalThis as typeof globalThis & {
                  __adminOpeningProbe?: {
                    samples: Array<{ text: string; at: number }>;
                  };
                }
              ).__adminOpeningProbe?.samples ?? []
            );
            expect(openingSamples).toEqual([]);
            await expect(
              page.getByText(/正在打开[^\n]{0,16}控制台/u),
            ).toHaveCount(0);
          }
        } finally {
          await context.close();
        }
      });

      test("已登录运营者可以进入管理界面", async ({ browser }) => {
        const context = await loginE2E(browser);
        const page = await context.newPage();
        const errors = capturePageFailures(page);
        try {
          const manageURL = new URL(contract.manage.path, site.url).toString();
          await page.goto(manageURL, { waitUntil: "domcontentloaded" });
          await expect(page).toHaveURL(manageURL);
          await expect(
            page.locator(contract.manage.readySelector).first(),
          ).toBeVisible();
          await expect(
            page
              .getByRole("heading", {
                name: new RegExp(contract.manage.heading),
              })
              .first(),
          ).toBeVisible();
          await expect(
            page.locator('[data-admin-sidebar-appearance="commercial"]'),
          ).toBeVisible();
          await expect(
            page.locator('[data-admin-sidebar-brand] a[href="/"]'),
          ).toBeVisible();
          await expect(
            page.getByRole("button", { name: /打开.+站点菜单/ }),
          ).toHaveCount(0);
          await settleNuxt(page);
          expect(errors).toEqual([]);
        } finally {
          await context.close();
        }
      });

      test("设置页脏数据保护阻止意外离开", async ({ browser }) => {
        const settings = contract.manage.settings;
        test.skip(!settings, "该产品没有可编辑的通用设置页");
        if (!settings) return;
        const context = await loginE2E(browser);
        const page = await context.newPage();
        const errors = capturePageFailures(page);
        try {
          await page.goto(new URL(settings.path, site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          const field = page
            .getByRole("textbox", { name: settings.fieldLabel, exact: false })
            .first();
          await expect(field).toBeVisible();
          await expect(page.locator('[data-manage-dock="save"]')).toHaveCount(
            0,
          );
          await field.fill(`${await field.inputValue()} · 未保存`);
          await expect(page.locator('[data-manage-dock="save"]')).toBeVisible();

          const dialogHandled = new Promise<void>((resolve) => {
            page.once("dialog", async (dialog) => {
              expect(dialog.type()).toBe("confirm");
              expect(dialog.message()).toContain("未保存");
              await dialog.dismiss();
              resolve();
            });
          });
          await page
            .locator(`a[href="${contract.manage.path}"]`)
            .first()
            .click();
          await dialogHandled;
          await expect(page).toHaveURL(
            new RegExp(
              `${settings.path.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}(?:\\?|$)`,
            ),
          );
          expect(errors).toEqual([]);
        } finally {
          await context.close();
        }
      });

      test("空结果状态有明确反馈", async ({ browser, page }) => {
        const context = contract.empty.authenticated
          ? await loginE2E(browser)
          : undefined;
        const targetPage = context ? await context.newPage() : page;
        const errors = capturePageFailures(targetPage);
        try {
          await targetPage.goto(
            new URL(contract.empty.path, site.url).toString(),
            { waitUntil: "domcontentloaded" },
          );
          const emptyState = targetPage
            .getByText(contract.empty.text, { exact: true })
            .first();
          if (contract.empty.inputPlaceholder) {
            const input = targetPage.getByPlaceholder(
              contract.empty.inputPlaceholder,
            );
            await expect(async () => {
              await input.fill(missing);
              await expect(input).toHaveValue(missing);
              await expect(emptyState).toBeVisible({ timeout: 2_000 });
            }).toPass({ timeout: 15_000, intervals: [250, 500, 1_000] });
          }
          await expect(emptyState).toBeVisible();
          await settleNuxt(targetPage);
          expect(errors).toEqual([]);
        } finally {
          await context?.close();
        }
      });

      test("缺失实体返回产品错误状态", async ({ page }) => {
        const errors = capturePageFailures(page);
        const errorURL = new URL(contract.error.path, site.url).toString();
        const response = await page.goto(errorURL, {
          waitUntil: "domcontentloaded",
        });
        expect(response?.status()).toBe(contract.error.status);
        await expect(
          page.getByText(contract.error.text, { exact: false }).first(),
        ).toBeVisible();
        await settleNuxt(page);
        expect(
          errors.filter(
            (error) =>
              error !== `http ${contract.error.status}: ${errorURL}` &&
              !error.includes(`status of ${contract.error.status}`),
          ),
        ).toEqual([]);
      });
    });
  }
}
