# Keel K1: one canonical spelling per word (2026-10-09)

**Status:** done. Fixes keel finding K1 (and the mixed-token residue of K2) from
the [Bengali milestone keel review](2026-10-07-bengali-keel.md).

## Problem

Canonically equivalent Hindi spellings produced different output. The lexicon was
looked up on raw input (format characters and precomposed nukta letters missed),
the schwa model did not recognize precomposed nukta letters (U+0958–U+095F) and
fell back to the rules, and nukta/virama order was not normalized. Bengali avoided
this only through three language-local normalizers.

## Change

- `brahmic.Canonicalize` brings a word to the script's NFC form without external
  dependencies, from a per-language `CanonicalForms` table: composition-excluded
  letters decompose (क़…य़; ড় ঢ় য়), canonical pairs compose (ऩ ऱ ऴ; split ো ৌ),
  combining marks are put in canonical order (nukta before virama). Format
  characters (ZWJ, ZWNJ, BOM, soft hyphen) are removed only next to script
  characters, so emoji sequences keep their ZWJs even inside a mixed token.
- The engine applies it (`core.Canonicalizer`) before lexicon lookup, candidate
  ranking and parsing; the parser and the Hindi lexicon key use the same function
  (idempotent, with an allocation-free fast path for already canonical input).
  Hindi lexicon keys are canonicalized at load (all 8,367 were already
  canonical). Combining classes come from the Unicode data for each script's
  marks, including Vedic marks.
- Residue (tracked): Bengali's local `bengaliLexiconKey` and `vowelWordView`
  still exist; they now only ever see canonical input and can be retired in a
  follow-up.
- `TestCanonicalEquivalenceAllProfiles` checks equivalent spellings across six
  option profiles in both languages; with K1 reverted it reports 25 failures.
  `TestEngineCanonicalizesBeforeEveryComponent` checks that lexicon, reranker
  and parser all receive the canonical spelling (fails with the hook disabled).
- An independent review compared `Canonicalize` with Python NFC (after Cf
  removal) over every dataset word and ~700k synthetic strings: no disagreements
  once Vedic-mark classes were added. Long format-character or mark runs are
  linear (security review).

## Effect on output

**Headline accuracy is unchanged:**

| Metric | Value |
|---|---|
| Curated, pure / match-any | 86.2% / 92.9% |
| Held-out rules / schwa model / rerank | 69.3% / 69.8% / 70.7% |
| Aksharantar slices | unchanged |
| COMI-LINGUA | unchanged |
| Lyrics CER | 0.0492 (0.0394 with lexicon) |
| Frequency-weighted, rules / rules + lexicon | 82.8% / 97.4% |
| Bengali pins and records | unchanged |

Lexicon token coverage rises from 71.1% to 71.7% (45 more frequent words now hit
the lexicon). The full Dakshina bulk run moves 78,099 → 78,097 of 309,651 rows.

**Hindi frozen snapshot (deliberate update).** 110 of 264,064 inputs change, all
non-canonically encoded:

| Class | Inputs | Typical change |
|---|---:|---|
| Precomposed nukta (U+0958–U+095F) | 66 | The schwa model now applies (it previously fell back to rules), so model profiles give the same output as the decomposed spelling, e.g. उत्पीड़न `utpidan` → `utpidn` with the model, as main already gave for the decomposed form |
| Virama before nukta | 26 | Mostly corrections (उज़्ज़ा `ujza` → `uzza`, एक़्यूलिनो `ekyulino` → `eqyulino`); malformed Dakshina spellings such as द़्अष्टि change arbitrarily |
| Decomposed ऩ/ऱ/ऴ | 18 | Typo spellings in Dakshina (कमीशऩ, क्लोजऱ) now compose and match the precomposed form |

Changed outputs per profile: default 23, schwa model 102, lexicon 30, rerank 42,
lexicon + rerank 49, learned + style 102, every rules-only profile 23. Each changed
output equals what the engine produces for the canonical spelling, so no new
behavior is introduced; non-canonical input simply stops diverging.
