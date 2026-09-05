# Research Background

Why Hindi romanization is a hard problem, how it is evaluated,
what data it stands on, and what the experiments actually showed — including the
negative results. Primary sources for every claim are the dated decision records
in [`reviews/`](reviews/).

---

## 1. The problem: romanization has no single ground truth

Devanagari→Roman transliteration is **many-to-one**. Real people validly write
जनता as *janata*, *janta*, or *janataa*; नहीं as *nahin* or *nahi*; में as *mein*
or *me*. Concretely, in the Dakshina lexicon **49%** of common Hindi words carry more than one human-attested romanization
(avg 1.64 variants, max 5).

Two consequences follow:

1. **Scoring against a single gold string under-counts correctness.** Half of an
   engine's "errors" can be benchmark artifacts — valid spellings that happen to
   differ from the one reference kept.
2. **There is a floor on achievable agreement.** In Dakshina's own round-trip
   validation, humans re-transcribing romanized text disagreed with the original
   at **CER ≈ 0.054** (Roark et al. 2020). No system can meaningfully beat the
   rate at which humans disagree with themselves. Even neural SOTA tops out at
   60–72% single-reference top-1 on Hindi (IndicXlit 60.5%; GPT-4.5 70.7%).

The second hard sub-problem is **schwa deletion**: Hindi orthography implies an
inherent 'a' after every consonant, but spoken Hindi deletes many of them
(जनता → *janta*, not *janata*), governed by phonotactics, morphology, and
lexical convention. The literature's verdict:

| Approach | Per-schwa accuracy | Source |
|---|---|---|
| Hand-written rules (Narasimhan et al. 2004) | ~89% | IJST 7(4) |
| Prosodic rules (Tyson & Nagar 2009) | 86–94% | IJST 12(1) |
| Wiktionary's rule module (best rule system) | 94.2% | Arora et al. 2020 |
| **Gradient-boosted trees, ±5-char window** | **98.0%** | Arora, Gessler & Schneider, ACL 2020 |

The residual 3–4 points between the best rules and the learned classifier are
loanwords, compounds, and morphology — information surface rules cannot see.
This is the empirical basis for gomanize's architecture: rules to their ceiling,
then small learned components for the lexical gap.

## 2. Evaluation methodology

### Metrics
- **Strict top-1 (pure)** — exact match against one reference, no overrides.
  Deliberately harsh; used as the CI regression gate (floor: 85%).
- **Match-any-attested-variant** — correct if the output equals *any* human-
  attested romanization of the word. The headline metric, and the field standard
  (NEWS shared tasks; Kirov et al., *Computational Linguistics* 2024).
- **CER / minCER** — normalized edit distance (against the nearest attested
  variant). Credits near-misses; comparable to the ≈0.054 human floor.
- **Token-weighted** — words weighted by corpus frequency, so common words count
  proportionally to real usage. Used for the lexicon's production value.

### Honesty rules (enforced in code and process)
1. **Overrides are not accuracy.** `override_hi.csv` is an exception list; the
   historical "90.1%" headline was 86.1% pure + 53 hand-edited answers. Pure and
   match-any are what get reported; CI gates on pure.
2. **Contamination discipline.** All learned artifacts (schwa tree, lexicon,
   n-gram LM) train on the Dakshina **train** split only. Dakshina's splits are
   type-disjoint (0 shared words), so held-out results are genuine
   generalization. The lexicon coverage test *asserts* 0% held-out coverage.
3. **Benchmarks are never training/mining sources.** COMI-LINGUA overlap may
   estimate a miner's precision, never feed the lexicon.
4. **Every proposed change is measured before shipping**, and negative results
   are recorded (see §5).

## 3. Datasets

