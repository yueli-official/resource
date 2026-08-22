import {
  expect,
  type Browser,
  type BrowserContext,
  type BrowserContextOptions,
  type Page,
} from "@playwright/test";

export function requiredEnv(name: string): string {
  const value = process.env[name]?.trim();
  if (!value) throw new Error(`${name} is required`);
  return value;
}

export async function expectNoHorizontalOverflow(
  page: Page,
  label = "页面",
) {
  const measurement = await page.evaluate(() => {
    const viewportWidth = window.innerWidth;
    const documentWidth = Math.max(
      document.documentElement.scrollWidth,
      document.body.scrollWidth,
    );
    const culprits = [...document.querySelectorAll<HTMLElement>("body *")]
      .filter((element) => {
        const style = window.getComputedStyle(element);
        const bounds = element.getBoundingClientRect();
        return (
          style.display !== "none" &&
          style.visibility !== "hidden" &&
          bounds.width > 0 &&
          bounds.height > 0 &&
          (bounds.right > viewportWidth + 1 || bounds.left < -1)
        );
      })
      .slice(0, 5)
      .map((element) => {
        const bounds = element.getBoundingClientRect();
        const name =
          element.id ||
          [...element.classList].slice(0, 3).join(".") ||
          element.tagName.toLowerCase();
        return `${name} [${bounds.left.toFixed(1)}, ${bounds.right.toFixed(1)}]`;
      });
    return { viewportWidth, documentWidth, culprits };
  });

  expect(
    measurement.documentWidth,
    `${label}横向溢出：viewport=${measurement.viewportWidth}px, document=${measurement.documentWidth}px；${measurement.culprits.join("; ")}`,
  ).toBeLessThanOrEqual(measurement.viewportWidth + 1);
}

export interface E2ECredentials {
  email: string;
  password: string;
}

export async function loginE2E(
  browser: Browser,
  options: BrowserContextOptions = {},
  colorMode?: "light" | "dark",
  siteURL?: string,
): Promise<BrowserContext> {
  return loginE2EWithCredentials(
    browser,
    {
      email: requiredEnv("RESOURCE_E2E_EMAIL"),
      password: requiredEnv("RESOURCE_E2E_PASSWORD"),
    },
    options,
    colorMode,
    siteURL,
  );
}

export async function loginE2EWithCredentials(
  browser: Browser,
  credentials: E2ECredentials,
  options: BrowserContextOptions = {},
  colorMode?: "light" | "dark",
  siteURL?: string,
): Promise<BrowserContext> {
  const context = await browser.newContext(options);
  if (colorMode) {
    await context.addInitScript((value) => {
      window.localStorage.setItem("nuxt-color-mode", value);
    }, colorMode);
  }
  const response = await context.request.post(
    `${requiredEnv("RESOURCE_E2E_IDENTITY_URL")}/api/v1/auth/login`,
    {
      data: {
        email: credentials.email,
        password: credentials.password,
      },
    },
  );
  expect(
    response.ok(),
    `identity login failed with HTTP ${response.status()}`,
  ).toBeTruthy();
  if (siteURL) {
    const sessionResponse = await context.request.get(
      new URL("/auth/login?return_to=/", siteURL).toString(),
    );
    expect(
      sessionResponse.ok(),
      `site session bootstrap failed with HTTP ${sessionResponse.status()}`,
    ).toBeTruthy();
  }
  return context;
}

export async function ensureRegisteredE2EIdentity(
  browser: Browser,
  credentials: E2ECredentials,
  displayName: string,
  siteURL: string,
): Promise<BrowserContext> {
  const initialContext = await browser.newContext();
  const identityURL = requiredEnv("RESOURCE_E2E_IDENTITY_URL");
  const initialLogin = await initialContext.request.post(
    `${identityURL}/api/v1/auth/login`,
    { data: credentials },
  );
  if (initialLogin.ok()) {
    const sessionResponse = await initialContext.request.get(
      new URL("/auth/login?return_to=/", siteURL).toString(),
    );
    expect(
      sessionResponse.ok(),
      `site session bootstrap failed with HTTP ${sessionResponse.status()}`,
    ).toBeTruthy();
    return initialContext;
  }
  await initialContext.close();

  const registrationContext = await browser.newContext();
  const registration = await registrationContext.request.post(
    `${identityURL}/api/v1/auth/register`,
    {
      data: {
        ...credentials,
        displayName,
      },
    },
  );
  expect(
    registration.ok(),
    `identity registration failed with HTTP ${registration.status()}`,
  ).toBeTruthy();
  await registrationContext.close();
  return loginE2EWithCredentials(
    browser,
    credentials,
    {},
    undefined,
    siteURL,
  );
}

export async function settleNuxt(page: Page) {
  await page.waitForLoadState("load", { timeout: 30_000 });
  await page.waitForLoadState("networkidle", { timeout: 30_000 });
  await page.waitForFunction(
    async () => {
      const root = document.querySelector("#__nuxt");
      if (!root || !("__vue_app__" in root)) return false;
      await document.fonts.ready;
      await new Promise<void>((resolve) =>
        requestAnimationFrame(() => requestAnimationFrame(() => resolve())),
      );
      const settledRoot = document.querySelector("#__nuxt");
      return Boolean(settledRoot && "__vue_app__" in settledRoot);
    },
    undefined,
    { timeout: 30_000 },
  );
}

export function capturePageFailures(page: Page): string[] {
  const failures: string[] = [];
  page.on("console", (message) => {
    if (message.type() === "error")
      failures.push(`console: ${message.text()}`);
  });
  page.on("pageerror", (error) =>
    failures.push(`pageerror: ${error.message}`),
  );
  page.on("response", (response) => {
    if (response.status() >= 400)
      failures.push(`http ${response.status()}: ${response.url()}`);
  });
  return failures;
}
