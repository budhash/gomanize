# Learnings

Durable insights, gotchas, and decisions — the "why" that isn't obvious from the
code or git history. Newest first.

## Parser fix — independent vowels vs matras (2026-09, F-0010)

- **Bug:** `गई` (ग + independent ई) romanized to `gi`, same as `गी` (ग + matra
  ी). Root cause: `script/brahmic/categories.go` maps *both* independent vowels
  (`core.CatVowel`) and matras (`CatMatra`) to `core.UnitVowel`, and
  `renderer.go` suppressed the consonant's inherent schwa before *any*
  `UnitVowel`. So an independent vowel behaved like a matra and dropped the
  consonant's `a`.
- **Fix:** store `IsMatra` on `BrahmicData` (set from the source category) and
  suppress the schwa only before a matra — an independent vowel starts its own
  syllable, so the consonant keeps its `a` (`गई→gai`).
- **Gotcha the regression harness caught:** independent **अ** *is* the bare-`a`
  vowel, so keeping the schwa *and* अ doubled it (`दरअसल→daraasal`). अ must
  coalesce with the schwa. The refined rule: suppress before a matra **or** the
  independent अ (`BaseRom=="a"`); keep it before ई/उ/ए/…. Two exact regressions
  in the transition-matrix diff surfaced this immediately — aggregate accuracy
  (which *rose* slightly) would not have.
- **Process that worked:** baseline snapshot → fix → `make regression-diff`
  (net +44 exact, 0 regressions, after the अ refinement) → un-skip the staged
  invariant/golden anchors → pure gate + full suite + multi-mode spot-check. The
  gold-free differential (`गी ≠ गई`) is the cheapest guard; keep writing those.

## v1.1.0 — WASM demo, npm distribution, tokenless release (2026-09)

### Distribution: ship the engine, don't reimplement it
gomanize is one engine, written in Go. The npm package `@budhash/gomanize` ships
that engine **compiled to WebAssembly** plus a thin JS/TS wrapper — it is *not* a
JavaScript reimplementation. Consequence: there is only ever one implementation,
so output is byte-identical to the CLI and cannot drift. This is the opposite
call from the sibling tools (confix, zap-sh), whose logic is small enough to port
to JS — those keep a bash reference **and** a JS port, kept honest by a shared
conformance suite (bash is the oracle). Rule of thumb: reimplement only when the
logic is small and a conformance suite is cheaper than shipping a runtime; for a
real engine, ship the runtime (WASM).

### Go → WASM gotchas
- The codebase compiled to `GOOS=js GOARCH=wasm` on the first try because OS
  dependencies (`os.Stdin`, etc.) live only in `cmd/`, not the library, and
  `go:embed` is compile-time so all models bundle in automatically. Keep it that
  way: no OS/syscall use in the importable packages.
- A package tagged `//go:build js && wasm` (e.g. `cmd/gomanize-wasm`) is silently
  skipped by `go build/vet/test ./...` when other packages exist — no host stub
  needed, and `make ci` stays green.
- `wasm_exec.js` MUST match the Go toolchain that built the `.wasm`. Copy it fresh
  in `make wasm` (from `$(go env GOROOT)/lib/wasm/`, older layout `misc/wasm/`);
  never commit it.
- The `syscall/js` `main` sets the exported function then blocks on `select{}`;
  the function is registered synchronously during `go.run`, so callers can invoke
  it immediately after. Under Node this does not hang the process on exit.

### npm packaging + OIDC trusted publishing
- Options cross the JS boundary as a plain object of booleans mapped through one
  place (`webdemo.Options`), so JS, the demo, and the CLI share a single source
  of truth — add a flag in Go once.
- Trusted publishing (no `NPM_TOKEN`) needs **npm ≥ 11.5.1**. Node 20 ships npm
  10 and `npm@latest` (12.x) needs Node ≥ 22, so the release workflow pins
  `npm@11` on Node 20. `id-token: write` is required; provenance is automatic.
  `registry-url` in `setup-node` coexists fine with OIDC.
- Prefer **verifying** the tag matches `npm/package.json` (fail on mismatch) over
  mutating the version in CI — so the published version is always what's
  committed. Bump `npm/package.json` and the `version` constants before tagging.
- gomanize's npm rides the project's own `v*` tag (npm == the Go engine), so one
  `git tag vX.Y.Z` publishes both the binaries (`release.yml`) and the npm
  package (`release-npm.yml`). confix/zap-sh instead use a separate `js-v*` tag
  because their JS is an independent port.

### GitHub Pages user-site cascade + custom domain
- The user site repo (`budhash.github.io`) owns `budhash.com`; every project repo
  with Pages enabled is then served automatically at `budhash.com/<repo>/`. Do
  **not** set a custom domain (or a `CNAME` file) on a project repo — that breaks
  it out of the path model.
- Pages from a **private** repo needs a paid plan; the served site is public
  regardless (truly-private Pages is Enterprise-only).
- Setting a custom domain makes `*.github.io` 301-redirect to it — so a brand-new
  domain looks "down" until DNS + the Let's Encrypt cert resolve. If the cert
  stalls > ~1h after a domain/visibility change, the unstick is Settings → Pages
  → Remove custom domain → Save → re-enter → Save.
- Cloudflare: keep the record **DNS-only (grey cloud)** so GitHub can issue the
  cert; if proxied later, SSL/TLS mode must be **Full** (never Flexible).

### Demo consumes the package's loader
The browser demo imports the same loader shipped in the npm package
(`web/gomanize.mjs`, vendored from `npm/index.mjs` by `make wasm`), pointed at the
local `.wasm`. One loader, and the demo stays self-contained/offline — no CDN and
no dependency on the package being published. A CDN import of the published
package was considered and rejected (worse reliability, no real gain).

### Verifying UI without the browser extension
When the Claude-in-Chrome extension is unavailable, headless Chrome
(`--headless=new --screenshot`) renders pages for visual checks, and running the
built `.wasm` under Node with `wasm_exec.js` validates the engine end-to-end —
both used throughout this work.
