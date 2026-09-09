/**
 * @budhash/gomanize — the gomanize engine (Go, compiled to WebAssembly) with a
 * thin JS wrapper. Romanizes Devanagari (Hindi) into readable, diacritic-free
 * Latin script. This is the same engine as the CLI, so output is identical.
 */

/** Transliteration options. All default to `false` (the rule-engine baseline). */
export interface Options {
  /** Rule tweak: write "aa" for every ā (गाना → gaanaa). */
  longVowels?: boolean;
  /** Rule tweak: simplified nasal endings (करें → karen, not karein). */
  simpleNasals?: boolean;
  /** Rule tweak: disable medial (CCV) schwa deletion (जनता → janata). */
  keepMedialSchwa?: boolean;
  /** Learned model: decision-tree schwa classifier for inherent-schwa decisions. */
  schwaModel?: boolean;
  /** Learned model: prefer attested spellings from the embedded lexicon. */
  lexicon?: boolean;
  /** Learned model: re-rank candidates with a character n-gram model. */
  rerank?: boolean;
}

export interface Gomanize {
  /**
   * Romanize Devanagari text into Latin script. Whitespace and punctuation are
   * preserved; words are romanized in place.
   */
  translit(text: string, options?: Options): string;
}

/** Where to load the engine assets from. Defaults resolve next to the module. */
export interface LoadOptions {
  /** URL/path to gomanize.wasm. */
  wasmURL?: string | URL;
  /** URL/path to wasm_exec.js (ignored if a host page already defines `Go`). */
  execURL?: string | URL;
}

/**
 * Load and instantiate the WebAssembly engine. The result is cached, so calling
 * `load()` again returns the same instance. `translit` is synchronous once
 * loaded.
 *
 * ```js
 * import { load } from "@budhash/gomanize";
 * const g = await load();
 * g.translit("नमस्ते दुनिया");            // "namaste duniya"
 * g.translit("गाना", { longVowels: true }); // "gaanaa"
 * ```
 */
export function load(opts?: LoadOptions): Promise<Gomanize>;

/** Package version, e.g. "1.1.0". */
export const version: string;

declare const _default: { load: typeof load; version: string };
export default _default;
