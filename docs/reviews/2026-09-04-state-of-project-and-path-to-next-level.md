# Gomanize — State of the Project & Path to the Next Level

**Date:** 2026-09-04
**Author:** Deep-dive review (multi-agent: architecture, benchmarks, dataset research, methods research)
**Status:** Findings + roadmap. Tracked breakdown lives in [`/TASKS.md`](../../TASKS.md) (features F-0001…F-0004).

---

## Executive assessment

Gomanize is a well-architected, rule-based Devanagari→Roman transliterator that has reached the natural accuracy ceiling of hand-written rules. The engine design is clean and above-average for its class; the remaining gap to "the next level" is **not** more rules. It is two things: (1) the project is *mismeasuring* itself against a single-reference benchmark when romanization is fundamentally many-to-one, and (2) the last ~10% of failures are lexical/morphological — loanwords, compounds, and ambiguous medial schwa — which the published research shows rules structurally cannot solve, but a small learned component can.

The headline **90.1%** is really **86.1% pure rules + a 54-word hand-edited answer key** (the overrides account for exactly the 4-point gap, and most are relabelings of the dataset's own human answer, not fixes). On *broad* Dakshina data the engine matches any attested human spelling only ~38% of the time — the 86% is fit to a cleaned benchmark, not real-world skill.

None of this calls for a rewrite. The pipeline (`Parse → Prepare → Rules → Render`), the scoped/named/toggleable rule system, and the debug tooling are worth keeping and building on. What follows is where things stand precisely, and a three-horizon plan to move forward.

---

## 1. Where it stands (verified numbers)

All numbers re-verified by running the suite on 2026-09-04.

| Metric | Number | Note |
|---|---|---|
| Curated set, **pure rules** | **86.1%** (1150/1335) | The honest headline |
| Curated set, **with overrides** | 90.1% (1203/1335) | = 86.1% + 54 answer-key edits |
| With `--keep-medial-schwa` | 82.5% | This is what the stale session hook reports |
| **Full Dakshina** (all attest., match-any-variant) | **~37.7%** | Generalization reality |
| Aksharantar (357K rows) | 25.0% | |
| `--long-vowels` on curated | 60.4% | Blanket ā→aa is a net **loss** |

Three consequences:

- **The overrides are a lexicon in a CSV, not engine rules.** The engine never sees them; the harness swaps the expected answer. 29 of 54 are labelled "override from dakshina" (the human answer replaced with a gomanize-friendlier one). **86.1% pure is the number to quote.**
- **The "82.5%" in the session hooks is stale** — it is the Phase-1 target and coincidentally the `--keep-medial-schwa` score. Default config is genuinely 86.1/90.1.
- **The CI gate is 82%-with-overrides** — loose enough to hide a ~5-point regression. It should track *pure* accuracy.

**Two bugs found in passing:**
- `make test-dakshina` and `make test-unit` only exercise the *legacy* package — the real Dakshina integration test no longer exists, yet `test-dakshina` prints "✓ Dakshina test complete" while running zero tests.
- One override (कर्मकांड) contradicts its own engine output.

## 2. Architecture verdict: good bones, honest debt

**Strong.** A clean layered pipeline where `core/` knows nothing about Devanagari, `script/brahmic/` holds all script logic, and `lang/hindi/` holds all phonology. Rules are declarative, scoped (Universal/Script/Language/Scheme), priority-ordered, mode-typed (Exclusive/Always/Fallback), and runtime-toggleable, with genuine debug tracing (`--debug`, `--list-rules`). The `ConsonantRun.DeletedAt` "at most one deletion per run" invariant is a smart global constraint.

**Debt is concentrated exactly where the linguistics is hard:**
- **~10 of 21 rules are surface-pattern special cases** with hardcoded rune literals (~20) and — worse — matching on the *output* romanization (`BaseRom == "sh"`, ~15 sites) rather than source characters. Any symbol-table change silently changes rule behavior.
- Positional hacks ported from the legacy engine (`u.Start.Rune != 1`) and cross-rule coupling (the `ांव` carve-out that exists only because a later render rule handles it) — the classic signs of a leaking abstraction.
- **The scope/scheme abstractions are aspirational, not cashed in.** "Universal" and "Script" rules physically live in `lang/hindi/rules.go`, so a new language (Marathi/Nepali) would copy-paste 21 rules; no second scheme (IAST) exists; the symbol table bakes in colloquial choices, so IAST needs a second symbol map the interface doesn't provide.
- Reverse (Roman→Devanagari) is effectively a rewrite — the pipeline is lossy (schwa deletion, i/ī collapse, श/ष→"sh") and one-directional by construction.
- **Test coverage hole:** the `core` engine is well-tested with mocks, but `lang/hindi`, `script/brahmic`, `scheme/colloquial`, and `cmd` have **zero** direct unit tests — Hindi correctness is validated only end-to-end via the benchmark.

## 3. The real blockers (the ceiling)

The ~14% pure failures break into buckets with a clear message:

| Bucket | ~Count | Rule-fixable? |
|---|---|---|
| Medial **ee/oo** spelling (संगीत→*sangit* vs *sangeet*) | ~29 | ✅ Yes — biggest tractable win (~+2 pts) |
| Contextual **ā→aa** misses (गाने→*gane* vs *gaane*) | ~12 | ⚠️ Narrowly (blanket version drops to 60%) |
| **Medial schwa** retention (जनता→*janta* vs *janata*) | ~46 | ❌ Lexical/morphological |
| **Compound** schwa (उत्तराखंड→*uttarakhand* vs *uttrakhand*) | 15 | ❌ Needs morpheme segmentation |
| **Loanwords/names** (दीक्षित→*dikshit* vs *dixit*, वक्त→*waqt*) | ~13 | ❌ Needs a lexicon |

**The ceiling is now proven, not speculative** (see §4 research):
- Hand-written rule systems for Hindi schwa deletion top out around **94–95% per-schwa** (Wiktionary's rule module, the closest analog to gomanize).
- A small **gradient-boosted tree** with a ±5-char window hits **98.0% per-schwa / 97.8% word-level** (Arora et al., ACL 2020).
- The ~3–4 point gap *is* loanwords, compounds, and morphology — exactly gomanize's MISSING_SCHWA + EXTRA_SCHWA buckets (48% of failures).

**Gomanize is near the rule-based ceiling.** Realistic rule-only headroom on the curated set is ~89–91% pure.

## 4. The insight that reframes everything

**Half the "failures" are a benchmark artifact.** Romanization is many-to-one — there is no single ground truth:
- Human round-trip consistency on Dakshina has a **CER floor of 0.054** — humans disagree with *themselves* ~5% of the time.
- Even neural SOTA tops out at **60–72% single-reference top-1** on Hindi (IndicXlit 60.5%; GPT-4.5 70.7%). *Nobody* achieves ~95% single-reference — it is not achievable.
- Dakshina *ships multiple attested romanizations with counts per word*, and the field standard (Kirov et al., *Computational Linguistics* 2024; the NEWS shared tasks) scores with **minCER / match-any-attested-variant**, not one gold string.

So scoring `janta` "wrong" against `janata` is partly the harness's fault. **Adopting multi-reference evaluation would honestly lift true accuracy several points** and let the override CSV be reserved for genuine errors, not variant disagreements.

## 5. Research landscape (2020–2026)

**Schwa deletion — the hard core:**
- Rule-based line: Choudhury & Basu (2002); Narasimhan et al. (2004, ~89%); Tyson & Nagar (2009, 86–94%) — all conclude no single rule family suffices.
- **Arora, Gessler & Schneider (2020), ACL** — [paper](https://aclanthology.org/2020.acl-main.696/), [code+models](https://github.com/aryamanarora/schwa-deletion): binary per-schwa classification from orthography (±5 phone window), XGBoost **98.0% per-schwa**, vs Wiktionary rules 94.2%. Training data is public (McGregor lexicon, 36K labelled schwas). The paper *itself demonstrates reading the trees back as linguistic rules* — the distillation blueprint.

**Neural / LLM:**
- **IndicXlit** (AI4Bharat, EMNLP Findings 2023, [arXiv 2205.03018](https://arxiv.org/abs/2205.03018)): 11M-param char transformer. Ships an **Indic→Roman** model (Hindi 52.3% words / 59.2% named entities on Aksharantar). Distributable (pip `ai4bharat-transliteration`, CPU-capable, offline-batchable).
- GPT-4.5 / fine-tuned GPT-4o now beat IndicXlit ([arXiv 2505.19851](https://arxiv.org/pdf/2505.19851)).
- ByT5 beats subword models at word-level (CER 8–10).

**Hybrid precedent:** rule-extraction from GBDTs (Arora); statistical model stored in human-editable rule format (ACM TALLIP 2025, [10.1145/3720542](https://dl.acm.org/doi/10.1145/3720542)); noisy-channel + LM re-ranking (Roark 2020; Kirov 2024) — the +12% re-ranking win IndicXlit reports.

**Newer datasets (Dakshina era is over — repo archived read-only April 2026):**
- **Aksharantar** ([HF](https://huggingface.co/datasets/ai4bharat/Aksharantar)): **5,693 human-annotated Hindi test pairs** (4× Dakshina), CC-BY, with frequency-weighted (**AK-Freq** ≈ lyrics vocabulary) and named-entity slices. **The upgrade.**
- **COMI-LINGUA** (EMNLP 2025, [HF](https://huggingface.co/datasets/LingoIITGN/COMI-LINGUA)): 181K rows, CC-BY, parallel Devanagari/Roman sentences of *naturally-typed colloquial* Hindi — closest to gomanize's register (needs word-alignment extraction).
- **Gap worth owning:** no aligned song-lyrics romanization benchmark exists anywhere. A curated 500-line lyrics set would be a unique public contribution.

## 6. Recommendations & roadmap

Sequenced in three horizons (plus H0, the tooling foundation). Tracked in [`/TASKS.md`](../../TASKS.md).

### H0 — Absorb tooling & process from `../skein` (foundation, in progress → F-0001)
Skein has a mature, portable process substrate. Absorb what's language-agnostic:
- **`tools/tasks.py` + `./tools/tasks`** — zero-dependency single-file task tracker over `TASKS.md` ([canonical](https://github.com/budhash/tasks)). **Vendored in this change.** The roadmap below is already seeded into it.
- Fix the dead `make` targets; add a `docs/PROCESS.md`, a PR template (the "acceptance is a floor, not the review" pattern), and pre-commit parity (no-commit-to-main, large-file/whitespace guards).

### H1 — Fix the measurement (do first; ~days → F-0002)
You cannot improve what you mismeasure.
1. **Multi-reference eval** — score match-any-attested-variant / minCER against Dakshina's full attestation sets. Likely *raises* honest accuracy several points.
2. **Add CER** alongside top-1 (credits near-misses like gaana/gana; field-standard second metric).
3. **Adopt the Aksharantar Hindi test set** as primary; report AK-Freq / AK-Uni / NE separately (NE will expose proper-noun weakness the curated set hides). Keep Dakshina as frozen continuity.
4. **Tighten the CI gate** to track *pure* accuracy; make pure the headline.
5. Reconcile/prune the override CSV.

### H2 — Push the rule ceiling honestly (~1–2 weeks; → ~89–91% pure → F-0003)
6. **Medial ee/oo rule** (ी→ee, ū→oo medially; i/u finally) — biggest tractable win, ~+2 pts.
7. **Narrow** contextual ā→aa (never the blanket version).
8. Pay down the debt that blocks everything after: match rules on **source chars, not `BaseRom`**; extract universal/script rules out of `lang/hindi`; add the missing `lang/hindi` + `script/brahmic` unit tests.

### H3 — Break the ceiling with a learned component (the real "next level" → F-0004)
The research is unambiguous: past ~91% you must add a learned component, and both options stay offline/embeddable in Go.
9. **Distill Arora et al.'s schwa classifier into Go** — train shallow XGBoost on the public McGregor data, export trees, compile to dependency-free Go `if/else` (~50 lines of inference). Targets 48% of failures with a proven +3–4 pt technique. **Highest ROI.**
10. **Lexicon layer** — attestation-weighted top-~50K words (Dakshina + Aksharantar) with rules as OOV fallback (the standard TTS architecture). Handles the loanwords/names/compounds rules provably can't. The override CSV is this in embryo.
11. **Offline model-mining** — batch IndicXlit/LLM over the failure set to auto-propose override candidates where the model agrees with the dataset and gomanize disagrees. Pure offline tooling, no runtime Python dependency.
12. **Candidate generation + tiny n-gram re-ranker** — emit 2–4 candidates at ambiguous points, re-rank with a few-MB char 4-gram (the +12% mechanism, trivially embeddable in Go).

---

## One-line takeaway

Gomanize has hit the natural ceiling of pure rules (~86–91%), and the research confirms that *is* the ceiling for the approach. The next level is not more rules — it is **measuring honestly against multiple references (H1)** and **grafting one small learned component (a distilled schwa tree + a lexicon, H3)** onto the clean engine that already exists. Both attack the exact 48% of failures rules structurally cannot, and both stay offline and embeddable in Go.

---

### Sources
Arora et al. 2020 ([ACL](https://aclanthology.org/2020.acl-main.696/) · [code](https://github.com/aryamanarora/schwa-deletion)) · Narasimhan et al. 2004 · Tyson & Nagar 2009 · Aksharantar/IndicXlit ([arXiv 2205.03018](https://arxiv.org/abs/2205.03018) · [GitHub](https://github.com/AI4Bharat/IndicXlit)) · Roark et al. 2020 Dakshina ([arXiv 2007.01176](https://arxiv.org/pdf/2007.01176)) · Kirov et al. 2024 ([CL](https://aclanthology.org/2024.cl-2.2/)) · LLM benchmark ([arXiv 2505.19851](https://arxiv.org/pdf/2505.19851)) · COMI-LINGUA ([HF](https://huggingface.co/datasets/LingoIITGN/COMI-LINGUA)) · Hybrid statistical+rule ([ACM TALLIP 2025](https://dl.acm.org/doi/10.1145/3720542)) · NEWS 2018 shared task ([ACL](https://aclanthology.org/W18-2409/))
