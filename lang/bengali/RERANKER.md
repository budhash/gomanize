# Experimental Bengali native selector

`selector.json` is the artifact selected by the
[cross-fitting experiment](../../docs/reviews/2026-10-04-bengali-b2-crossfit.md),
retrained once on the schema-2 training partitions (see the hash note below).
It is derived from Google's Dakshina dataset (CC BY-SA 4.0) and Google Bengali
pronunciation data (CC BY 4.0). Credit both sources; the derived selector is
**CC BY-SA 4.0**, independently of the repository's MIT code license. Source
URLs, attribution, license links and pinned hashes are in the
[training data record](../../training/data/bengali/README.md).

Artifact SHA-256:
`252781600c834993beaa303dd0a918a8c0cd51cf40428851919c47c363f40961` (retrained 2026-10-06 after excluding variant spellings
of held-out words from training; dev selection and held-out results unchanged —
see the [retrain record](../../docs/reviews/2026-10-06-bengali-retrain.md)).
Threshold **0.6** and maximum **eight** single-slot flips are frozen from dev
selection. No runtime training or test-time selection is performed.

Enable with both `SchwaModel: true` and `Rerank: true`. The selector uses the
model candidate, B1, then single inherent-vowel flips in source order. Identical
strings are deduplicated. Only a strictly higher preference fraction can replace
the current best; ties retain earlier candidates. These fractions are empirical
training frequencies, not calibrated probabilities.

Scope is the vowel model's supported simple-word inventory and default style.
`InherentVowelA`, `LongVowels`, `SimpleNasals`, or `KeepMedialSchwa` bypass selection
and preserve the caller's original pipeline. Unsupported words do likewise.
Lexicon hits win first; a miss allows ranking. Each candidate retains the caller's
scheme and rule enable/disable state. Candidate closures are local to the call.
A handled rerank returns no single-path debug trace, matching the legacy reranker
contract; a declined rerank retains the normal pipeline's debug trace.

Held-out match-any is 63.04%, versus the vowel model's 62.76%. Aksharantar improves
slightly, but BanglaTLit sentence CER worsens slightly while exact matches remain
unchanged. This remains experimental and opt-in: it never runs by default, and the CLI
(`--language=bengali --schwa-model --rerank`), npm and web expose it only on
request. See the [runtime evaluation](../../docs/reviews/2026-10-05-bengali-b2-runtime.md)
for all profiles, losses, caveats and reproduction.
