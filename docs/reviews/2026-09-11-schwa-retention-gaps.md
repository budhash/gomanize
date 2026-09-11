# Schwa-retention "gaps" in the held-out set (T-0048)

**Date:** 2026-09-11 · **Status:** investigated — **no engine change; held-out
gold corrected** · **Trigger:** `TestBenchmarkHeldoutConstructs` recorded misses
अंततः→anttah, दरअस्ल→darasl, बाईं→bain that looked like schwa bugs.

## Method

Measure-first (task discipline): debug-trace each miss, check what the corpora
actually attest, and only touch the schwa rules if a candidate fix helps the
Dakshina pure gate. Result: **none needed an engine change** — the "gaps" were
held-out gold-quality issues or legitimate variants.

## Findings, case by case

| native | engine | my gold | verdict |
|---|---|---|---|
| दरअस्ल | darasl | ~~darasal~~ | **my gold was wrong** — I romanized the *halant-less* word दरअसल. The mined token has an explicit halant (स्ल conjunct); `darasl` is correct for that literal string. Gold → `darasl`. (दरअसल itself is in Dakshina + the lexicon, i.e. training data — which is why the miner picked the halant variant; it can't be swapped in without contaminating the held-out set.) |
| बाईं | bain | ~~bai~~ | **my gold was wrong** — `bai` dropped the anusvara nasal. `bain` (cf. बायीं→bayin) is correct. Gold → `bain`/`baain`. |
| अम्मां | ammaan | ~~amma~~ | **my gold was wrong** — `amma` (= अम्मा) dropped the nasal; the native अम्मां has anusvara → `ammaan`. Gold → `ammaan`/`amma`. |
| अंततः | anttah | antatah | **legitimate variant** — the medial schwa deletion is human-attested for this visarga-tatsama class (Dakshina has मूलतः→**multah** alongside moolatah). Gold → match-any `antatah`\|`anttah`. |

## The schwa rules are not the problem

अंततः→anttah comes from `schwa.delete.ccv` (the same rule behind जनता→janta),
and गomanize treats the whole visarga-tatsama class consistently
(मूलतः→multah, अधिकांशतः→adhikanshtah) — forms Dakshina's own annotators
produce. Changing the ccv rule to force retention would fight attested human
behaviour and risk the pure gate, for no clear gain (RESEARCH §4: the eight
schwa rules are near-optimal). So the rule stands; the gold accepts the variant.

## Recorded misses left in place (deliberately, honest)

Three related misses were **not** rigged to pass — the evidence does not support
the engine's output as clearly valid, so they stay recorded:

- **अनन्त → annt** (want anant): the common spelling अनंत→`anant` is correct and
  Dakshina-attested; the rare conjunct spelling अनन्त deletes the schwa (न+न्त
  ccv) → `annt`, which humans do **not** attest here. A genuine, minor rare-word
  limitation — belongs in the lexicon, not a ccv-rule change.
- **वरदासुंदरी → vardasundri**, **शिंभूदयाल → shimbhudyaal**: proper names with
  optional-schwa divergence (sundari/sundri, dayaal/dyaal), no attestation either
  way — left strict rather than accept a debatable form.

The held-out set is designed to **record** such misses, not hide them.

## Outcome

Held-out default match-any **93.0% → 96.1%** (indep-vowel and visarga now 100%),
purely by correcting gold-quality issues — no engine change, no contamination,
pure gate untouched. Vowel-length misses (फाँसी, तांगा) are the separate T-0049.
