import { expect, test, type Page } from "@playwright/test";
import { productSites, type BrowserContract } from "./contracts";
import {
  expectNoHorizontalOverflow,
  loginE2E,
  settleNuxt,
} from "./runtime";

function scenario(contract: BrowserContract, key: string) {
  const value = contract.visual?.scenarios?.find((item) => item.key === key);
  if (!value) throw new Error(`缺少响应式验收场景：${key}`);
  return value;
}

async function openReady(
  page: Page,
  siteURL: string,
  route: { path: string; readySelector: string; status?: number },
) {
  const response = await page.goto(new URL(route.path, siteURL).toString(), {
    waitUntil: "domcontentloaded",
  });
  expect(response?.status()).toBe(route.status || 200);
  await expect(page.locator(route.readySelector).first()).toBeVisible();
  await settleNuxt(page);
  await expectNoHorizontalOverflow(page, `${route.path} `);
  await expect(page.locator("main").first()).toBeVisible();
}

export function registerResponsiveSuite(product: string) {
  for (const site of productSites(product)) {
    const contract = site.contract;
    test.describe(`${site.slug} (${site.product}) 极端响应式合同`, () => {
      test("320px 窄屏关键消费路径完整且无横向溢出", async ({ page }) => {
        await page.setViewportSize({ width: 320, height: 720 });
        for (const key of ["catalog", "collections", "viewer"]) {
          await openReady(page, site.url, scenario(contract, key));
        }
        await expect(
          page.getByRole("button", { name: "查看下一张图片" }),
        ).toBeVisible();
      });

      test("1920px 宽屏内容有上限并保持居中", async ({ browser }) => {
        const viewport = { width: 1920, height: 1080 };
        const publicContext = await browser.newContext({ viewport });
        try {
          const page = await publicContext.newPage();
          await openReady(page, site.url, {
            path: contract.public.path,
            readySelector: contract.public.readySelector,
          });
          const publicBounds = await page
            .locator(".resource-page")
            .first()
            .boundingBox();
          expect(publicBounds).not.toBeNull();
          expect(publicBounds!.width).toBeLessThanOrEqual(1601);
          expect(
            Math.abs(
              publicBounds!.x - (viewport.width - publicBounds!.width) / 2,
            ),
          ).toBeLessThanOrEqual(2);
        } finally {
          await publicContext.close();
        }

        const context = await loginE2E(browser, { viewport });
        try {
          const managePage = await context.newPage();
          await openReady(
            managePage,
            site.url,
            scenario(contract, "manage-authorization"),
          );
          const manageBounds = await managePage
            .locator("#authorization")
            .boundingBox();
          expect(manageBounds).not.toBeNull();
          expect(manageBounds!.width).toBeLessThanOrEqual(1537);
        } finally {
          await context.close();
        }
      });

      test("200% 有效缩放下目录可重排且关键控件可操作", async ({
        browser,
      }) => {
        const context = await browser.newContext({
          viewport: { width: 640, height: 450 },
          screen: { width: 1280, height: 900 },
          deviceScaleFactor: 2,
        });
        try {
          const page = await context.newPage();
          await openReady(page, site.url, scenario(contract, "catalog"));
          expect(
            await page.evaluate(() => ({
              width: window.innerWidth,
              pixelRatio: window.devicePixelRatio,
            })),
          ).toEqual({ width: 640, pixelRatio: 2 });
          await expect(
            page.getByRole("button", { name: "筛选", exact: true }),
          ).toBeVisible();
          await page.getByRole("button", { name: "筛选", exact: true }).click();
          await expect(page.getByRole("dialog")).toBeVisible();
        } finally {
          await context.close();
        }
      });
    });
  }
}
