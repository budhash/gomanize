// Smoke test for the @budhash/gomanize package: load the wasm engine and check
// output matches the CLI across a few option combos (ESM + CJS entry points).
import { load } from "./index.mjs";
import { createRequire } from "node:module";
import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";

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

for (const [name, engine] of [["ESM", g], ["CJS", g2]]) {
  assert.equal(engine.translit("আমি বাংলা", { language: "bengali" }), "ami bangla", name);
  assert.equal(engine.translit("আমি বাংলা", { language: "BENGALI" }), "ami bangla", name);
  assert.equal(engine.translit("নমস্কার", { language: "bengali" }), "nomoskar", name);
  assert.equal(engine.translit("गाना", { language: "hindi", longVowels: true }), "gaanaa", name);
  assert.equal(engine.translit("गाना"), "gana", name);
  assert.equal(engine.translit("গান", { language: "bengali" }), "gan", name);
  assert.equal(engine.translit("नमस्ते दुनिया"), "namaste duniya", name);
  assert.equal(engine.translit("আমি বাংলা"), "আমি বাংলা", name);
  assert.equal(engine.translit(null, { language: "bengali" }), "", name);
  for (const language of ["", "bn", "unknown", null, false, 7, {}, 1n, Symbol("bn")]) {
    assert.throws(() => engine.translit("আমি বাংলা", { language }), /language/, name);
    assert.equal(engine.translit("भारत"), "bharat", `${name}: runtime survives invalid language`);
  }
  // Non-boolean flags (a BigInt used to panic the Go runtime) and non-object
  // options are rejected in JS; the instance must survive every one.
  for (const value of [1n, Symbol("x"), "false", 1, {}]) {
    assert.throws(() => engine.translit("गाना", { longVowels: value }), /must be a boolean/, name);
    assert.equal(engine.translit("गाना"), "gana", `${name}: runtime survives bad flag`);
  }
  assert.throws(() => engine.translit("গান", "bengali"), /options must be an object/, name);
  // A getter or Proxy that answers differently on a second read must not reach
  // Go with an unvalidated value (it used to panic and later kill the host).
  let reads = 0;
  const flaky = { get longVowels() { return reads++ === 0 ? true : 7n; } };
  assert.equal(engine.translit("गाना", flaky), "gaanaa", `${name}: getter read once`);
  let proxyReads = 0;
  const proxy = new Proxy({}, { get: (_, key) => (key === "longVowels" ? (proxyReads++ === 0 ? true : 1n) : undefined) });
  assert.equal(engine.translit("गाना", proxy), "gaanaa", `${name}: Proxy read once`);
  assert.equal(engine.translit("गाना"), "gana", `${name}: runtime survives getter/Proxy options`);
  assert.equal(engine.translit("गाना", { longVowels: false, schwaModel: null }), "gana", name);
  // make npm-test supplies the freshly built CLI for cross-interface parity.
  if (process.env.GOMANIZE_CLI) {
    const profiles = [{}, { schwaModel: true }, { lexicon: true },
      { schwaModel: true, lexicon: true }, { schwaModel: true, rerank: true },
      { schwaModel: true, rerank: true, lexicon: true },
      { longVowels: true, schwaModel: true, rerank: true, lexicon: true },
      { simpleNasals: true }, { keepMedialSchwa: true }];
    for (const language of ["hindi", "bengali"]) {
      const text = "সমতা, অডিও। আমি বাংলা\nनमस्ते दुनिया, जनता। English 123";
      for (const profile of profiles) {
        const flags = Object.keys(profile).map(k => "--" + k.replace(/[A-Z]/g, c => "-"+c.toLowerCase()));
        const cli = execFileSync(process.env.GOMANIZE_CLI, ["--language="+language, ...flags, text], { encoding: "utf8" });
        assert.equal(engine.translit(text, { language, ...profile })+"\n", cli, `${name} ${language} ${JSON.stringify(profile)}`);
      }
    }
  }
  console.log(`PASS ${name}: Bengali, language/option reset, invalid language recovery and CLI parity`);
}

console.log(fail ? `\n${fail} FAILED` : "\nall passed");
process.exit(fail ? 1 : 0);
