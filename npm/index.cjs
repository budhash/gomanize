// @budhash/gomanize — CommonJS loader (Node). Same engine as index.mjs.
"use strict";

/** Package version (kept in sync with package.json on release). */
const version = "1.2.1";

const FLAGS = ["longVowels", "simpleNasals", "keepMedialSchwa", "schwaModel", "lexicon", "rerank"];
let _instance;

async function load(opts = {}) {
  if (_instance) return _instance;
  const fs = require("node:fs");
  const path = require("node:path");
  const execPath = opts.execURL || path.join(__dirname, "dist", "wasm_exec.js");
  const wasmPath = opts.wasmURL || path.join(__dirname, "dist", "gomanize.wasm");
  if (typeof globalThis.Go === "undefined") {
    const vm = require("node:vm");
    vm.runInThisContext(fs.readFileSync(execPath, "utf8"));
  }
  const bytes = fs.readFileSync(wasmPath);
  const go = new globalThis.Go();
  const { instance } = await WebAssembly.instantiate(bytes, go.importObject);
  go.run(instance);
  const fn = globalThis.gomanizeTranslit;
  if (typeof fn !== "function") throw new Error("gomanize: wasm did not initialize");
  _instance = {
    translit(text, options) {
      // Reject unsupported JS types before crossing syscall/js: a BigInt flag
      // panics inside the Go runtime and kills the instance for every later call.
      if (options != null) {
        if (typeof options !== "object") throw new TypeError("gomanize: options must be an object");
        if (options.language !== undefined && typeof options.language !== "string") {
          throw new TypeError("gomanize: language must be hindi or bengali");
        }
        for (const flag of FLAGS) {
          const value = options[flag];
          if (value != null && typeof value !== "boolean") {
            throw new TypeError(`gomanize: option ${flag} must be a boolean`);
          }
        }
      }
      const result = fn(text == null ? "" : String(text), options || {});
      if (result instanceof Error) throw result;
      return result;
    },
  };
  return _instance;
}

module.exports = { load, version };
module.exports.default = module.exports;
