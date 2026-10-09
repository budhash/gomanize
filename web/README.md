# gomanize web demo

A static, single-page browser demo of gomanize. The Go library is compiled to
WebAssembly, so **all transliteration runs in the visitor's browser** — no
server, no backend, and no text ever leaves the machine. The embedded models
(Hindi schwa tree, lexicon and n-gram LM; Bengali vowel model, lexicon and
selector) are baked into the `.wasm` via `go:embed`.

## Files

| File | Source | Committed? |
|------|--------|------------|
| `index.html` | hand-written page | yes |
| `README.md` | this file | yes |
| `gomanize.wasm` | `make wasm` (from `cmd/gomanize-wasm`) | no — gitignored, built |
| `wasm_exec.js` | copied from the Go toolchain by `make wasm` | no — gitignored, built |
| `NOTICE.md` | copied from `npm/NOTICE.md` by `make wasm` | no — gitignored, built |
| `gomanize.mjs` | vendored from `npm/index.mjs` by `make wasm` | no — gitignored, built |

`wasm_exec.js` must match the Go version that built the `.wasm`, which is why it
is copied fresh on every build rather than committed. `gomanize.mjs` is the same
loader shipped in the `@budhash/gomanize` npm package — the page consumes it
(pointed at the local `.wasm`) so the demo and the package share one loader.

## Build & run locally

```bash
make wasm         # build web/gomanize.wasm + wasm_exec.js + vendored gomanize.mjs
make wasm-serve   # build, then serve web/ at http://localhost:8080
```

Opening `index.html` via `file://` will not work — browsers block `fetch()` of
the `.wasm` from the filesystem. Use `make wasm-serve` (or any static server).

## Language selection

The page starts in Hindi with its existing Long vowels checkbox selected.
Bengali starts with all flags off; toggles are remembered separately per language
for the current page session. Switching languages preserves entered text,
updates the input label/placeholder, and immediately rerenders it. Sample text
uses the selected language. The Hindi-only style group is hidden for Bengali;
Bengali re-ranking needs the vowel model enabled. Bengali remains experimental.
The same ESM loader forwards `language` to WASM; no network request carries input.

`make npm-test` builds the CLI and WASM and checks ESM/CJS parity for both
languages across baseline, learned and style profiles, missing-language defaults,
invalid-language recovery and option reset. Browser interaction is checked
locally before a PR; every merge to `main` redeploys the demo (see Deployment).

## Deployment

`.github/workflows/pages.yml` rebuilds the `.wasm` and publishes this directory
to GitHub Pages on every push to `main`. The site is served from the repo root
of the Pages artifact, so `index.html` fetches `gomanize.wasm` and
`wasm_exec.js` by relative path.

**One-time setup:** in the repo's *Settings → Pages*, set **Source** to
**GitHub Actions**.

## Size

The `.wasm` embeds both language implementations and their learned data,
downloaded once and then cached. Measure the built artifact for current size. If that ever needs trimming, the levers are
lazy-loading the embedded models as separate fetched assets, or building with
TinyGo; tracked as keel follow-up K24 in TASKS.md.
