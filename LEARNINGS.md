# Learnings

Durable insights, gotchas, and decisions — the "why" that isn't obvious from the
code or git history. Newest first.

## Structured parser-QA complete — coverage nets vs bug-finding (2026-09, F-0010)

The 5-tier plan (Tiers 1–5) is fully shipped. What each tier actually bought:
- **The bugs were found by Tiers 1–2 + real usage** (per-construct analyzer,
  gold-free invariants) and the corpus-diff harness — गई (T-0040), chandrabindu
  (T-0045). **Tiers 3–4 found none** — and that's the point of a coverage net:
  it's built *after* the known bugs are fixed, to prove the space is exercised
  and to fail loudly on *future* regressions, not to find today's bugs.
- **Tier 3 (combinatorial, T-0041):** ~1,900 constructs off the symbol map,
  asserting parse well-formedness. Two of my initial assertions were over-strict
  and had to be relaxed to documented behaviour (nukta combines only where a
  precomposed mapping exists; halant is consumed so conjunct spans have a gap).
  Writing a coverage net teaches you your own invariants.
- **Tier 4 (akshara differential, T-0042):** the key realization is that
  gomanize splits *finer* than aksharas (matra + each conjunct member are their
  own units), so the correct relation is **refinement** (akshara boundaries ⊆
  gomanize boundaries), not equality. A missing boundary = a merge bug; extra
  boundaries are by design. Also honest: T-0040 was a *render* bug, not a
  segmentation merge — the parser always kept गई as two units — so Tier 4 guards
  segmentation, it would not have caught T-0040. A negative control proves the
  reference isn't vacuously passing.
- **Dependency-free constraint shapes test design too:** no UAX #29 library, so
  the akshara reference is ~15 lines of codepoint rules in-repo. Cheaper than a
  dependency and it encodes exactly the boundary semantics we care about.

## Held-out construct regression set (2026-09, T-0044)

- **Purpose, and why it's not another curated_hi.** The frequency/curation-sampled
  benchmarks barely cover rare constructs (गई-family ≈0.2% of curated), so
  systematic parser bugs in them are invisible to aggregate accuracy. This set is
  stratified *by construct* and held out from all training/benchmarks — a targeted
  regression net, the coverage complement to curated_hi.
- **Build pipeline that worked:** `tools/mine_constructs.py` frequency-ranks
  construct-bearing tokens from external corpora, dedupes against every
  `benchmark/data` native (contamination guard), and fills engine output as a
  *non-anchoring* reference column → human triages `AUTO_OK`/`CHANGED`/`drop`.
  Scored **match-any** by reusing `loadReferenceSets` + `matchesAny` (one CSV row
  per accepted variant; नहिं→`nahin`/`nahi`), with per-construct floors below
  current rates (overall 93%, chandrabindu 96%) so a construct-level regression
  trips them but intended small changes pass.
- **Source register matters more than volume.** Two corpora, ~167 mined tokens →
  **129 kept**. The frequency-list batch (modern prose) was clean; the verse
  batch (RAW.tsv) was **Awadhi/Braj/Chhattisgarhi + hyper-technical Jain verse** —
  off-domain for a colloquial/lyrics romanizer, so most of it was dropped. Mining
  volume is cheap; on-register, license-clean source text is the scarce input.
- **A gold set records known misses, it doesn't hide them.** 9/129 default misses
  survived validation (vowel-length aa: फाँसी→faansi, तांगा→taanga; schwa: अंततः→
  antatah, दरअस्ल→darasal) — real future work, logged by the test rather than
  papered over by setting gold = current engine output.

## Parser fix — chandrabindu nasal, and lexical vs structural rules (2026-09, T-0045)

- **Bug:** `render.chandrabindu.final-silent` suppressed the nasal 'n' for *any*
  `ा`+`ँ` sequence, so it dropped mid-word (चाँद→"chaad", पाँच→"paach") and in
  longer word-final forms (कहाँ→"kahaa"). Only माँ→"maa" was right — by luck.
- **The trap — a structural guard that can't work.** The obvious fix was to
  mirror the anusvara sibling (`word-final` + `consonantCount<=1`). It passed the
  30k-Dakshina regression harness (+2/0)… but Codex review found the
  counterexample: **हाँ→"haan"** and माँ→"maa" have the *identical* shape
  (C+ा+ँ, word-final, one consonant) yet romanize differently. No structural
  guard can separate them — **the distinction is lexical, not structural.** The
  fix is a whole-word **allowlist** (`{माँ}`); everything else keeps the nasal.
  Lesson: when two inputs share every structural feature but need different
  output, stop looking for a cleverer predicate — it's a lexical exception.
- **Why the harness missed हाँ (argues for T-0044).** The regression harness
  runs over Dakshina natives only; हाँ isn't in that 30k, so `consonantCount<=1`
  looked clean there. Codex caught it by grepping the *other* bundled corpora
  (COMI-LINGUA, Aksharantar) where हाँ→haan is attested. A single-corpus
  regression gate has blind spots exactly where that corpus is thin — this is
  the case for the multi-source held-out construct set (T-0044).
- **Data settles scheme calls, but the maintainer owns iconic ones.** माँ→"maan"
  is the *plurality* in the aggregate data (count 13 vs maa's 8), which argued
  for dropping the exception entirely. But माँ→"maa" is iconic and Dakshina's
  preference; the maintainer's call was maa. Frequency informs; it doesn't
  override a deliberate convention (cf. DESIGN §3 divergences).
- **Process:** Codex design/code review on the *diff + regression evidence*
  (not a pre-design) was the high-signal use — it found the हाँ counterexample
  and confirmed the labial-`m` inconsistency (साँप→saanp vs संबंध→sambandh) as a
  separate task (T-0046), not scope creep here.

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
