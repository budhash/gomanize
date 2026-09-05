# Track C — Real-world (frequency-weighted) validation

**Date:** 2026-09-05
**Goal:** Measure gomanize on the words people *actually use*, and quantify the
lexicon's production value — the gap the disjoint Dakshina splits left open (H3
lexicon record). Backed by dataset research (see below).

## Data adopted

- **`benchmark/data/freq_hi.csv`** — top 15,000 Devanagari word types by corpus
  frequency, from the **Shabd** psycholinguistic database (1.4B-token Hindi news
  corpus), **CC0 1.0** → freely committable in an MIT repo. Source:
  https://osf.io/xfbhd/ . Regenerate: `python3 tools/build_freq.py <shabd96k.csv>`.

## The honest real-world metric

`TestBenchmarkFrequencyWeighted` weights each word by its corpus frequency
(common words count more) and scores match-any against attested Dakshina
spellings on the frequency∩gold intersection (9,987 words):

| Metric (frequency-weighted) | Value |
|---|---|
| Match-any, **rules only** | **82.8%** |
| Match-any, **rules + lexicon** | **88.7%** (+5.9 pts) |
| Lexicon token coverage | **29.4%** of running-text tokens (1,279 / 15,000 words) |

This is not circular: words are weighted by real corpus frequency and scored
against human-attested gold. It shows the lexicon's genuine value — **+5.9 points
on real usage** — which the type-disjoint held-out benchmark structurally could
not credit (H3 lexicon record: 0% held-out coverage). Common words are both more
likely to be in the lexicon and more likely to be loanwords/names the rules miss.

## Headroom (from the research coverage curves)

The current lexicon (1,896 words, Dakshina attestations ≥4) covers 29.4% of
tokens. Computed coverage curves (Leipzig news / OpenSubtitles) show:

| Lexicon size | token coverage (news / spoken) |
|---|---|
| 5,000 | 90% / 96% |
| 10,000 | 95% / 98% |

So a **frequency-ranked lexicon of 5–10k high-confidence entries** (prioritizing
common words, sourced from Dakshina + Aksharantar + Xlit-Crowd) would lift token
coverage from ~29% toward ~90%+ — the clearest path to real-world quality. The
current lexicon is small only because it is limited to Dakshina's ≥4-attestation
words; frequency-ranking the source pool is the natural next step (not done here).

## Dataset research summary (2026)

**Frequency (for lexicon ranking & coverage):**
- **Shabd** (CC0, OSF) — adopted. The only CC0 list; embeddable with no obligations.
- Leipzig `hin_news_2022` (CC-BY) — good held-out corpus for coverage estimation.
- hermitdave/FrequencyWords OpenSubtitles (CC-BY-SA) — smaller but closest to the
  spoken/lyrics register; useful as a re-weighting signal.

**Lyrics eval (the flagship use case):**
- **No aligned Devanagari↔Roman lyrics dataset exists** (confirmed gap).
- Best build path: Devanagari from **Giitaayan / v9y/giit** (ITRANS→Devanagari,
  deterministic), human Roman from **LyricsTranslate**, distributed as a
  fetch-and-build script (lyrics are copyrighted; ship only public-domain +
  self-attested lines in-repo). A ~500-line gold set is a 1–2 day curation job and
  would be a unique public contribution.
- **COMI-LINGUA** (CC-BY-4.0) MT subset — sentence-aligned Devanagari↔Roman
  colloquial Hindi; the best *openly-licensed, redistributable* real-world set.
- **Xlit-Crowd** (14,919 human word pairs, CC-BY-NC-SA) — good secondary word-level
  eval beyond Dakshina.
- Avoid as gold: arXiv 2511.22769 and codebyam (machine/LLM-generated).

## Status
- Frequency-weighted eval shipped (`TestBenchmarkFrequencyWeighted`), lexicon
  value quantified. Follow-ups (tracked): frequency-rank/expand the lexicon toward
  5–10k entries; build the curated lyrics gold set; add COMI-LINGUA as a
  redistributable sentence benchmark.

## Update (T-0020, 2026-09-05): frequency-ranked lexicon expansion

Expanded the lexicon using the frequency ranking, still sourcing **only the
Dakshina train split** (dev/test stay out — contamination guard still reports
0/2,500 held-out words covered). Policy: keep the attestations-≥4 core
regardless of frequency; add top-15k-frequency words whose best train spelling
has ≥2 attestations and is a strict winner (ties skipped as ambiguous).

| Metric | Before (1,896 entries) | After (8,367 entries) |
|---|---|---|
| Token coverage | 29.4% | **71.1%** |
| Freq-weighted match-any (rules+lexicon) | 88.7% | **97.4%** |
| Embedded size | ~46 KB | ~260 KB |

Honest caveat on 97.4%: a lexicon hit emits an attested variant, so it matches
the reference set by construction — that is the *point* of a lexicon (emit the
human spelling for known words), but the number to watch is the pair
(coverage, rules-tail accuracy), not 97.4% alone. Max achievable coverage from
train gold within the top-15k is 78.2%; the remaining gap to ~90%+ needs gold
for frequent words outside Dakshina train (Aksharantar/Xlit-Crowd — future work).
