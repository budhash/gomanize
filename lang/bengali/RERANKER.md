# Experimental Bengali native selector

`selector.json` is the unchanged artifact from the
[cross-fitting experiment](../../docs/reviews/2026-10-04-bengali-b2-crossfit.md).
It is derived from Google's Dakshina dataset (CC BY-SA 4.0) and Google Bengali
pronunciation data (CC BY 4.0). Credit both sources; the derived selector is
**CC BY-SA 4.0**, independently of the repository's MIT code license. Source
URLs, attribution, license links and pinned hashes are in the
[training data record](../../training/data/bengali/README.md).

Artifact SHA-256:
`37215f8f26586d831e32d6c2cb84cca80526a8d1ed6edfc23180f6d2b8f7d5d2`.
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
unchanged. This remains experimental and opt-in, with no default or CLI language
change. See the [runtime evaluation](../../docs/reviews/2026-10-05-bengali-b2-runtime.md)
for all profiles, losses, caveats and reproduction.
