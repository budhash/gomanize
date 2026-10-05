# @budhash/gomanize

Romanize Hindi (Devanagari) or experimental Bengali into readable, diacritic-free Latin script — the
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
g.translit("আমি বাংলা", { language: "bengali" }); // "ami bangla"
g.translit("गाना", { longVowels: true });      // "gaanaa"
g.translit("मेरे सपनों की रानी", { rerank: true }); // "mere sapanon ki rani"
```

CommonJS works too: `const { load } = require("@budhash/gomanize");`

## Options

Language defaults to `"hindi"` on every call; boolean flags default to `false`.
Unsupported languages (including empty strings) throw `Error`; language names
are case-insensitive at runtime. The TypeScript type uses canonical lowercase names.

| Group | Option | Effect |
|-------|--------|--------|
| language | `language` | `"hindi"` (default) or `"bengali"` |
| Hindi rule tweak | `longVowels` | write "aa" for every ā (गाना → gaanaa) |
| Hindi rule tweak | `simpleNasals` | करें → karen, not karein |
| Hindi rule tweak | `keepMedialSchwa` | disable CCV deletion (जनता → janata) |
| learned | `schwaModel` | decision-tree schwa classifier |
| learned | `lexicon` | prefer attested spellings |
| learned | `rerank` | rank candidates; Bengali requires `schwaModel: true` |

For Bengali, `schwaModel` enables the three-way vowel model; `rerank` requires
that option too. `lexicon` remains opt-in and can change spelling style. The
three Hindi rule tweaks do not implement Bengali rendering styles; supplying
any of them bypasses Bengali lexicon lookup and reranking. Language selection
is explicit, not script detection. Input in other scripts is preserved.

## Notes

- `load()` is async (instantiates the wasm) and cached; `translit` is synchronous after.
- The `.wasm` embeds both languages and their learned data; Bengali needs no additional download.
- Runs on **Node ≥ 18** and modern browsers/bundlers (the wasm + `wasm_exec.js` are resolved relative to the module).
- Romanization, not strict transliteration — see the [project README](https://github.com/budhash/gomanize). MIT code; embedded data has separate licenses in [NOTICE.md](NOTICE.md).
