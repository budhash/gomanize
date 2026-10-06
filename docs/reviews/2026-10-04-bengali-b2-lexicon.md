# Bengali B2 spelling lexicon (T-0062)

Bengali now supports an opt-in 8,980-entry spelling lexicon, independently or
combined with the vowel model. Default B1 output is unchanged. This completes
the spelling-lexicon portion of B2; candidate reranking remains T-0053 and lyrics
gold remains T-0054. The PR is stacked above #118 and stays draft for review.

## Selection and isolation

The fixed policy accepts a unique highest-vote spelling with at least three
attestations. The threshold was chosen from training-only vote distributions;
no external accuracy sweep or held-out spelling selection was performed. Of
23,117 approved native training types, 8,980 pass, 13,994 are below threshold,
139 have tied top spellings, and four contain digits rather than only Bengali
letters/marks. Roman case is preserved. The 246,620-byte TSV and manifest rebuild
exactly from the pinned training fixture.

Every Google and Dakshina dev/test key is excluded by T-0057's frozen policy.
Both Python output checks and Go runtime lookup tests require zero hits on all
four held-out sets. The Bengali-block normalizer handles Cf removal, the three
nukta aliases, split matras and canonical combining-mark order. It is checked
against 5,750 synthetic Python NFC cases, and every dictionary entry is tested
through aliases and the engine. No new Unicode dependency is introduced.

## Style-aware lookup

The engine adds the optional `OptionsLexiconProvider` interface. It takes
precedence over the existing `LexiconProvider`, including a declined lookup;
otherwise a style-aware miss could silently fall back to an incompatible
spelling. Languages implementing only the old interface, including Hindi,
retain their behavior.

Bengali declines lookup for any alternate-style flag: `InherentVowelA`,
`LongVowels`, `SimpleNasals`, or `KeepMedialSchwa`. It then uses the requested
rule/model pipeline. `SchwaModel` affects only fallback on unknown words; a
known default-style spelling wins first. Sentence whitespace and punctuation
are preserved. Lexicon hits retain the existing nil-debug-info contract.

## Primary word results

Pure results remain the headline: Bengali B1 test match-any is 56.52%; the
opt-in vowel model's pure test match-any is 62.76%. Lexicon-enabled dev/test
outputs are **identical word-for-word** to the corresponding pure outputs.
These sets have zero lexicon coverage; the unchanged score is required evidence
of isolation, not a failed attempt to learn unseen words.

| Dakshina test (2,500 words) | Strict | Match-any | CER | Lookup hits |
|---|---:|---:|---:|---:|
| B1 pure / B1 + lexicon | 800 (32.00%) | 1,413 (56.52%) | .08772388 | 0 |
| Model pure / model + lexicon | 914 (36.56%) | 1,569 (62.76%) | .07642207 | 0 |

Training coverage is 8,980/25,000 (35.92%) over the full original train split.
Its model-only match-any is 15,261/25,000 (61.04%); model + lexicon is
18,113/25,000 (72.45%). **This increase is in-sample dictionary coverage, not a
generalization result.** No training-score gain is used as the quality headline.

## External transfer

[Machine-readable results](2026-10-04-bengali-b2-lexicon-external.json) include
B1, model, lexicon-only and combined results plus source/artifact hashes.

| Evaluation | B1 pure | Model pure | B1 + lexicon | Model + lexicon |
|---|---:|---:|---:|---:|
| Aksharantar full, 7,134 types: match-any | 1,967 (27.57%) | 2,200 (30.84%) | 2,071 (29.03%) | 2,296 (32.18%) |
| Aksharantar full: CER | .22691465 | .21640005 | .22127213 | .21117049 |
| Aksharantar unseen, 6,036 types: match-any | 1,835 (30.40%) | 2,052 (34.00%) | 1,835 (30.40%) | 2,052 (34.00%) |
| BanglaTLit, 2,500 sentences: exact | 8 | 12 | 13 | 17 |
| BanglaTLit: CER | .33149376 | .31286250 | .29401742 | .28142890 |

