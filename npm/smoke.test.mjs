// Smoke test for the @budhash/gomanize package: load the wasm engine and check
// output matches the CLI across a few option combos (ESM + CJS entry points).
import { load } from "./index.mjs";
import { createRequire } from "node:module";

const cases = [
  ["नमस्ते दुनिया", {}, "namaste duniya"],
  ["मेरे सपनों की रानी", { rerank: true }, "mere sapanon ki rani"],
  ["गाना", { longVowels: true }, "gaanaa"],
  ["नमस्ते दुनिया", { lexicon: true }, "namaste duniya"],
];

let fail = 0;
const g = await load();
for (const [text, opts, want] of cases) {
  const got = g.translit(text, opts);
  const ok = got === want;
  if (!ok) fail++;
  console.log(`${ok ? "PASS" : "FAIL"}  ${JSON.stringify(opts)}  -> ${JSON.stringify(got)}${ok ? "" : "  want " + JSON.stringify(want)}`);
}

// exercise the CJS entry point too
const require = createRequire(import.meta.url);
const cjs = require("./index.cjs");
const g2 = await cjs.load();
const cjsOut = g2.translit("भारत");
const cjsOk = cjsOut === "bharat";
if (!cjsOk) fail++;
console.log(`${cjsOk ? "PASS" : "FAIL"}  (cjs) भारत -> ${JSON.stringify(cjsOut)}`);

console.log(fail ? `\n${fail} FAILED` : "\nall passed");
process.exit(fail ? 1 : 0);