| Dataset | Size (as used) | License | Role |
|---|---|---|---|
| [Dakshina](https://github.com/google-research-datasets/dakshina) (Google, 2020; repo archived 2026) | 53K rows; 25K/2.5K/2.5K disjoint splits with attestation counts | CC BY-SA 4.0 | Curated benchmark (1,330 high-attestation words), multi-reference variant sets, held-out test, and the sole training source for all learned components |
| [Aksharantar test set](https://huggingface.co/datasets/ai4bharat/Aksharantar) (AI4Bharat, 2022) | 10,112 human-annotated pairs; slices AK-Freq / AK-NEF / AK-NEI / Dakshina | CC-BY 4.0 | Independent human benchmark (Karya native-speaker annotators); exposes named-entity weakness and cross-dataset convention shift |
| [COMI-LINGUA](https://huggingface.co/datasets/LingoIITGN/COMI-LINGUA) MT split (IIT-GN, EMNLP Findings 2025) | 9,606 word types / 152K token occurrences extracted from expert-annotated parallel Devanagari↔Roman sentences | CC-BY 4.0 | *Naturally-typed* colloquial Hindi — the closest proxy to song lyrics; token-weighted scoring |
| [Shabd](https://osf.io/xfbhd/) (psycholinguistic DB, 1.4B-token corpus) | Top-15K Devanagari words by frequency | CC0 | Frequency weighting; lexicon ranking; token-coverage estimation |
| Lyrics gold seed (in-repo) | 43 lines: Kabir, Rahim, Meera, Vande Mataram, Sarfaroshi, Raghupati, Jana Gana Mana | Public domain (life+60, India); maintainer-attested romanizations | Line-level evaluation of the primary use case through the sentence API |
| Aksharantar bulk (`aksharantar_hi.csv`) | 357K machine-mined rows | CC-BY/CC0 | Bulk smoke test + override-miner source (NOT gold — ~93% mining precision upstream) |

Data pipelines are reproducible: `tools/build_*.py`, `tools/train_ngram.py`,
`tools/schwa/train.py` regenerate every derived artifact from sources.

A survey (Sept 2026) found no aligned Devanagari-Roman song-lyrics dataset;
modern film lyrics are copyrighted. The in-repo public-domain seed is a partial
substitute; expansion paths are
Giitaayan (ITRANS, deterministic conversion) for Devanagari and LyricsTranslate
for human romanizations, kept out-of-repo for copyright reasons.

## 4. Results

### Headline (curated Dakshina, 1,330 words)

| Metric | Score |
|---|---|
| Match-any + `--rerank` | **94.7%** |
| Match-any, default rules | 92.8% |
| Strict top-1, pure (CI gate ≥85%) | 86.1% |
| Mean minCER | 0.0116 (human floor ≈ 0.054) |

### Generalization & real-world

| Benchmark | Default rules | Best configuration |
|---|---|---|
| Held-out Dakshina test (2,500 unseen words) | 69.0% | **70.4%** (`--rerank`), minCER 0.0955 |
| COMI-LINGUA, token-weighted | 78.9% | **85.7%** (`--lexicon`) |
| Frequency-weighted (Shabd ∩ gold, 9,987 words) | 82.8% | **97.4%** (`--lexicon`) |
| Lyrics gold, mean line CER | 0.0492 | **0.0394** (`--lexicon`); 0.0465 (`--rerank`) |
| Aksharantar AK-NEI (Indian names) | 43.0% | 50.9% (`--lexicon`) |
| Aksharantar AK-Freq | 42.6% | — (convention shift, see below) |

### Learned-component numbers
- **Schwa classifier**: CART tree trained on 24,473 force-aligned schwa
  instances (56.7% deletion rate ≈ Arora's McGregor 52.9%); **90.67% per-schwa
  on the disjoint test split** (majority baseline 58.65%); word-level it ties/
  slightly beats the eight hand-written schwa rules — evidence those rules are
  near-optimal.
- **Lexicon**: 8,367 entries (best train spelling, attestation-gated), **71.1%
  token coverage**; +5.9 to +7.9 points on three independent evaluations;
  provably 0% held-out type coverage (by split design).
- **Re-ranker**: char 4-gram LM (31K grams) over {rules, schwa-model}
  candidates; improved all three benchmarks it was measured on — held-out,
  curated multi-reference, lyrics CER (ablation in §5).

### Cross-dataset convention shift (why AK-Freq is "only" 42.6%)
Aksharantar's annotators systematically prefer doubled vowels
(*atyaachaarapoorn*) where Dakshina's curated set prefers single
(*atyacharpurn*); ~29% of AK-Freq failures are pure aa/ee/oo/v-w convention
differences, and the re-included Dakshina slice scores 68.8% — matching the project's own
held-out 69.0% and validating the harness. Romanization conventions differ
**between annotation efforts**, not just between annotators. For scale:
IndicXlit — an 11M-parameter transformer trained on Aksharantar itself —
reports ~52% top-1 in this direction.

## 5. Negative results (kept deliberately)

1. **Medial ee/oo rule** (ी→ee, ू→oo word-medially): every variant net-negative
   (match-any 92.8% → 89.0–91.2%). Medial long vowels are a ~50/50 attested
   lexical split; no positional rule can satisfy both halves.
   → `reviews/2026-09-04-h2-vowel-length-experiments.md`
2. **Broader ā→aa** (medial open syllable): 92.8% → 80.0% (भारत→*bhaarat* is
   wrong). The blanket `--long-vowels` option scores 60.4%.
3. **Unreviewed lexicon mining**: candidates mined from Aksharantar-train
   measure **43% precision** (57% with an engine-distance filter) vs
   independently-typed spellings — far below the ≥4-human-attestation bar.
   Machine-mined single sources cannot reach gold quality without human review.
   → `reviews/2026-09-05-p3-mining-and-reranker.md`
4. **Re-ranker candidate ablation**: with 4 candidate configs the re-ranker
   *drops* to 61.7% (a char LM structurally favors vowel-rich strings); with 3,
   66.0%; only with 2 individually-strong candidates does it win (70.4%).
   Candidate quality gates re-ranking.

Together these bound the design: the rule-based ceiling is real (~86% pure /
~93% match-any on this data), the remaining gap is lexical, and closing it
requires human-attested data, not more rules.

## 6. Key literature

- Arora, Gessler & Schneider (2020). *Supervised Grapheme-to-Phoneme Conversion
  of Orthographic Schwas in Hindi and Punjabi.* ACL.
  [aclanthology.org/2020.acl-main.696](https://aclanthology.org/2020.acl-main.696/) ·
  [code/data](https://github.com/aryamanarora/schwa-deletion)
- Roark et al. (2020). *Processing South Asian Languages Written in the Latin
  Script: the Dakshina Dataset.* LREC.
  [arXiv:2007.01176](https://arxiv.org/abs/2007.01176)
- Madhani et al. (2023). *Aksharantar: Open Indic-language Transliteration
  datasets and models.* Findings of EMNLP.
  [arXiv:2205.03018](https://arxiv.org/abs/2205.03018) ·
  [IndicXlit](https://github.com/AI4Bharat/IndicXlit)
- Kirov et al. (2024). *Context-aware Transliteration of Romanized South Asian
  Languages.* Computational Linguistics 50(2).
  [aclanthology.org/2024.cl-2.2](https://aclanthology.org/2024.cl-2.2/)
- Sheth et al. (2025). *COMI-LINGUA.* Findings of EMNLP.
  [arXiv:2503.21670](https://arxiv.org/abs/2503.21670)
- *Beyond Specialization: Benchmarking LLMs for Transliteration of Indian
  Languages* (2025). [arXiv:2505.19851](https://arxiv.org/abs/2505.19851)
- Narasimhan, Sproat & Kiraz (2004). *Schwa-deletion in Hindi TTS.* IJST 7(4).
- Choudhury & Basu (2002). *A Rule Based Schwa Deletion Algorithm for Hindi.*
- *Shabd: A psycholinguistic database for Hindi* (2021). Behavior Research
  Methods. [osf.io/xfbhd](https://osf.io/xfbhd/)

Full provenance for every number above: [`reviews/`](reviews/) (2026-09-04 onward).
