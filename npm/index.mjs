// @budhash/gomanize — ESM loader for the gomanize WebAssembly engine.
// Works in Node (reads the .wasm from disk) and in browsers/bundlers (fetches
// it, resolved relative to this module via import.meta.url). No top-level Node
// imports, so bundling for the browser stays clean.

/** Package version (kept in sync with package.json on release). */
export const version = "1.2.1";

const FLAGS = ["longVowels", "simpleNasals", "keepMedialSchwa", "schwaModel", "lexicon", "rerank"];
let _instance;

const isNode = () =>
  typeof process !== "undefined" && !!(process.versions && process.versions.node);

// wasm_exec.js is a plain script (not a module) shipped by the Go toolchain; it
// defines globalThis.Go. Evaluate it once in the global scope. If a host page
// already loaded it (e.g. via a <script> tag), this is a no-op.
async function ensureGo(url) {
  if (typeof globalThis.Go !== "undefined") return;
  if (isNode()) {
    const { readFile } = await import("node:fs/promises");
    const vm = await import("node:vm");
    vm.runInThisContext(await readFile(url, "utf8"));
  } else {
    // eslint-disable-next-line no-eval
    (0, eval)(await (await fetch(url)).text());
  }
}

async function wasmBytes(url) {
  if (isNode()) {
    const { readFile } = await import("node:fs/promises");
    return readFile(url);
  }
  return (await fetch(url)).arrayBuffer();
}

/**
 * @param {{ wasmURL?: string|URL, execURL?: string|URL }} [opts]
 *   Override where the .wasm and wasm_exec.js are loaded from. Defaults resolve
 *   to ./dist/* next to this module (npm layout).
 */
export async function load(opts = {}) {
  if (_instance) return _instance;
  const execURL = opts.execURL ?? new URL("./dist/wasm_exec.js", import.meta.url);
  const wasmURL = opts.wasmURL ?? new URL("./dist/gomanize.wasm", import.meta.url);
  await ensureGo(execURL);
  const go = new globalThis.Go();
  const { instance } = await WebAssembly.instantiate(await wasmBytes(wasmURL), go.importObject);
  // main() sets globalThis.gomanizeTranslit and then blocks, keeping the
  // exported function callable for the lifetime of the instance.
  go.run(instance);
  const fn = globalThis.gomanizeTranslit;
  if (typeof fn !== "function") throw new Error("gomanize: wasm did not initialize");
  _instance = {
    translit(text, options) {
      // Read each option exactly once into a fresh plain object of validated
      // primitives; only that crosses syscall/js. Reading the caller's object
      // again (a getter or Proxy can change its answer) let a BigInt reach Go,
      // whose panic killed the instance and later the host process.
      const clean = {};
      if (options != null) {
        if (typeof options !== "object") throw new TypeError("gomanize: options must be an object");
        const language = options.language;
        if (language !== undefined) {
          if (typeof language !== "string") throw new TypeError("gomanize: language must be hindi or bengali");
          clean.language = language;
        }
        for (const flag of FLAGS) {
          const value = options[flag];
          if (value != null && typeof value !== "boolean") {
            throw new TypeError(`gomanize: option ${flag} must be a boolean`);
          }
          clean[flag] = value === true;
        }
      }
      const result = fn(text == null ? "" : String(text), clean);
      if (result instanceof Error) throw result;
      return result;
    },
  };
  return _instance;
}

export default { load, version };
