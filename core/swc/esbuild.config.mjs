import * as esbuild from "esbuild";
import { mkdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const outFile = join(__dirname, "../engine/assets/swc/swc.js");

mkdirSync(dirname(outFile), { recursive: true });

await esbuild.build({
  entryPoints: [join(__dirname, "src/main.ts")],
  bundle: true,
  format: "iife",
  globalName: "SumeruSWC",
  outfile: outFile,
  target: "es2022",
  sourcemap: true,
  minify: process.env.NODE_ENV === "production",
  define: {
    "process.env.NODE_ENV": JSON.stringify(process.env.NODE_ENV ?? "development"),
  },
});

console.log(`SWC bundle → ${outFile}`);
