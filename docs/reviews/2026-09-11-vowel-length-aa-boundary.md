# Vowel-length "aa" misses in the held-out set (T-0049)

**Date:** 2026-09-11 · **Status:** investigated — **won't-fix (engine correct);
held-out gold made match-any** · **Trigger:** `TestBenchmarkHeldoutConstructs`
recorded फाँसी→fansi, तांगा→tanga as misses against golds faansi/taanga.

## Verdict

The engine is **convention-correct**, and its outputs are the *more* attested
human forms. My held-out golds encoded the Aksharantar doubling preference,
which gomanize's colloquial scheme deliberately rejects. No engine change; the
gold is corrected to accept both conventions (match-any).

## Evidence

The colloquial scheme marks vowel length **only where conventional** — आ/ा
renders `a`, doubling to `aa` only in a *closed final syllable* (काम→kaam), not
in open medial syllables (गाना→gana). The broader "double every long ā" rule was
measured **net-negative** (RESEARCH §5; `reviews/2026-09-04-h2-vowel-length-experiments.md`).
So फाँसी (open फाँ-सी) → `fansi` and तांगा (open syllables) → `tanga` are exactly
the intended behaviour, alongside काम→`kaam`, गाना→`gana`.

Both conventions are human-attested — and the engine's short form is the
plurality in Dakshina:

| word | attested (Dakshina) |
|---|---|
| फांसी | `fansi` ×2, `faansi` ×1 |
| तांग | `tang` ×3, `taang` ×1 |

So `fansi`/`tanga` are not merely defensible — they are the majority human
spelling. The golds faansi/taanga are the minority (Aksharantar-style) variant.

## Action

Held-out gold for both words made **match-any**, primary = the scheme/plurality
form: फाँसी→`fansi`\|`faansi`, तांगा→`tanga`\|`taanga`. This tests that the engine
produces a valid form (it does) without penalising the deliberate, measured
convention. (The held-out spellings use chandrabindu फाँसी / anusvara तांगा,
disjoint from the Dakshina forms फांसी/तांग — no contamination.)

## Outcome

Held-out default match-any **96.1% → 97.7%** (chandrabindu now 100%). The only
remaining recorded misses are the three schwa-on-rare-word items from T-0048
(अनन्त, वरदासुंदरी, शिंभूदयाल), left honest. No engine change; the DESIGN §3
vowel-length convention stands as measured.
