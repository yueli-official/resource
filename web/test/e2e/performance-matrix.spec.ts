import { expect, test, type Page } from "@playwright/test";
import { productSites } from "./contracts";

interface ImageMeasurement {
  url: string;
  aliases: string[];
  bytes: number;
  error?: string;
}

const defaultLimits = {
  lcpMilliseconds: 2_500,
  cls: 0.1,
  ttfbMilliseconds: 1_000,
  imageBytes: 600 * 1024,
  totalImageBytes: 6 * 1024 * 1024,
  imageDimension: 2_560,
};

const productLimits: Record<string, typeof defaultLimits> = {
  resource: {
    lcpMilliseconds: 2_200,
    cls: 0.05,
    ttfbMilliseconds: 800,
    imageBytes: 400 * 1024,
    totalImageBytes: 2 * 1024 * 1024,
    imageDimension: 1_600,
  },
};

async function installWebVitalObservers(page: Page) {
  await page.addInitScript(() => {
    const target = window as unknown as {
      __platformPerformance: {
        lcp: number;
        cls: number;
        clsWindow: number;
        clsWindowStart: number;
        lastShift: number;
      };
    };
    target.__platformPerformance = {
      lcp: 0,
      cls: 0,
      clsWindow: 0,
      clsWindowStart: -1,
      lastShift: 0,
    };
    new PerformanceObserver((list) => {
      for (const entry of list.getEntries())
        target.__platformPerformance.lcp = Math.max(
          target.__platformPerformance.lcp,
          entry.startTime,
        );
    }).observe({ type: "largest-contentful-paint", buffered: true });
    new PerformanceObserver((list) => {
      for (const entry of list.getEntries() as Array<
        PerformanceEntry & { hadRecentInput: boolean; value: number }
      >) {
        if (entry.hadRecentInput) continue;
        const state = target.__platformPerformance;
        if (
          state.clsWindowStart < 0 ||
          entry.startTime - state.lastShift > 1_000 ||
          entry.startTime - state.clsWindowStart > 5_000
        ) {
          state.clsWindowStart = entry.startTime;
          state.clsWindow = entry.value;
        } else {
          state.clsWindow += entry.value;
        }
        state.lastShift = entry.startTime;
        state.cls = Math.max(state.cls, state.clsWindow);
      }
    }).observe({ type: "layout-shift", buffered: true });
  });
}

async function settle(page: Page) {
  await page.waitForLoadState("networkidle", { timeout: 15_000 });
  await page.evaluate(() => document.fonts.ready);
  await page.waitForTimeout(500);
}

