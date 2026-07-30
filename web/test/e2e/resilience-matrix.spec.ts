import { expect, test } from "@playwright/test";
import { productSites } from "./contracts";
import { settleNuxt } from "./runtime";

export function registerResilienceSuite(product: string) {
  for (const site of productSites(product)) {
    test.describe(`${site.slug} (${site.product}) resilience contract`, () => {
      test("投稿列表依赖失败可见且重试后恢复服务器真值", async ({
        browser,
      }) => {
        const context = await browser.newContext();
        const page = await context.newPage();
        let failRequest = true;
        await page.route("**/api/resource/me/submissions**", async (route) => {
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
          await page.goto(new URL("/submissions", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          await expect(
            page.getByText("投稿记录加载失败", { exact: true }),
          ).toBeVisible();
          const retry = page.getByRole("button", {
            name: "重试",
            exact: true,
          });
          await expect(retry).toBeVisible();

          failRequest = false;
          const recoveredResponse = page.waitForResponse(
            (response) =>
              new URL(response.url()).pathname ===
                "/api/resource/me/submissions" && response.status() === 200,
          );
          await retry.click();
          await recoveredResponse;
          await expect(
            page.getByText("投稿记录加载失败", { exact: true }),
          ).toHaveCount(0);
          await expect(
            page.getByText("还没有投稿记录", { exact: true }),
          ).toBeVisible();
        } finally {
          await context.close();
        }
      });
    });
  }
}
