# @budhash/gomanize

Romanize Hindi (Devanagari) into readable, diacritic-free Latin script — the
[gomanize](https://github.com/budhash/gomanize) engine (Go) compiled to
WebAssembly, with a thin JS/TS wrapper. It is the **same engine as the CLI**, so
output is identical; nothing is reimplemented.

Live demo: <https://budhash.com/gomanize>

## Install

```bash
npm install @budhash/gomanize
```

## Use

```js
import { load } from "@budhash/gomanize";

const g = await load();                       // instantiate once
g.translit("नमस्ते दुनिया");                    // "namaste duniya"
g.translit("गाना", { longVowels: true });      // "gaanaa"
g.translit("मेरे सपनों की रानी", { rerank: true }); // "mere sapnon ki rani"
```

CommonJS works too: `const { load } = require("@budhash/gomanize");`

## Options

All optional, default `false`:

| Group | Option | Effect |
|-------|--------|--------|
| rule tweak | `longVowels` | write "aa" for every ā (गाना → gaanaa) |
| rule tweak | `simpleNasals` | करें → karen, not karein |
| rule tweak | `keepMedialSchwa` | disable CCV deletion (जनता → janata) |
| learned | `schwaModel` | decision-tree schwa classifier |
| learned | `lexicon` | prefer attested spellings |
| learned | `rerank` | re-rank candidates with a char n-gram model |

## Notes

- `load()` is async (instantiates the wasm) and cached; `translit` is synchronous after.
- The `.wasm` is ~1.4 MB gzipped — it embeds the schwa model, lexicon, and n-gram data.
- Runs on **Node ≥ 18** and modern browsers/bundlers (the wasm + `wasm_exec.js` are resolved relative to the module).
- Romanization, not strict transliteration — see the [project README](https://github.com/budhash/gomanize). MIT licensed.
