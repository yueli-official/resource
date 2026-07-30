import { expect, test, type BrowserContext, type Page } from "@playwright/test";
import { productSites } from "./contracts";
import {
  capturePageFailures,
  ensureRegisteredE2EIdentity,
  loginE2E,
  settleNuxt,
} from "./runtime";

const managementRoutes = [
  "/manage",
  "/manage/images",
  "/manage/submissions",
  "/manage/collections",
  "/manage/classification",
  "/manage/cases",
  "/manage/discovery",
  "/manage/assets",
  "/manage/authorization",
] as const;

async function saveDiscoverySettings(page: Page): Promise<void> {
  const response = page.waitForResponse(
    (candidate) =>
      candidate.request().method() === "PATCH" &&
      candidate.url().includes("/api/resource/admin/site-settings"),
  );
  await page
    .locator('[data-manage-dock="save"]')
    .getByRole("button", { name: "保存设置", exact: true })
    .click();
  expect((await response).ok()).toBeTruthy();
  await expect(
    page.getByText("站点与首页设置已保存", { exact: true }),
  ).toBeVisible();
}

async function restoreDiscoveryName(
  page: Page,
  siteURL: string,
  originalName: string,
): Promise<void> {
  await page.goto(new URL("/manage/discovery", siteURL).toString(), {
    waitUntil: "domcontentloaded",
  });
  await settleNuxt(page);
  const field = page
    .getByRole("textbox", { name: "站点名称", exact: false })
    .first();
  if ((await field.inputValue()) === originalName) return;
  await field.fill(originalName);
  await saveDiscoverySettings(page);
}

async function openFirstImageEditor(page: Page): Promise<void> {
  await page
    .getByRole("button", { name: "编辑图片", exact: true })
    .first()
    .click();
  await expect(page.getByRole("dialog", { name: "编辑图片" })).toBeVisible();
  await expect(
    page
      .getByRole("dialog", { name: "编辑图片" })
      .getByRole("textbox", { name: "标题", exact: false }),
  ).toBeVisible();
}

async function saveImageEditor(page: Page): Promise<void> {
  const response = page.waitForResponse(
    (candidate) =>
      candidate.request().method() === "PATCH" &&
      /\/api\/resource\/admin\/images\/[^/?]+$/.test(candidate.url()),
  );
  await page
    .getByRole("dialog", { name: "编辑图片" })
    .getByRole("button", { name: "保存更改", exact: true })
    .click();
  expect((await response).ok()).toBeTruthy();
  await expect(page.getByRole("dialog", { name: "编辑图片" })).toHaveCount(0);
}

async function searchManagedImages(page: Page, query: string): Promise<void> {
  const search = page.getByPlaceholder("搜索标题、说明或替代文本…");
  await search.fill(query);
  await search.press("Enter");
  await expect(page.getByText(query, { exact: true }).first()).toBeVisible();
}

async function restoreImageTitle(
  page: Page,
  siteURL: string,
  temporaryTitle: string,
  originalTitle: string,
): Promise<void> {
  await page.goto(new URL("/manage/images", siteURL).toString(), {
    waitUntil: "domcontentloaded",
  });
  await settleNuxt(page);
  await searchManagedImages(page, temporaryTitle);
  await openFirstImageEditor(page);
  const title = page
    .getByRole("dialog", { name: "编辑图片" })
    .getByRole("textbox", { name: "标题", exact: false });
  if ((await title.inputValue()) === originalTitle) return;
  await title.fill(originalTitle);
  await saveImageEditor(page);
}

async function saveCollectionSettings(page: Page): Promise<void> {
  const response = page.waitForResponse(
    (candidate) =>
      candidate.request().method() === "PATCH" &&
      /\/api\/resource\/admin\/collections\/[^/?]+$/.test(candidate.url()),
  );
  await page
    .getByRole("button", { name: "保存专题设置", exact: true })
    .click();
  expect((await response).ok()).toBeTruthy();
  await expect(
    page.getByText("专题设置已保存", { exact: true }),
  ).toBeVisible();
}

