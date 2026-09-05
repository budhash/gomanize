# gomanize web demo

A static, single-page browser demo of gomanize. The Go library is compiled to
WebAssembly, so **all transliteration runs in the visitor's browser** — no
server, no backend, and no text ever leaves the machine. The embedded models
(schwa tree, lexicon, n-gram LM) are baked into the `.wasm` via `go:embed`.

## Files

| File | Source | Committed? |
|------|--------|------------|
| `index.html` | hand-written page | yes |
| `README.md` | this file | yes |
| `gomanize.wasm` | `make wasm` (from `cmd/gomanize-wasm`) | no — gitignored, built |
| `wasm_exec.js` | copied from the Go toolchain by `make wasm` | no — gitignored, built |

`wasm_exec.js` must match the Go version that built the `.wasm`, which is why it
is copied fresh on every build rather than committed.

## Build & run locally

```bash
make wasm         # build web/gomanize.wasm + copy web/wasm_exec.js
make wasm-serve   # build, then serve web/ at http://localhost:8080
```

Opening `index.html` via `file://` will not work — browsers block `fetch()` of
the `.wasm` from the filesystem. Use `make wasm-serve` (or any static server).

## Deployment

`.github/workflows/pages.yml` rebuilds the `.wasm` and publishes this directory
to GitHub Pages on every push to `main`. The site is served from the repo root
of the Pages artifact, so `index.html` fetches `gomanize.wasm` and
`wasm_exec.js` by relative path.

**One-time setup:** in the repo's *Settings → Pages*, set **Source** to
**GitHub Actions**.

## Size

The `.wasm` is ~5.2 MB raw, ~1.4 MB gzipped (GitHub Pages gzips automatically),
downloaded once and then cached. If that ever needs trimming, the levers are
lazy-loading the embedded models as separate fetched assets, or building with
TinyGo — see `docs/ROADMAP.md` (F-0006).
