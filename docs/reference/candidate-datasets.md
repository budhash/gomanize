# Candidate Datasets (evaluated, not yet used)

A ledger of external Hindi datasets evaluated for possible use in gomanize.
Distinct from [`RESEARCH.md` §3](../RESEARCH.md), which lists only datasets
actually **in use** (benchmarks + the single training source). Anything here is
a candidate awaiting a decision; before any of it ships it must clear two gates:

1. **License** — a permissive, redistribution-compatible license (matching the
   discipline in RESEARCH §3). "No license" = all rights reserved = not
   committable. Aggregate statistics derived from text (e.g. word-frequency
   counts) are facts and far less sensitive than the text itself, but the
   provenance is still noted.
2. **Contamination** — learned components (lexicon, schwa model, re-ranker)
   train **only** on the Dakshina train split, and benchmarks are never mined
   (RESEARCH §2). A new source may feed a *held-out* set only if it is disjoint
   from Dakshina/Aksharantar/COMI-LINGUA — enforced mechanically by deduping
   against every `benchmark/data/*.csv` native column.

**The core gap none of these fill:** gomanize's real bottleneck is aligned
Devanagari↔Roman **song-lyrics** pairs. Every candidate below is either
monolingual Devanagari or Roman-only — useful for frequency weighting,
construct mining, or a re-ranker LM, but not a drop-in romanization gold set.

## Ledger

