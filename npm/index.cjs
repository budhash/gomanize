// @budhash/gomanize — CommonJS loader (Node). Same engine as index.mjs.
"use strict";

/** Package version (kept in sync with package.json on release). */
const version = "1.2.0";

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
      return fn(text == null ? "" : String(text), options || {});
    },
  };
  return _instance;
}

module.exports = { load, version };
module.exports.default = module.exports;