async function restoreCollectionName(
  page: Page,
  collectionURL: string,
  originalName: string,
): Promise<void> {
  await page.goto(collectionURL, { waitUntil: "domcontentloaded" });
  await settleNuxt(page);
  const name = page
    .getByRole("textbox", { name: "名称", exact: false })
    .first();
  if ((await name.inputValue()) === originalName) return;
  await name.fill(originalName);
  await saveCollectionSettings(page);
}

async function saveAssetSettings(page: Page): Promise<void> {
  const response = page.waitForResponse(
    (candidate) =>
      candidate.request().method() === "POST" &&
      /\/api\/v1\/admin\/assets-proxy\/sites(?:\?|$)/.test(candidate.url()),
  );
  await page.getByRole("button", { name: "保存", exact: true }).click();
  expect((await response).ok()).toBeTruthy();
  await expect(page.getByText("资源设置已保存", { exact: true })).toBeVisible();
}

async function restoreAssetSiteName(
  page: Page,
  siteURL: string,
  originalName: string,
): Promise<void> {
  await page.goto(new URL("/manage/assets", siteURL).toString(), {
    waitUntil: "domcontentloaded",
  });
  await settleNuxt(page);
  const name = page
    .getByRole("textbox", { name: "站点名称", exact: false })
    .first();
  if ((await name.inputValue()) === originalName) return;
  await name.fill(originalName);
  await saveAssetSettings(page);
}

async function authenticatedPage(
  context: BrowserContext,
  siteURL: string,
): Promise<Page> {
  const page = await context.newPage();
  await page.goto(new URL("/manage", siteURL).toString(), {
    waitUntil: "domcontentloaded",
  });
  await expect(page).toHaveURL(new URL("/manage", siteURL).toString());
  await settleNuxt(page);
  return page;
}