| Dataset | Type | Domain | Parallel Deva↔Roman? | License | Status / best use |
|---|---|---|---|---|---|
| **RAW.tsv** ([Mendeley cp6htsbbpp](https://data.mendeley.com/datasets/cp6htsbbpp/1)) | Devanagari **verse** | poetry (closest to lyrics) | No (Devanagari only) | Verify (Mendeley CC?) | On-domain construct mining for T-0044 context lines; **verify license before committing derived text** |
| **hi_part_1.txt** (Twine list; misnamed `.gz`, actually plain UTF-8) | Devanagari news/prose | web/news | No | Unclear | Frequency; construct mining (has encoding noise — filter to pure Devanagari) |
| **Hindi-Aesthetics-Corpus** ([GitHub](https://github.com/gayatrivenugopal/Hindi-Aesthetics-Corpus)) | Devanagari literary prose | novels/short stories (incl. translated Russian novels) | No | **None declared — blocker** | Local frequency only; **do not commit text**. Cite: Wairagade-Venugopal et al., IJACSA 11(1), 2020 |
| ↳ `freq.txt` (146K types, `word,count`) | word-frequency list | derived from Hindi-Aesthetics | n/a | Derived (counts = facts; corpus unlicensed) | Frequency ranking of mined tokens |
| ↳ `clean_lemma_freq.txt` (117K lemmas, `word<TAB>count`, freq-sorted) | lemma-frequency list | derived from Hindi-Aesthetics | n/a | Derived (as above) | **Primary source for T-0044 construct mining** — ranks candidates by real-world frequency |
| **hi_rom.txt** (Twine list) | **Roman** Hindi (noisy) | song titles / Latin | No (Roman only) | Unclear | Marginal: re-ranker LM only, needs heavy cleaning |
| **midas-research/hindi-nli-data** ([GitHub](https://github.com/midas-research/hindi-nli-data)) | Devanagari NLI pairs | entailment | No (not romanization) | Verify | Not applicable to romanization |
| **odaigen_hindi_pre_trained_sp** ([HF](https://huggingface.co/datasets/Hindi-data-hub/odaigen_hindi_pre_trained_sp)) | Devanagari pretraining corpus | mixed/web | No (monolingual) | Verify (HF card) | Frequency / LM; future evaluation |

## Lyrics-domain candidates (the target register)

These are song-lyrics sources — the register gomanize actually targets. None is
*confirmed* parallel Devanagari↔Roman (still the core gap), but several are
on-domain and one is permissively licensed and obtainable today.

| Dataset | Type | Script | Parallel? | License | Status / best use |
|---|---|---|---|---|---|
| **[Bollyrics](https://github.com/budhash/Bollyrics)** (fork of lingo-iitgn; [arXiv 2007.12916](http://arxiv.org/abs/2007.12916)) | Bollywood film lyrics, 1934–2019, year-partitioned | **Romanised Hindi** (Roman only) | **Apache-2.0** ✓ | **Strongest lead:** permissive + on-domain. Re-ranker LM training (Roman side); the Roman half of a pair if aligned to Devanagari |
| **aczoom iSongs** — "[The (Hindi) ITRANS Song Book](https://www.aczoom.com/isongs/hindi/)" | film-song lyrics, browsable pages (not a dataset) | **ITRANS** (romanized) | None visible — treat as ©, out-of-repo | ITRANS→Devanagari is *deterministic*, so this could yield Devanagari + a reference romanization — the Giitaayan path noted in RESEARCH §3. Keep out-of-repo (copyright) |
| **Sanchay** ([NeuroQuantology 2022](https://www.neuroquantology.com/open-access/Sanchay%3A+A+Literary+Dataset+of+Indian+Filmy+Songs_7942/)) | Hindi film-song lyrics + metadata (web-crawled) | Unspecified (paper says "transliterate") | Paper CC-BY-NC-ND; dataset access/terms unstated | Verify: no download link in the paper; script + size unknown |
| **LyricSense** ([UBC MDS](https://masterdatascience.ubc.ca/why-data-science/data-stories/lyricsense-multilingual-song-lyrics-corpus)) | multilingual lyrics (EN/FR/ZH/**HI**), 10K songs / 1K annotated | Unspecified for Hindi | Unstated — contact UBC | Verify: obtainability, script, and license all unknown |
| **[Bollywood-Hindi-Songs](https://www.kaggle.com/datasets/ranirathore/bollywood-hindi-songs)** (Kaggle, ranirathore) | Bollywood song lyrics/metadata | Unknown | Kaggle license unread | Verify: page is JS-rendered — needs Kaggle API/login to inspect columns, script, and license |
| **[Bollywood-Songs-Dataset](https://github.com/devensinghbhagtani/Bollywood-Songs-Dataset)** (devensinghbhagtani) | ~1,000 songs 2017–23, scraped AZLyrics/YouTube (`music_name, singer, album, release, lyrics, thumbnail`) | Likely Roman (AZLyrics) — verify | None (scraped ©) | Out-of-repo (scraped from copyrighted sources, no license); verify script |

## General / English lyrics corpora (future reference)

Not usable for **Hindi** romanization (no Devanagari or romanized-Hindi
content), but kept on file for a possible future *general* lyrics-corpus effort
— e.g. lyric-domain LM pretraining, genre/emotion tagging, or the English side
of a cross-lingual setup.

| Dataset | Content | Script | License | Note |
|---|---|---|---|---|
| [theelderemo/genius-lyrics-cleaned](https://huggingface.co/datasets/theelderemo/genius-lyrics-cleaned) | 3.18M Genius lyrics (title/artist/genre/year/lyrics), 15 genres | English only (non-English stripped) | MIT | Largest, cleanest, permissive — the go-to general English lyrics corpus |
| [mamoon-17/Emotion-Classification-Study](https://github.com/mamoon-17/Emotion-Classification-Study) | ~12.2K English lyric excerpts + GoEmotions | English | None | Emotion-classification study; useful as an annotated-lyrics reference |

## Notes

- **Translated-novel contamination:** the Hindi-Aesthetics corpus includes Hindi
  translations of Russian novels, so its frequency lists surface foreign proper
  names (रस्कोलनिकोव/Raskolnikov, स्विद्रिगाइलोव/Svidrigailov, सोन्या/Sonya).
  Romanization of these is arbitrary — prune them from any gold set.
- **Encoding artifacts:** the same lists contain malformed tokens where ई was
  split into र्+इ (कोर्इ, गर्इ). `tools/mine_constructs.py` rejects
  halant→independent-vowel sequences as malformed.
- **T-0044 pipeline:** `tools/mine_constructs.py` mines stratified,
  frequency-ranked construct tokens from these sources, dedupes against all
  benchmarks, and emits a candidate TSV for human validation. Only validated
  rows (native + human romanization = our own work) graduate into
  `benchmark/data`, with a citation.