Aksharantar has 210 lexicon-hit types. Relative to model-only output, lookup
changes 112 full-type outputs: 98 become matches and two cease to match (net
+96). On BanglaTLit it changes 1,477 sentences, gaining five exact matches and
losing none; sentence CER remains the more informative measure. Its unseen subset is defined conservatively
against the union of both frozen training vocabularies, and the evaluation tool
asserts that lookup cannot change any unseen word. Aksharantar also contains the
entire Dakshina test vocabulary, so these are not independent confirmations.

BanglaTLit provides sentence references, not isolated word labels. Its 38
sentences whose Bengali tokens are all unseen remain unchanged by lookup.
There is no claim of unseen-token accuracy from sentence references, nor of
lyrics readiness. The combined system reduces sentence CER by about 10.0%
relative to the model alone, while exact-sentence accuracy is still only 0.68%.

## Reproduction and validation

```sh
python3 tools/bengali/build_lexicon.py --verify
make test-bengali-lexicon
go test ./benchmark -run '^TestBenchmarkBengaliLexicon$' -count=1 -v
python3 tools/bengali/evaluate_external.py /path/to/ben.zip --lexicon
make ci
make npm-test
```

Tests cover unique-winner selection, normalization, preserved Roman case,
byte-for-byte reconstruction, all four source exclusions, option-provider
precedence, legacy compatibility, alternate-style bypass, fallback, debug
short-circuit behavior, and sentence boundaries. Attribution and the CC BY-SA
4.0 derived-data license are in [LEXICON.md](../../lang/bengali/LEXICON.md).

Negative controls: a temporary Go overlay that discards caller options causes
engine precedence, Bengali style-bypass and sentence-API tests to fail. Removing
the Python final-output isolation guard in memory makes the injected held-out
row test fail (`ValueError not raised`). No repository source or corpus is
changed by these controls. `make npm-test` passes, including the WASM build
and ESM/CommonJS Hindi smoke tests.

Full local `make ci` passed, including the frozen Hindi replay, accuracy
benchmarks, B1 gate, source isolation, vowel-model and lexicon checks. Hindi
curated pure strict remains 1,147/1,330 (86.2%) and match-any 1,236/1,330
(92.9%). The prior vowel-only external report also reproduces byte-for-byte.
Repository commit hooks pass.

## Review notes (2026-10-05)

Independent stack review. Tables above record the lexicon as submitted.

- **Numbers after the B1 review fixes.** Rule-composed outputs moved by one word
  (canonical nukta letters; explicit final hasant); the lexicon and model artifacts
  are unchanged. The external JSON was regenerated: Aksharantar full match-any B1
  1,968, lexicon 2,072, model 2,201, model + lexicon 2,297 (each +1); unseen B1 and
  lexicon 1,836, model and model + lexicon 2,053 (+1); BanglaTLit CERs move in the
  sixth decimal. Train (in-sample): model 15,262, model + lexicon 18,114 match-any
  (each +1). Held-out results are unchanged.
- **Cross-report consistency is asserted.** The lexicon report's B1 and model
  columns must equal the vowel-model report's, so a stale rerun of either fails.
- **Train lexicon outputs are pinned** (`bnPinnedLexiconTrain`: strict, match-any,
  hits, output digest); dropping eleven entries now fails (hits 8,980 → 8,969).
- **Variant-spelling collision.** `আবার` collides with held-out অাবার under the
  stricter collision key; runtime lookup misses it, so held-out outputs are
  unaffected. Pinned and disclosed in [LEXICON.md](../../lang/bengali/LEXICON.md);
  `assert_lexicon_isolation` now rejects any other such collision.
- **Licensing.** The lexicon is embedded in every build (including Hindi-only CLI
  and WASM). The README License section now states the embedded lexicons are
  CC BY-SA 4.0; release archives and the npm license field are handled with the
  frontend NOTICE work.
- Tracked: the Bengali lexicon style bypass is an allowlist of today's four style
  options; a malformed embed becomes an empty map (caught by the size test).