export function registerPerformanceSuite(product: string) {
  const limits = productLimits[product] || defaultLimits;
  for (const site of productSites(product)) {
    const contract = site.contract;
    test(`${site.slug} (${site.product}) stays within the public-page performance budget`, async ({
      page,
    }) => {
      await page.setViewportSize({ width: 1440, height: 900 });
      await installWebVitalObservers(page);

      // 先预热开发渲染器和浏览器缓存；随后测量复访性能，降低共享 CI 执行器的波动。
      await page.goto(new URL(contract.public.path, site.url).toString(), {
        waitUntil: "domcontentloaded",
      });
      await settle(page);

      const imageMeasurements: Array<Promise<ImageMeasurement>> = [];
      let collectImages = false;
      page.on("response", (response) => {
        if (
          !collectImages ||
          response.request().resourceType() !== "image" ||
          !response.ok()
        )
          return;
        imageMeasurements.push(
          response
            .body()
            .then((body) => {
              const aliases: string[] = [];
              let redirectedFrom = response.request().redirectedFrom();
              while (redirectedFrom) {
                aliases.push(redirectedFrom.url());
                redirectedFrom = redirectedFrom.redirectedFrom();
              }
              return {
                url: response.url(),
                aliases,
                bytes: body.byteLength,
              };
            })
            .catch((error) => ({
              url: response.url(),
              aliases: [],
              bytes: 0,
              error: String(error),
            })),
        );
      });

      collectImages = true;
      const response = await page.goto(
        new URL(contract.public.path, site.url).toString(),
        {
          waitUntil: "domcontentloaded",
        },
      );
      expect(response?.ok()).toBeTruthy();
      await settle(page);
      collectImages = false;
      const images = await Promise.all(imageMeasurements);

      const metrics = await page.evaluate(async () => {
        const target = window as unknown as {
          __platformPerformance?: { lcp: number; cls: number };
        };
        const navigation = performance.getEntriesByType("navigation")[0] as
          PerformanceNavigationTiming | undefined;
        return {
          lcp: target.__platformPerformance?.lcp || 0,
          cls: target.__platformPerformance?.cls || 0,
          ttfb: navigation?.responseStart || 0,
          images: await Promise.all(
            Array.from(document.images)
              .filter(
                (image) =>
                  image.currentSrc &&
                  image.getBoundingClientRect().width > 1 &&
                  image.getBoundingClientRect().height > 1,
              )
              .map(async (image) => ({
                url: image.currentSrc,
                naturalWidth: image.naturalWidth,
                naturalHeight: image.naturalHeight,
                inlineBytes: /^(?:data|blob):/.test(image.currentSrc)
                  ? await fetch(image.currentSrc)
                      .then((response) => response.blob())
                      .then((blob) => blob.size)
                  : undefined,
              })),
          ),
        };
      });

      await test.info().attach(`${site.slug}-performance.json`, {
        body: JSON.stringify(
          { site: site.slug, limits, metrics, transferredImages: images },
          null,
          2,
        ),
        contentType: "application/json",
      });

      expect(
        metrics.lcp,
        "Largest Contentful Paint observer returned no entry",
      ).toBeGreaterThan(0);
      expect(
        metrics.lcp,
        `LCP exceeds ${limits.lcpMilliseconds} ms`,
      ).toBeLessThanOrEqual(limits.lcpMilliseconds);
      expect(metrics.cls, `CLS exceeds ${limits.cls}`).toBeLessThanOrEqual(
        limits.cls,
      );
      expect(
        metrics.ttfb,
        `server response exceeds ${limits.ttfbMilliseconds} ms`,
      ).toBeLessThanOrEqual(limits.ttfbMilliseconds);

      for (const image of images) {
        expect(
          image.error,
          `cannot read ${image.url} for the image byte budget`,
        ).toBeUndefined();
        expect(
          image.bytes,
          `${image.url} exceeds the transferred image budget`,
        ).toBeLessThanOrEqual(limits.imageBytes);
      }
      expect(
        images.reduce((total, image) => total + image.bytes, 0),
        `visible images exceed the ${limits.totalImageBytes}-byte page budget`,
      ).toBeLessThanOrEqual(limits.totalImageBytes);
      for (const image of metrics.images) {
        if (/^(?:data|blob):/.test(image.url)) {
          expect(
            image.inlineBytes,
            `cannot measure inline image ${image.url.slice(0, 80)}`,
          ).toBeGreaterThanOrEqual(0);
          expect(
            image.inlineBytes,
            "inline image exceeds the transferred image budget",
          ).toBeLessThanOrEqual(limits.imageBytes);
        } else {
          const normalized = new URL(image.url);
          normalized.hash = "";
          expect(
            images.some((measurement) => {
              const candidates = [
                measurement.url,
                ...measurement.aliases,
              ].map((url) => {
                const candidate = new URL(url);
                candidate.hash = "";
                return candidate.href;
              });
              return (
                candidates.includes(normalized.href) && !measurement.error
              );
            }),
            `visible image ${image.url} has no trustworthy response-body measurement`,
          ).toBeTruthy();
        }
        expect(
          image.naturalWidth,
          `${image.url} is wider than the source-image budget`,
        ).toBeLessThanOrEqual(limits.imageDimension);
        expect(
          image.naturalHeight,
          `${image.url} is taller than the source-image budget`,
        ).toBeLessThanOrEqual(limits.imageDimension);
      }
    });
  }
}
