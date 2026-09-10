# Chandrabindu before a labial: n vs m (T-0046)

**Date:** 2026-09-10 · **Status:** decided — **no rule change (rejected)** ·
**Trigger:** Codex review of T-0045 flagged that anusvara before a labial
becomes `m` (संबंध→sa**m**bandh) while chandrabindu stays `n` (साँप→saa**n**p),
and asked whether chandrabindu should assimilate too.

## Question

`render.anusvara.before-labial` (Language:35) maps anusvara **ं** to `m` before
प/फ/ब/भ/म (standard homorganic nasal assimilation). Chandrabindu **ँ** has no
such rule, so it renders its base `n`. Should we extend the labial-assimilation
rule to chandrabindu for consistency?

## Evidence (data-driven review)

Scanned every `benchmark/data/*.csv` for `ँ` immediately followed by a labial:
**70 native pairs** (English loans like *ambulance*/*desktop*, where `ँ`
transcribes a foreign nasal, excluded). The attested spellings **split by word**,
and the split is lexical, not phonological:

| tendency | words (attested) |
|---|---|
| **`n`** (dominant) | काँप→kaanp, साँप→saanp, हाँफ→haanf, धाँप→dhaanp, भाँप→bhaanp — the काँपना / हाँफना / साँप families, many rows each |
| **`m`** | ताँबा→taamba, सँभाल→sambhaal, साँभर→sambhar |

The `m` words are all etymologically **anusvara** words (तांबा, संभालना, सांभर)
that happen to be *spelled* with chandrabindu; the `n` words are true
nasalized-vowel tadbhava forms (/kãːp/, /sãːp/). So the two nasal marks genuinely
differ: **ं** is a homorganic nasal that assimilates to `m`; **ँ** is vowel
nasalization that does not.

## Decision — reject the rule

A blanket `ँ→m` before labials would fix 3 rare lexical items (ताँबा, सँभाल,
साँभर) while **breaking the common `n` class** (काँप/साँप/हाँफ/धाँप families —
far more, and far more frequent). Net-negative and phonologically wrong. The
current default (`ँ→n`, साँप→saanp) is correct for the dominant class, so the
`render.anusvara.before-labial` rule stays **anusvara-only**. The Codex-flagged
"inconsistency" is correct behaviour: the marks are not equivalent.

## Residue (not fixed here)

ताँबा→`tanba` (want *taamba*), सँभाल→`sanbhaal` (want *sambhaal*), साँभर→`sanbhar`
are genuine misses, but **lexical** (specific anusvara-origin words), not
rule-governed, and rare. They belong in the lexicon layer, not a render rule; the
lexicon is attestation-gated on the Dakshina train split (contamination
discipline, RESEARCH §2), so they are left as known lexical gaps rather than
hand-added. The `n`-class default is regression-guarded by साँप→saanp in the
held-out construct set (`benchmark/data/heldout_constructs_hi.csv`, T-0044).
