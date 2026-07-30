import AxeBuilder from "@axe-core/playwright";
import { expect, test, type Page } from "@playwright/test";
import { productSites } from "./contracts";
import { loginE2E, settleNuxt } from "./runtime";

const wcagTags = [
  "wcag2a",
  "wcag2aa",
  "wcag21a",
  "wcag21aa",
  "wcag22a",
  "wcag22aa",
];

async function settle(page: Page) {
  await page.emulateMedia({ reducedMotion: "reduce" });
  await settleNuxt(page);
  await page.addStyleTag({
    content:
      "nuxt-devtools-frame, nuxt-devtools-client, vue-devtools, #nuxt-devtools-container, [data-nuxt-devtools] { display: none !important; }",
  });
}

async function assertWcag(page: Page, label: string) {
  const result = await new AxeBuilder({ page })
    .exclude("nuxt-devtools-frame")
    .withTags(wcagTags)
    .analyze();
  expect(result.violations, `${label} WCAG 2.2 AA violations`).toEqual([]);
}

async function resetKeyboardFocus(page: Page) {
  await page.evaluate(() => {
    document.body.tabIndex = -1;
    document.body.focus();
    document.body.removeAttribute("tabindex");
  });
}

async function assertKeyboardFocus(page: Page, label: string) {
  await resetKeyboardFocus(page);
  const focused = new Set<string>();
  for (let attempt = 0; attempt < 16 && focused.size < 3; attempt += 1) {
    await page.keyboard.press("Tab");
    const state = await page.evaluate(() => {
      const element = document.activeElement as HTMLElement | null;
      if (!element || element === document.body) return null;
      const snapshot = () => {
        const style = getComputedStyle(element);
        return {
          outlineColor: style.outlineColor,
          outlineStyle: style.outlineStyle,
          outlineWidth: style.outlineWidth,
          boxShadow: style.boxShadow,
        };
      };
      const hasVisibleColor = (value: string) => {
        if (!value || value === "none" || value === "transparent") return false;
        const colors =
          value.match(
            /(?:rgba?|hsla?|oklab|oklch|lab|lch|color)\([^)]*\)|#[\da-f]{3,8}/gi,
          ) || [];
        return colors.some((color) => {
          if (color.length === 9)
            return Number.parseInt(color.slice(7), 16) > 2;
          if (color.length === 5)
            return Number.parseInt(color.slice(4), 16) > 0;
          const body = color.slice(color.indexOf("(") + 1, -1);
          const slashAlpha = body.match(/\/\s*([\d.]+)%?\s*$/);
          if (slashAlpha) {
            const alpha = Number.parseFloat(slashAlpha[1] || "0");
            return body.includes("%", body.lastIndexOf("/"))
              ? alpha > 1
              : alpha > 0.01;
          }
          if (/^rgba|^hsla/i.test(color)) {
            const channels = body.split(",");
            return (
              channels.length < 4 ||
              Number.parseFloat(channels[3] || "1") > 0.01
            );
          }
          return !color.toLowerCase().includes("transparent");
        });
      };
      const hasShadowGeometry = (value: string) => {
        const withoutColors = value.replace(
          /(?:rgba?|hsla?|oklab|oklch|lab|lch|color)\([^)]*\)|#[\da-f]{3,8}/gi,
          "",
        );
        return (withoutColors.match(/-?[\d.]+px/g) || []).some(
          (length) => Math.abs(Number.parseFloat(length)) > 0.01,
        );
      };
      const shadowLayers = (value: string) => {
        if (!value || value === "none") return [];
        const layers: string[] = [];
        let depth = 0;
        let start = 0;
        for (let index = 0; index < value.length; index += 1) {
          const character = value[index];
          if (character === "(") depth += 1;
          else if (character === ")") depth -= 1;
          else if (character === "," && depth === 0) {
            layers.push(value.slice(start, index).trim());
            start = index + 1;
          }
        }
        layers.push(value.slice(start).trim());
        return layers.filter(Boolean);
      };
      const focused = snapshot();
      const focusVisible = element.matches(":focus-visible");
      const previousBodyTabIndex = document.body.getAttribute("tabindex");
      document.body.tabIndex = -1;
      document.body.focus({ preventScroll: true });
      const unfocused = snapshot();
      element.focus({ preventScroll: true });
      if (previousBodyTabIndex === null)
        document.body.removeAttribute("tabindex");
      else document.body.setAttribute("tabindex", previousBodyTabIndex);

      const outlineVisible =
        focused.outlineStyle !== "none" &&
        Number.parseFloat(focused.outlineWidth) > 0.01 &&
        hasVisibleColor(focused.outlineColor) &&
        (focused.outlineStyle !== unfocused.outlineStyle ||
          focused.outlineWidth !== unfocused.outlineWidth ||
          focused.outlineColor !== unfocused.outlineColor);
      const unfocusedShadows = shadowLayers(unfocused.boxShadow);
      const shadowVisible = shadowLayers(focused.boxShadow).some(
        (layer) =>
          !unfocusedShadows.includes(layer) &&
          hasVisibleColor(layer) &&
          hasShadowGeometry(layer),
      );
      const name =
        element.getAttribute("aria-label") ||
        element.getAttribute("title") ||
        element.textContent?.trim() ||
        element.getAttribute("placeholder") ||
        element.tagName.toLowerCase();
      return {
        key: `${element.tagName}:${name}`,
        focusVisible,
        indicatorVisible: outlineVisible || shadowVisible,
      };
    });
    if (!state) continue;
    expect(
      state.focusVisible,
      `${label}: ${state.key} is not keyboard focus-visible`,
    ).toBeTruthy();
    expect(
      state.indicatorVisible,
      `${label}: ${state.key} has no visible outline or focus ring`,
    ).toBeTruthy();
    focused.add(state.key);
  }
  expect(
    focused.size,
    `${label}: keyboard did not reach three distinct controls`,
  ).toBeGreaterThanOrEqual(3);

  await page.evaluate(() =>
    document.activeElement?.setAttribute("data-e2e-focus-before", ""),
  );
  await page.keyboard.press("Shift+Tab");
  const reverse = await page.evaluate(() => ({
    tag: document.activeElement?.tagName || "",
    stayed:
      document.activeElement?.hasAttribute("data-e2e-focus-before") || false,
  }));
  expect(reverse.tag, `${label}: reverse tab navigation lost focus`).not.toBe(
    "BODY",
  );
  expect(
    reverse.stayed,
    `${label}: reverse tab navigation did not move`,
  ).toBeFalsy();
}

