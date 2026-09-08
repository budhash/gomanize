// @budhash/gomanize — CommonJS loader (Node). Same engine as index.mjs.
"use strict";

let _instance;

async function load() {
  if (_instance) return _instance;
  const fs = require("node:fs");
  const path = require("node:path");
  if (typeof globalThis.Go === "undefined") {
    const vm = require("node:vm");
    vm.runInThisContext(fs.readFileSync(path.join(__dirname, "dist", "wasm_exec.js"), "utf8"));
  }
  const bytes = fs.readFileSync(path.join(__dirname, "dist", "gomanize.wasm"));
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

module.exports = { load };
module.exports.default = module.exports;
