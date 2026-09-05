# H3 — Learned schwa classifier (distilled to Go)

**Date:** 2026-09-04
**Task:** F-0004 / T-0015. A learned schwa-deletion component that stays fully
offline and embeddable in Go, following Arora et al. (ACL 2020)'s "trees →
readable rules" blueprint — but trained on data we already have.

## Approach

The paper trains on the McGregor lexicon (external). We avoid that download by
**force-aligning the Dakshina train split** (which we ship) to derive per-schwa
delete/keep labels, then evaluating on the **disjoint test split** for a genuine
held-out generalization number.

Pipeline (pure stdlib, no ML dependencies):
1. `tools/schwa/align.py` — align (native, roman) pairs; for each consonant
   carrying an inherent schwa, label it deleted/kept from the romanization.
2. `tools/schwa/features.py` — orthographic features from the raw rune sequence
   (`cons, prev, next, next2, first, last`). **Kept byte-identical to the Go side.**
3. `tools/schwa/train.py` — a small CART decision tree (Gini, equality splits),
   exported to `lang/hindi/schwa_tree.json` (~34 KB).
4. `lang/hindi/schwa_model.go` — `go:embed` the tree, compute the same features
   from `Word.Original` + `Unit.Start.Rune`, predict. Zero runtime deps.
5. `schwa.model.predict` rule (Language:90, Exclusive, gated on the `SchwaModel`
   option / `--schwa-model`) takes over all inherent-schwa decisions when enabled.

## Validation

**Alignment** (train, attestations ≥ 2): 81.4% of pairs align → 24,473 schwa
instances at a **56.7% deletion rate** — closely matching Arora et al.'s 52.94%
on McGregor, and per-consonant rates are linguistically sane (न 79%, र 60%).

**Per-schwa accuracy** (decision tree, depth 18 / min-leaf 6):

| Split | Accuracy |
|---|---|
| Test (held-out, disjoint natives) | **90.67%** |
| Dev | 90.44% |
| Train | 91.76% |
| Majority baseline (always delete) | 58.65% |

Small train↔test gap → generalizes; far above the 58.65% baseline.

**Word-level, Dakshina TEST split (2,500 held-out words, match-any attested):**

| System | Match-any | Mean minCER |
|---|---|---|
| Heuristic schwa rules | 69.0% | 0.0972 |
| **Learned schwa model** | **69.5%** | **0.0960** |

## Finding

The learned model **slightly beats** the hand-tuned heuristics on genuinely
unseen words (+0.5 pts match-any, lower CER), at 90.7% per-schwa. Two honest
readings:

1. **The existing schwa rules are near-optimal.** A data-driven model trained
   from scratch matches and marginally exceeds eight hand-written rules — strong
   evidence the schwa heuristics are already at their ceiling, consistent with
   the literature (rule systems ~89–95% per-schwa).
2. **The remaining word-level gap is not schwa.** Improving schwa moved the
   held-out number by only +0.5 pts because ~31% of errors are vowel-length and
   lexical (loanwords, names) — exactly what H2 found. This points the next
   effort at the **lexicon layer (T-0016)**, not more schwa work.

The model is a legitimate, maintainable alternative to the schwa rule stack, and
the pipeline is a foundation: better alignment (currently 81%) or phonological
features (Arora reaches 98% with them) would raise it further without changing
the Go integration.

## Status
- Opt-in via `--schwa-model` / `core.Options{SchwaModel: true}`. Default behavior
  is byte-identical (heuristic rules; curated pure still 86.1%).
- T-0015 **done**. T-0016 (lexicon layer) remains the highest-value next step for
  real-world token coverage; T-0018 (candidate re-ranker) could combine model +
  rules per-word. Retraining: `python3 tools/schwa/train.py` → re-embeds the tree.
