import { expect, test } from "@playwright/test";
import { productSites } from "./contracts";
import { loginE2E, settleNuxt } from "./runtime";

export function registerResilienceSuite(product: string) {
  for (const site of productSites(product)) {
    test.describe(`${site.slug} (${site.product}) resilience contract`, () => {
      test("资源列表依赖失败可见且重试后恢复服务器真值", async ({
        browser,
      }) => {
        const context = await loginE2E(browser);
        const page = await context.newPage();
        let failRequest = true;
        await page.route("**/api/v1/resources/mine**", async (route) => {
          if (!failRequest) return route.continue();
          await route.fulfill({
            status: 503,
            contentType: "application/problem+json",
            body: JSON.stringify({
              type: "about:blank",
              title: "Service unavailable",
              status: 503,
            }),
          });
        });
        try {
          await page.goto(new URL("/manage", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          await expect(
            page.getByText("资源加载失败", { exact: true }),
          ).toBeVisible();
          const retry = page.getByRole("button", {
            name: "重新加载",
            exact: true,
          });
          await expect(retry).toBeVisible();

          failRequest = false;
          const recoveredResponse = page.waitForResponse(
            (response) =>
              new URL(response.url()).pathname ===
                "/api/v1/resources/mine" && response.status() === 200,
          );
          await retry.click();
          await recoveredResponse;
          await expect(
            page.getByText("资源加载失败", { exact: true }),
          ).toHaveCount(0);
          await expect(page.locator('[aria-label="资源列表"]')).toBeVisible();
        } finally {
          await context.close();
        }
      });

      test("管理写入 Problem 使用产品文案原位反馈", async ({ browser }) => {
        const context = await loginE2E(browser);
        const page = await context.newPage();
        await page.route("**/api/v1/resources", async (route) => {
          if (route.request().method() !== "POST") return route.continue();
          await route.fulfill({
            status: 400,
            contentType: "application/problem+json",
            body: JSON.stringify({
              type: "https://errors.yueli.dev/problems/resource.invalid_input",
              status: 400,
              code: "resource.invalid_input",
              violations: [],
              traceId: "resource-e2e-invalid-input",
            }),
          });
        });
        try {
          await page.goto(new URL("/manage", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          await page.getByRole("button", { name: "新建资源" }).click();
          await page.getByRole("textbox", { name: "标题" }).fill("HTTP Result E2E");
          await page.getByRole("button", { name: "创建并编辑" }).click();
          const dialog = page.getByRole("dialog", { name: "新建资源（草稿）" });
          await expect(dialog.getByText("暂时无法创建", { exact: true })).toBeVisible();
          await expect(dialog.getByText("提交内容不符合要求。", { exact: true })).toBeVisible();
          await expect(page).toHaveURL(new URL("/manage", site.url).toString());
        } finally {
          await context.close();
        }
      });
    });
  }
}
