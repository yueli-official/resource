import fs from "node:fs";
import path from "node:path";
import type {
  FullResult,
  Reporter,
  TestCase,
  TestResult,
} from "@playwright/test/reporter";
import type { SiteContract } from "./contracts";

export default class EvidenceReporter implements Reporter {
  private readonly root = process.cwd();
  private readonly runID =
    process.env.RESOURCE_E2E_RUN_ID?.trim() || "local";
  private readonly sites = JSON.parse(
    process.env.RESOURCE_E2E_SITES || "[]",
  ) as SiteContract[];

  onTestEnd(test: TestCase, result: TestResult) {
    if (result.status === "passed" || result.status === "skipped") return;
    const site = this.sites.find((candidate) =>
      test.titlePath().some((title) => title.includes(candidate.slug)),
    );
    if (!site) return;
    const destination = path.join(
      this.root,
      "test-results",
      "e2e",
      this.runID,
      site.product,
      site.slug,
    );
    fs.mkdirSync(destination, { recursive: true });
    for (const [index, attachment] of result.attachments.entries()) {
      if (!attachment.path || !fs.existsSync(attachment.path)) continue;
      const name = `${safeName(test.title)}-${index}-${path.basename(attachment.path)}`;
      fs.copyFileSync(attachment.path, path.join(destination, name));
    }
    if (result.error?.message) {
      fs.writeFileSync(
        path.join(destination, `${safeName(test.title)}-error.txt`),
        result.error.message,
      );
    }
  }

  onEnd(_result: FullResult) {
    const runRoot = path.join(this.root, "test-results", "e2e", this.runID);
    fs.mkdirSync(runRoot, { recursive: true });
    fs.writeFileSync(
      path.join(runRoot, "visual-contact-sheet.html"),
      renderContactSheet(this.root, runRoot, this.sites),
    );
  }
}

function renderContactSheet(
  root: string,
  runRoot: string,
  sites: SiteContract[],
) {
  const groups = sites
    .map((site) => {
      const screenshots = path.join(
        root,
        "test",
        "e2e",
        "screenshots",
        site.slug,
      );
      const images = fs.existsSync(screenshots)
        ? walkPNGs(screenshots)
            .map((file) => {
              const relative = path
                .relative(screenshots, file)
                .replaceAll("\\", "/");
              const source = path.relative(runRoot, file).replaceAll("\\", "/");
              return `<figure><img src="${escapeHTML(source)}" alt="${escapeHTML(relative)}"><figcaption>${escapeHTML(relative)}</figcaption></figure>`;
            })
            .join("")
        : "<p>尚无基线截图。</p>";
      return `<section><h2>${escapeHTML(site.product)} / ${escapeHTML(site.slug)}</h2><div class="grid">${images}</div></section>`;
    })
    .join("");
  return `<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>视觉验收联系表</title><style>body{font:14px/1.5 system-ui;margin:24px;background:#111;color:#eee}section{margin-block:32px}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(240px,1fr));gap:16px}figure{margin:0;padding:10px;background:#1d1d1d;border-radius:8px}img{display:block;width:100%;height:auto;background:#fff}figcaption{margin-top:8px;overflow-wrap:anywhere;color:#bbb}</style></head><body><h1>视觉验收联系表</h1>${groups}</body></html>`;
}

function walkPNGs(directory: string): string[] {
  return fs
    .readdirSync(directory, { withFileTypes: true })
    .flatMap((entry) =>
      entry.isDirectory()
        ? walkPNGs(path.join(directory, entry.name))
        : entry.name.endsWith(".png")
          ? [path.join(directory, entry.name)]
          : [],
    )
    .sort();
}

function safeName(value: string) {
  return value.replace(/[^\p{L}\p{N}._-]+/gu, "-").slice(0, 96);
}

function escapeHTML(value: string) {
  return value.replace(
    /[&<>"']/g,
    (character) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        character
      ] || character,
  );
}
