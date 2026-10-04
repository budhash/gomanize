# Frozen Hindi outputs

Baseline engine: `85fadb9`, before the F-0011 Part A refactor. These are observed
outputs, not gold romanizations or training data. Preserve incorrect outputs too:
the purpose is to detect any behavior change, not to measure aggregate accuracy.

The snapshot includes every distinct first-column input from all ten
`benchmark/data/*_hi.csv` files, including full Aksharantar, all Dakshina splits,
COMI-LINGUA, frequency, lyrics, held-out, curated, override and ignore inputs.
It also includes invariant constructs and whitespace/punctuation examples.
`manifest.json` records source SHA-256 hashes, input count, and the ordered option
profiles. Each compressed JSONL row stores its input and one exact output per
profile. Files contain at most 10,000 inputs each, sorted globally by input.

Verify from the repository root:

```sh
go test . -run '^TestHindiFrozenSnapshot$' -count=1 -v
```

The test fails on changed sources, profiles, missing rows, or any output delta;
it never silently skips missing datasets. Rule-disabled profiles exercise Pending
schwa rendering; lexical and learned profiles do not replace the rules-only net.

Regeneration is reserved for a separately reviewed behavioral change:

```sh
go test . -run '^TestHindiFrozenSnapshot$' -update-hindi-snapshot -count=1 -v
```

Record the baseline commit and explain per-input changes when regenerating.
Never regenerate to make a behavior-preserving refactor pass. Corpus sources
and their licenses are documented in `docs/RESEARCH.md`.
