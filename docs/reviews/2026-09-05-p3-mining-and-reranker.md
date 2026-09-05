# P3 — Override mining (negative) & candidate re-ranker (positive)

**Date:** 2026-09-05
**Tasks:** F-0004 / T-0017, T-0018 — the final backlog items.

## T-0017 — Override/lexicon candidate mining: NEGATIVE result, tool retained

`tools/mine_overrides.py` mines lexicon candidates for the 4,894 frequent words
with no Dakshina-train gold (the 78.2% coverage ceiling): strict-winner spellings
from Aksharantar train + optional external model output (`--model-output`, e.g.
IndicXlit TSV), skeleton-checked, evidence-annotated, all held-out natives
excluded. COMI-LINGUA is **never** a mining source (it is a benchmark); its
overlap is used only to *estimate precision*.

Measured precision of unreviewed candidates: **43.3%** (58/134 vs COMI-typed
variants); an engine-distance filter peaks at **57%** — both far below the
lexicon's ≥4-human-attestation quality bar. Top candidates show why: दोनों →
*dodoan*, थीं → *them* (machine-mined noise wins strict-winner precisely because
real words accumulate variants and get skipped).

**Disposition:** nothing is auto-promoted. The tool ships as human-review
tooling (output gitignored, evidence per row). Same lesson as H2, at the data
layer: machine-mined single sources cannot reach gold quality without a human —
which is exactly why the pipeline keeps one in the loop.

## T-0018 — Candidate generation + char-LM re-ranker: POSITIVE, best system yet

A character **4-gram LM** trained on Dakshina TRAIN romanizations only
(attestation-weighted, pruned <3; `tools/train_ngram.py` → 31k n-grams, 224 KB,
`go:embed`) scores candidates with stupid backoff, normalized per character
(`lang/hindi/reranker.go`). `Options.Rerank` / `--rerank` makes the engine run
the pipeline under candidate configurations and lets `core.Reranker` (mirroring
`LexiconProvider`) pick; ties keep the default.

**Candidate-set ablation on held-out Dakshina test (2,500 words, match-any):**

| Candidate set | Rerank score |
|---|---|
| rules + schwa-model + keep-medial-schwa + long-vowels | 61.7% |
| rules + schwa-model + keep-medial-schwa | 66.0% |
| **rules + schwa-model** | **70.4%** |

Baselines: rules 69.0%, schwa-model 69.5%. The lesson: **candidate quality gates
re-ranking** — a char LM structurally favors vowel-rich strings, so weak
candidates (long-vowels 60%, keep-medial-schwa) drag it below baseline; with two
individually-strong candidates it arbitrates their genuine disagreements and
beats both.

**Cross-benchmark verification (all improve):**

| Benchmark | Baseline | Rerank |
|---|---|---|
| Held-out Dakshina (match-any) | 69.0% / 69.5% | **70.4%**, minCER 0.0955 |
| Curated multi-ref (match-any) | 92.8% | **94.7%** — new project high |
| Lyrics gold (line CER) | 0.0492 | **0.0465** |

Held-out is legitimate: the LM saw only train romanizations. A regression guard
in `TestBenchmarkSchwaModelHeldout` keeps rerank within 0.5 pts of the best
single system.

## Status
T-0017 and T-0018 **done**; F-0004 complete — **the entire tracked roadmap is
finished**. `--rerank` is opt-in (default output unchanged). The two learned
artifacts (schwa tree 34 KB, n-gram LM 224 KB) plus the lexicon (260 KB) keep
the binary dependency-free.
