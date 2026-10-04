# Bengali B1 structural prototype

This is T-0056, stacked above B0. Before implementing the prototype, its gate is:

- Keep the public Bengali B0 rule catalog and corpus outputs unchanged. Prototype
  rules live only in tests; passing examples does not justify broader rules.
- Prove RepeatPrevious and GeminateSelf independently, including aspirated pairs,
  following matras, final-onset rewrites, explicit vowel retention/deletion,
  and rejected multi-character/already-clustered cases.
- Prove the exact design contexts মধ্য, দুঃখ, হও, হওয়া (both য় aliases), plus
  independent ও and unrelated vowel/য় contexts. Each rule owns its own unit.
- Show metadata-only Render rule traces in normal and fallback passes, no traces
  for unchanged non-Schwa rules, and no metadata extraction with debug disabled.
- Preserve all frozen Hindi outputs, pass full CI, and demonstrate negative
  controls for the new rendering and tracing behavior.

These structural gates are independent of the eventual B1 accuracy gate.
T-0055 remains open: the baseline's whole-word o/a support is not a positional
vowel distribution, and its 103 four-attestation dev words do not establish a
curation threshold. No expanded catalog or accuracy tuning is included here.
The B1 accuracy bar must still be recorded before T-0052 starts.

## Implementation and focused evidence

`BrahmicData.Gemination` is zero by default. Render-time pairing preserves source
identity and derives the pair from the final onset, rejecting atomic units,
multi-rune onsets, dead consonants, and larger clusters. A phala owns the right
member and its vowel slot; self-gemination uses its own onset. Rules do not mutate
neighbors. Non-default gemination and quality appear in Brahmic debug metadata.

The rule engine compares before/after metadata in both passes and avoids calling
the extractor with debug disabled. The public Bengali catalog is unchanged.
`lang/bengali/prototype_test.go` contains the exact source-scoped prototype;
`make test-bengali` now includes those tests. These examples validate machinery,
not the linguistic coverage of a future general rule.

Focused core/Brahmic/Bengali tests and the Bengali benchmark passed. Public dev
match-any remains 48.40%; test match-any remains 50.08%, strict 28.88%, and macro
minCER 0.10548. No corpus failures were mined for prototype rules.

Two temporary Go overlays served as sanity reverts without editing workspace
sources: bypassing onset rendering failed gemination/vowel-ownership tests;
restoring the old trace filter failed both normal and fallback metadata-only
trace tests. Debug-off and unchanged-rule trace tests pass.

An additional full-output comparison ran all 30,000 Bengali native types in
both vowel styles with current rendering and with the new onset path bypassed.
All 60,000 outputs matched (SHA-256 of split/input/style/output records:
`f5033d152ce317a1bf31d8d2ff9362e22669df12db748cfbbfaf3efc0dce6c15`).
`make npm-test` also passed the WASM build and ESM/CommonJS smoke tests.

Full `make ci` passed on 2026-10-04: formatting, lint, build, race/coverage,
all 3,432,832 frozen Hindi outputs, and accuracy benchmarks. Hindi pure strict
remains 1,147/1,330 (86.2%), match-any 1,236/1,330 (92.9%), and held-out
126/129 (97.7%). T-0056's structural gate is complete; T-0055 and T-0052
remain open.

The root race/coverage package took 615.354s (B0 run: 551.038s). These are
individual CI timings, not a controlled throughput benchmark.