export function registerManagementSuite(product: string) {
  for (const site of productSites(product)) {
    test.describe(`${site.slug} (${site.product}) management contract`, () => {
      test("匿名调用 Resource 管理 API 被拒绝", async ({ request }) => {
        for (const path of [
          "/api/resource/admin/overview",
          "/api/resource/admin/site-settings",
          "/api/resource/admin/images?size=1",
          "/api/resource/admin/collections",
        ]) {
          const response = await request.get(new URL(path, site.url).toString());
          expect(
            [401, 403],
            `${path} unexpectedly returned HTTP ${response.status()}`,
          ).toContain(response.status());
        }
      });

      test("普通会员无法进入管理页面或调用管理 API", async ({ browser }) => {
        const context = await ensureRegisteredE2EIdentity(
          browser,
          {
            email: "resource-member-acceptance@example.test",
            password: "Resource-member-acceptance-2026!",
          },
          "Resource Acceptance Member",
          site.url,
        );
        const page = await context.newPage();
        try {
          await page.goto(new URL("/manage", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await expect(page).toHaveURL(new URL("/", site.url).toString());
          await settleNuxt(page);

          for (const path of [
            "/api/resource/admin/overview",
            "/api/resource/admin/site-settings",
            "/api/resource/admin/images?size=1",
            "/api/resource/admin/collections",
          ]) {
            const response = await context.request.get(
              new URL(path, site.url).toString(),
            );
            expect(
              response.status(),
              `${path} unexpectedly returned HTTP ${response.status()}`,
            ).toBe(403);
          }
        } finally {
          await context.close();
        }
      });

      test("管理员可无 5xx 与浏览器错误遍历全部管理入口", async ({
        browser,
      }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        const failures = capturePageFailures(page);
        try {
          for (const path of managementRoutes) {
            const url = new URL(path, site.url).toString();
            const response = await page.goto(url, {
              waitUntil: "domcontentloaded",
            });
            expect(response?.ok(), `${path} returned ${response?.status()}`).toBeTruthy();
            await expect(page).toHaveURL(url);
            await expect(page.locator("main h1").first()).toBeVisible();
            await settleNuxt(page);
          }
          expect(failures).toEqual([]);
        } finally {
          await context.close();
        }
      });

      test("站点设置保存后持久化且测试结束恢复原值", async ({ browser }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        let originalName = "";
        try {
          await page.goto(new URL("/manage/discovery", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          const field = page
            .getByRole("textbox", { name: "站点名称", exact: false })
            .first();
          originalName = await field.inputValue();
          const temporaryName = `${originalName.slice(0, 65)} · 验收`;
          await field.fill(temporaryName);
          await saveDiscoverySettings(page);

          await page.reload({ waitUntil: "domcontentloaded" });
          await settleNuxt(page);
          await expect(
            page
              .getByRole("textbox", { name: "站点名称", exact: false })
              .first(),
          ).toHaveValue(temporaryName);

          await page
            .getByRole("textbox", { name: "站点名称", exact: false })
            .first()
            .fill(originalName);
          await saveDiscoverySettings(page);
          await page.reload({ waitUntil: "domcontentloaded" });
          await settleNuxt(page);
          await expect(
            page
              .getByRole("textbox", { name: "站点名称", exact: false })
              .first(),
          ).toHaveValue(originalName);
        } finally {
          if (originalName)
            await restoreDiscoveryName(page, site.url, originalName).catch(
              () => undefined,
            );
          await context.close();
        }
      });

      test("图片编辑器使用可读分类标签且选择器有明确名称", async ({
        browser,
      }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        try {
          await page.goto(new URL("/manage/images", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          await openFirstImageEditor(page);

          const dialog = page.getByRole("dialog", { name: "编辑图片" });
          const category = dialog.getByRole("button", {
            name: "主分类",
            exact: true,
          });
          const tags = dialog.getByRole("button", {
            name: "标签",
            exact: true,
          });
          await expect(category).toBeVisible();
          await expect(tags).toBeVisible();

          const rawID =
            /(?:^[A-Za-z0-9_-]{22}$)|(?:^[0-9a-f]{8}-[0-9a-f-]{27}$)/iu;
          expect((await category.textContent())?.trim()).not.toMatch(rawID);
          expect((await tags.textContent())?.trim()).not.toMatch(rawID);

          await category.click();
          const categoryLabels = (
            await page.getByRole("option").allTextContents()
          ).map((value) => value.trim());
          expect(categoryLabels).toEqual(
            expect.arrayContaining(["壁纸", "插画", "摄影"]),
          );
        } finally {
          await context.close();
        }
      });

      test("图片标题保存后可检索且测试结束恢复原值", async ({ browser }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        let originalTitle = "";
        let temporaryTitle = "";
        try {
          await page.goto(new URL("/manage/images", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          await openFirstImageEditor(page);
          const title = page
            .getByRole("dialog", { name: "编辑图片" })
            .getByRole("textbox", { name: "标题", exact: false });
          originalTitle = await title.inputValue();
          temporaryTitle = `${originalTitle.slice(0, 140)} · 验收`;
          await title.fill(temporaryTitle);
          await saveImageEditor(page);

          await searchManagedImages(page, temporaryTitle);
          await openFirstImageEditor(page);
          await expect(
            page
              .getByRole("dialog", { name: "编辑图片" })
              .getByRole("textbox", { name: "标题", exact: false }),
          ).toHaveValue(temporaryTitle);
          await page
            .getByRole("dialog", { name: "编辑图片" })
            .getByRole("textbox", { name: "标题", exact: false })
            .fill(originalTitle);
          await saveImageEditor(page);

          await searchManagedImages(page, originalTitle);
          await expect(page.getByText(originalTitle, { exact: true }).first()).toBeVisible();
        } finally {
          if (originalTitle && temporaryTitle)
            await restoreImageTitle(
              page,
              site.url,
              temporaryTitle,
              originalTitle,
            ).catch(() => undefined);
          await context.close();
        }
      });

      test("专题名称保存后刷新持久化且测试结束恢复原值", async ({
        browser,
      }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        let originalName = "";
        let collectionURL = "";
        try {
          await page.goto(
            new URL("/manage/collections", site.url).toString(),
            { waitUntil: "domcontentloaded" },
          );
          await settleNuxt(page);
          const firstCollection = page
            .getByRole("link", { name: /^编辑专题：/u })
            .first();
          await expect(
            firstCollection,
            "验收夹具必须至少包含一个可编辑专题",
          ).toBeVisible();
          const href = await firstCollection.getAttribute("href");
          expect(href).toBeTruthy();
          collectionURL = new URL(href!, site.url).toString();

          await page.goto(collectionURL, { waitUntil: "domcontentloaded" });
          await settleNuxt(page);
          const name = page
            .getByRole("textbox", { name: "名称", exact: false })
            .first();
          originalName = await name.inputValue();
          const temporaryName = `${originalName.slice(0, 90)} · 验收`;
          await name.fill(temporaryName);
          await saveCollectionSettings(page);

          await page.reload({ waitUntil: "domcontentloaded" });
          await settleNuxt(page);
          await expect(
            page
              .getByRole("textbox", { name: "名称", exact: false })
              .first(),
          ).toHaveValue(temporaryName);

          await page
            .getByRole("textbox", { name: "名称", exact: false })
            .first()
            .fill(originalName);
          await saveCollectionSettings(page);
          await page.reload({ waitUntil: "domcontentloaded" });
          await settleNuxt(page);
          await expect(
            page
              .getByRole("textbox", { name: "名称", exact: false })
              .first(),
          ).toHaveValue(originalName);
        } finally {
          if (collectionURL && originalName)
            await restoreCollectionName(
              page,
              collectionURL,
              originalName,
            ).catch(() => undefined);
          await context.close();
        }
      });

      test("资源站点名称保存后刷新持久化且测试结束恢复原值", async ({
        browser,
      }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        let originalName = "";
        try {
          await page.goto(new URL("/manage/assets", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          const name = page
            .getByRole("textbox", { name: "站点名称", exact: false })
            .first();
          originalName = await name.inputValue();
          const temporaryName = `${originalName.slice(0, 90)} · 验收`;
          await name.fill(temporaryName);
          await saveAssetSettings(page);

          await page.reload({ waitUntil: "domcontentloaded" });
          await settleNuxt(page);
          await expect(
            page
              .getByRole("textbox", { name: "站点名称", exact: false })
              .first(),
          ).toHaveValue(temporaryName);

          await page
            .getByRole("textbox", { name: "站点名称", exact: false })
            .first()
            .fill(originalName);
          await saveAssetSettings(page);
          await page.reload({ waitUntil: "domcontentloaded" });
          await settleNuxt(page);
          await expect(
            page
              .getByRole("textbox", { name: "站点名称", exact: false })
              .first(),
          ).toHaveValue(originalName);
        } finally {
          if (originalName)
            await restoreAssetSiteName(page, site.url, originalName).catch(
              () => undefined,
            );
          await context.close();
        }
      });

      test("分类状态变更必须先生成影响预览且不会直接改写数据", async ({
        browser,
      }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        try {
          await page.goto(
            new URL("/manage/classification", site.url).toString(),
            { waitUntil: "domcontentloaded" },
          );
          await settleNuxt(page);
          const firstActions = page
            .getByRole("button", { name: /^更多分类操作：/u })
            .first();
          await expect(
            firstActions,
            "验收夹具必须至少包含一个可治理分类",
          ).toBeVisible();
          await firstActions.click();

          const previewResponse = page.waitForResponse(
            (candidate) =>
              candidate.request().method() === "POST" &&
              /\/api\/resource\/admin\/classification\/governance\/preview$/.test(
                candidate.url(),
              ),
          );
          await page.getByRole("menuitem", { name: "停用分类" }).click();
          expect((await previewResponse).ok()).toBeTruthy();

          const dialog = page.getByRole("dialog", { name: "治理影响预览" });
          await expect(dialog).toBeVisible();
          await expect(dialog.getByText("planned", { exact: true })).toBeVisible();
          await expect(
            dialog.getByRole("button", { name: "执行计划", exact: true }),
          ).toBeVisible();
          await dialog.getByRole("button", { name: "关闭", exact: true }).click();
          await expect(dialog).toHaveCount(0);
        } finally {
          await context.close();
        }
      });
    });
  }
}
