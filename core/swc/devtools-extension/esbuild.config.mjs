import * as esbuild from "esbuild";
import { mkdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const outdir = join(__dirname, "dist");

mkdirSync(outdir, { recursive: true });

const shared = {
  bundle: true,
  format: "iife",
  target: "es2020",
  platform: "browser",
  sourcemap: true,
  minify: process.env.NODE_ENV === "production",
  logLevel: "info",
};

await esbuild.build({
  ...shared,
  entryPoints: {
    panel: join(__dirname, "src/panel.ts"),
    content: join(__dirname, "src/content.ts"),
    devtools: join(__dirname, "src/devtools.ts"),
  },
  outdir,
});

console.log(`SWC Vision extension → ${outdir}`);