export function registerAccessibilitySuite(product: string) {
  for (const site of productSites(product)) {
    const contract = site.contract;
    test.describe(`${site.slug} (${site.product}) accessibility contract`, () => {
      test("public page passes WCAG 2.2 AA on mobile and desktop", async ({
        page,
      }) => {
        for (const viewport of [
          { name: "mobile", width: 390, height: 844 },
          { name: "desktop", width: 1440, height: 900 },
        ]) {
          await page.setViewportSize(viewport);
          const response = await page.goto(
            new URL(contract.public.path, site.url).toString(),
            { waitUntil: "domcontentloaded" },
          );
          expect(response?.ok()).toBeTruthy();
          await settle(page);
          await assertWcag(page, `${site.slug} public ${viewport.name}`);
        }
      });

      test("authenticated manage page passes WCAG 2.2 AA", async ({
        browser,
      }) => {
        const context = await loginE2E(browser);
        try {
          const page = await context.newPage();
          await page.goto(new URL(contract.manage.path, site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settle(page);
          await expect(
            page
              .getByRole("heading", { name: contract.manage.heading })
              .first(),
          ).toBeVisible();
          await assertWcag(page, `${site.slug} manage`);
        } finally {
          await context.close();
        }
      });

      test("public and manage controls support visible bidirectional keyboard focus", async ({
        browser,
        page,
      }) => {
        await page.goto(new URL(contract.public.path, site.url).toString(), {
          waitUntil: "domcontentloaded",
        });
        await settle(page);
        await assertKeyboardFocus(page, `${site.slug} public`);

        const context = await loginE2E(browser);
        try {
          const managePage = await context.newPage();
          await managePage.goto(
            new URL(contract.manage.path, site.url).toString(),
            { waitUntil: "domcontentloaded" },
          );
          await settle(managePage);
          await expect(
            managePage
              .getByRole("heading", { name: contract.manage.heading })
              .first(),
          ).toBeVisible();
          await assertKeyboardFocus(managePage, `${site.slug} manage`);
        } finally {
          await context.close();
        }
      });
    });
  }
}
