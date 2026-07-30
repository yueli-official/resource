import assert from "node:assert/strict";
import test from "node:test";

import {
  assetExtension,
  assetFileIcon,
  assetFileTone,
  formatAssetSize,
} from "./asset-display.mjs";

test("extracts normalized file extensions", () => {
  assert.equal(assetExtension("Release.V1.ZIP"), "zip");
  assert.equal(assetExtension("README"), "");
  assert.equal(assetExtension("archive.tar.gz"), "gz");
});

test("chooses semantic icons and tones from MIME type or extension", () => {
  assert.equal(assetFileIcon("preview.bin", "image/webp"), "i-tabler-photo");
  assert.equal(assetFileIcon("bundle.7z"), "i-tabler-file-zip");
  assert.equal(assetFileIcon("manual.pdf"), "i-tabler-file-type-pdf");
  assert.equal(assetFileTone("scene.blend"), "text-primary bg-primary/10");
  assert.equal(assetFileTone("unknown.bin"), "text-muted bg-[var(--resource-surface-card)]");
});

test("formats byte sizes without negative output", () => {
  assert.equal(formatAssetSize(-1), "0 B");
  assert.equal(formatAssetSize(1023), "1023 B");
  assert.equal(formatAssetSize(1024), "1 KB");
  assert.equal(formatAssetSize(1536), "1.5 KB");
  assert.equal(formatAssetSize(10 * 1024 * 1024), "10 MB");
});
