// @budhash/gomanize — ESM loader for the gomanize WebAssembly engine.
// Works in Node (reads the .wasm from disk) and in browsers/bundlers (fetches
// it, resolved relative to this module via import.meta.url). No top-level Node
// imports, so bundling for the browser stays clean.

let _instance;

const isNode = () =>
  typeof process !== "undefined" && !!(process.versions && process.versions.node);

// wasm_exec.js is a plain script (not a module) shipped by the Go toolchain; it
// defines globalThis.Go. Evaluate it once in the global scope.
async function ensureGo() {
  if (typeof globalThis.Go !== "undefined") return;
  const url = new URL("./dist/wasm_exec.js", import.meta.url);
  if (isNode()) {
    const { readFile } = await import("node:fs/promises");
    const vm = await import("node:vm");
    vm.runInThisContext(await readFile(url, "utf8"));
  } else {
    // eslint-disable-next-line no-eval
    (0, eval)(await (await fetch(url)).text());
  }
}

async function wasmBytes() {
  const url = new URL("./dist/gomanize.wasm", import.meta.url);
  if (isNode()) {
    const { readFile } = await import("node:fs/promises");
    return readFile(url);
  }
  return (await fetch(url)).arrayBuffer();
}

export async function load() {
  if (_instance) return _instance;
  await ensureGo();
  const go = new globalThis.Go();
  const { instance } = await WebAssembly.instantiate(await wasmBytes(), go.importObject);
  // main() sets globalThis.gomanizeTranslit and then blocks, keeping the
  // exported function callable for the lifetime of the instance.
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

export default { load };
