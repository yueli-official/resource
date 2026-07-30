import fs from "node:fs";
import path from "node:path";
import { defineConfig } from "@playwright/test";

const root = process.cwd();
const e2eRoot = path.resolve(root, "test/e2e");
const contract = JSON.parse(
  fs.readFileSync(path.join(e2eRoot, "contract.json"), "utf8"),
);
const localHost = process.env.LOCAL_LAN_HOST?.trim() || "127.0.0.1";
const identityPort = process.env.LOCAL_IDENTITY_PORT?.trim() || "8081";
const accountPort = process.env.LOCAL_ACCOUNT_PORT?.trim() || "3000";
const resourcePort = process.env.LOCAL_RESOURCE_WEB_PORT?.trim() || "3001";
const resourceURL =
  process.env.RESOURCE_E2E_URL?.trim() ||
  `http://${localHost}:${resourcePort}`;
process.env.RESOURCE_E2E_SITES ||= JSON.stringify([
  {
    slug: "resource-main",
    product: "resource",
    url: resourceURL,
    contract,
  },
]);
process.env.RESOURCE_E2E_IDENTITY_URL ||=
  process.env.IDENTITY_BASE_URL?.trim() ||
  `http://${localHost}:${identityPort}`;
process.env.RESOURCE_E2E_ACCOUNT_URL ||=
  process.env.NUXT_PUBLIC_ACCOUNT_URL?.trim() ||
  `http://${localHost}:${accountPort}`;

const channel = process.env.RESOURCE_E2E_BROWSER_CHANNEL || undefined;
const suite = process.env.RESOURCE_E2E_SUITE?.trim() || "all";
const inCI = Boolean(process.env.CI?.trim());
const supportedSuites = new Set([
  "all",
  "journeys",
  "management",
  "security",
  "resilience",
  "visual",
  "accessibility",
  "performance",
  "responsive",
]);
if (!supportedSuites.has(suite))
  throw new Error(`不支持的 Resource E2E suite：${suite}`);

const runID =
  process.env.RESOURCE_E2E_RUN_ID?.trim() ||
  new Date().toISOString().replace(/[:.]/g, "-");

export default defineConfig({
  testDir: e2eRoot,
  testMatch: "product.spec.ts",
  fullyParallel: false,
  workers: 1,
  retries: inCI ? 1 : 0,
  timeout: 90_000,
  globalTimeout: 25 * 60_000,
  expect: {
    timeout: 30_000,
    toHaveScreenshot: {
      animations: "disabled",
      caret: "hide",
      maxDiffPixelRatio: 0.02,
      threshold: 0.2,
    },
  },
  outputDir: path.resolve(root, "test-results/e2e", runID, "artifacts"),
  snapshotPathTemplate: path.resolve(e2eRoot, "{arg}{ext}"),
  reporter: [
    ["line"],
    [path.resolve(e2eRoot, "evidence-reporter.ts")],
    [
      "junit",
      {
        outputFile: path.resolve(
          root,
          "test-results/e2e",
          runID,
          "junit.xml",
        ),
      },
    ],
  ],
  use: {
    headless: true,
    actionTimeout: 30_000,
    navigationTimeout: 90_000,
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
    launchOptions: channel ? { channel } : undefined,
  },
});
