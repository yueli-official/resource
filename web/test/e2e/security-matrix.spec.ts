import { expect, test, type Page } from "@playwright/test";
import { productSites } from "./contracts";
import { loginE2E, settleNuxt } from "./runtime";

const browserSessionCookies = new Set([
  "rs_session",
  "yueli_guest",
  "__Host-yueli_guest",
]);
const jwtPattern =
  /\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b/u;

async function browserVisibleState(page: Page) {
  return page.evaluate(() => ({
    url: window.location.href,
    html: document.documentElement.outerHTML,
    localStorage: Object.fromEntries(
      Array.from({ length: window.localStorage.length }, (_, index) => {
        const key = window.localStorage.key(index) || "";
        return [key, window.localStorage.getItem(key) || ""];
      }),
    ),
    sessionStorage: Object.fromEntries(
      Array.from({ length: window.sessionStorage.length }, (_, index) => {
        const key = window.sessionStorage.key(index) || "";
        return [key, window.sessionStorage.getItem(key) || ""];
      }),
    ),
  }));
}

function expectNoBrowserVisibleToken(
  state: Awaited<ReturnType<typeof browserVisibleState>>,
) {
  expect(state.url).not.toMatch(jwtPattern);
  expect(state.html).not.toMatch(jwtPattern);
  expect(JSON.stringify(state.localStorage)).not.toMatch(jwtPattern);
  expect(JSON.stringify(state.sessionStorage)).not.toMatch(jwtPattern);
}

export function registerSecuritySuite(product: string) {
  for (const site of productSites(product)) {
    test.describe(`${site.slug} (${site.product}) security contract`, () => {
      test("公开读取不创建 Guest，私人入口按需创建安全 Guest Cookie", async ({
        browser,
      }) => {
        const context = await browser.newContext();
        const page = await context.newPage();
        try {
          await page.goto(new URL("/", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          expectNoBrowserVisibleToken(await browserVisibleState(page));
          expect(
            (await context.cookies(site.url)).filter((cookie) =>
              browserSessionCookies.has(cookie.name),
            ),
          ).toEqual([]);

          const submissionsResponse = page.waitForResponse(
            (response) =>
              new URL(response.url()).pathname ===
              "/api/resource/me/submissions",
          );
          await page.goto(new URL("/submissions", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          expect((await submissionsResponse).status()).toBe(200);
          const guestCookie = (await context.cookies(site.url)).find(
            (cookie) =>
              cookie.name === "yueli_guest" ||
              cookie.name === "__Host-yueli_guest",
          );
          expect(guestCookie).toMatchObject({
            httpOnly: true,
            sameSite: "Lax",
            path: "/",
          });
          expectNoBrowserVisibleToken(await browserVisibleState(page));
        } finally {
          await context.close();
        }
      });

      test("登录令牌只停留在 HttpOnly 会话，浏览器不直传 Bearer", async ({
        browser,
      }) => {
        const context = await loginE2E(browser, {}, undefined, site.url);
        const page = await context.newPage();
        const browserAuthorizations: string[] = [];
        page.on("request", (request) => {
          if (new URL(request.url()).origin !== new URL(site.url).origin)
            return;
          const authorization = request.headers().authorization;
          if (authorization) browserAuthorizations.push(authorization);
        });
        try {
          const sessionCookie = (await context.cookies(site.url)).find(
            (cookie) => cookie.name === "rs_session",
          );
          expect(sessionCookie).toMatchObject({
            httpOnly: true,
            sameSite: "Lax",
            path: "/",
          });

          await page.goto(new URL("/manage", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          expectNoBrowserVisibleToken(await browserVisibleState(page));
          expect(browserAuthorizations).toEqual([]);
        } finally {
          await context.close();
        }
      });

      test("BFF 对异常目标路径和危险方法失败关闭且不泄漏内部细节", async ({
        request,
      }) => {
        for (const candidate of [
          "/api/resource/http:%2F%2F127.0.0.1:8081%2Freadyz",
          "/api/resource/%252e%252e/%252e%252e/readyz",
        ]) {
          const response = await request.get(
            new URL(candidate, site.url).toString(),
          );
          expect(
            response.status(),
            `${candidate} unexpectedly returned HTTP ${response.status()}`,
          ).toBeGreaterThanOrEqual(400);
          const body = await response.text();
          expect(body).not.toContain("127.0.0.1");
          expect(body).not.toMatch(/\b(?:postgres|redis|SELECT|INSERT)\b/iu);
          expect(body).not.toMatch(/\b(?:goroutine|stack trace|node_modules)\b/iu);
        }

        const trace = await request.fetch(
          new URL("/api/resource/images", site.url).toString(),
          { method: "TRACE" },
        );
        expect(trace.status()).toBeGreaterThanOrEqual(400);
      });
    });
  }
}
