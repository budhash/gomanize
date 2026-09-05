# Roadmap

Forward-looking directions for gomanize after v1.0.0. This is a menu of options
with honest tradeoffs, not a committed plan. The live task backlog is `TASKS.md`
(`./tools/tasks tree`); design constraints for each item are in
[`DESIGN.md`](DESIGN.md) §6; the evidence behind the framing is in
[`RESEARCH.md`](RESEARCH.md).

## Where v1.0.0 leaves the project

The accuracy work is largely closed. Rules are at their measured ceiling, the
learned components (schwa classifier, lexicon, re-ranker) captured most of what
remained, and further benchmark points are a data problem with diminishing
returns. What is distinctive is elsewhere:

- A zero-dependency, embeddable, **native-to-Roman** engine in Go — a direction
  the field mostly ignores (most work goes Roman-to-native with neural models).
- An evaluation methodology (multi-reference scoring, contamination discipline,
  recorded negative results) more rigorous than much of the published work.
- The only lyrics-oriented Devanagari-Roman gold set that exists, small as it is.

The three directions below build on those, roughly in order of compounding value.

## Direction 1 — WASM build + web demo (highest leverage)

The stated purpose has always been song lyrics. The gap is not accuracy; it is
reach. Go compiles to WebAssembly cleanly, and the whole engine (including the
~518 KB of embedded model data) fits in a browser with no server.

The reason this matters beyond convenience: **it closes the data loop**. The
lexicon's coverage stops at 78.2% of the Dakshina-train vocabulary, and the one
thing that raises it is human-attested spellings — data that cannot be mined
(measured at 43% precision) and is expensive to license. A public tool where
users paste lyrics and can correct a romanization turns those users into the
annotators the research says are required. Tool brings users; users bring
attestations; attestations improve the tool.

Scope:
- `GOOS=js GOARCH=wasm` build target in the Makefile; verify embedded assets
  load under WASM (they should — `go:embed` is compile-time).
- A single static page: paste Devanagari, get romanization, toggle the flags,
  edit an output inline.
- Optional and deferred: a lightweight, consent-based mechanism to collect
  user corrections as candidate lexicon entries (with the same human-review gate
  the existing miner uses — corrections are proposals, not auto-promotions).
- Distribution: the page is static, so GitHub Pages or any CDN serves it.

Risks: WASM binary size (mitigate by confirming the model files dominate and are
acceptable, or gating `--rerank`/`--lexicon` behind lazy loading); browser
input-method quirks (the Cf-stripping and NFC work already done helps here).

## Direction 2 — Aksharantar convention parity

The Aksharantar AK-Freq slice scores 42.6% where the Dakshina-tuned rules score
~69% on Dakshina's own conventions. Most of that gap is **annotation-convention
shift**, not a quality defect: Aksharantar's annotators systematically double
vowels (*atyaachaarapoorn*) where Dakshina's curated set does not
(*atyacharpurn*). This is worth pursuing, but with eyes open about what "parity"
means.

The honest version of this goal is not one output that satisfies both
conventions — that is impossible, since they disagree — but **a selectable
convention**. The engine already has the machinery: this is what schemes are
for. An `aksharantar` (or `long-vowel-heavy`) scheme could:
- Apply the vowel-doubling conventions Aksharantar prefers (the ee/oo and aa
  rules that measured net-negative against Dakshina are likely net-positive
  against Aksharantar — this is directly testable with the suite already built).
- Be measured on the Aksharantar test set the same way the colloquial scheme is
  measured on Dakshina.

Prerequisite: the scheme layer currently selects rules from the language catalog
but base romanizations live in the symbol table, so a convention that changes
vowel-length defaults cleanly needs the per-scheme symbol-map work noted in
DESIGN §6. Do that first, then the two conventions coexist as two schemes rather
than one compromised default.

What to avoid: chasing a single blended output that lowers Dakshina scores to
raise Aksharantar ones. The value is giving users the convention they want, not
averaging two incompatible ones.

## Direction 3 — Additional languages (Marathi, Nepali)

`brahmic.SchwaRules()` already extracts the script-general schwa rules, so a
second Brahmic language is implementable without copying the rule set — it needs
a symbol map, a script config, and language-specific rules. One added language
would prove the shared-Brahmic abstraction is real rather than aspirational, and
"the Go romanizer for Indic scripts" is an identity nothing else in the ecosystem
holds.

Caveats (DESIGN §6): the renderer's inherent vowel is hardcoded `"a"` (correct
for Hindi, wrong for scripts with a different inherent vowel), and a few Hindi
rules carry Devanagari literals that need auditing per language. Less compounding
than Direction 1, but a natural way to broaden the library's reach.

## Supporting work (enables the above)

- **Grow the lyrics gold set toward ~500 lines.** At that size it is a citable
  dataset contribution and a much stronger signal for the primary use case.
  More public-domain verse can be committed directly; copyrighted lyrics stay
  out-of-repo behind a fetch script (RESEARCH §3).
- **Per-scheme symbol maps** (DESIGN §6) — the shared prerequisite for both an
  IAST scheme and the Aksharantar-convention scheme (Direction 2).
- **Close the P3 correctness items** already filed (NFC normalization T-0025,
  race-free debug traces T-0026, remaining source-rune conversions T-0024)
  before any large new surface lands on them.

## What to avoid

- **Embedding a neural model.** It would erase the zero-dependency, embeddable
  identity that makes this engine distinct, to chase a ceiling that is mostly
  convention variance rather than error.
- **Roman-to-Devanagari.** A different problem wearing the same name (lossy,
  sequence-disambiguation); it deserves its own project, not a bolt-on here.

## Not doing anything is also fine

v1.0.0 is a complete, documented, honest artifact. Small tools are allowed to be
done. If this sits at 1.0, the decision records and tracker mean a future
contributor — or a future you — can pick any direction above from a cold start.
