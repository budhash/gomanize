# Bengali B2 data isolation (T-0057)

The [frozen training manifests](../../training/data/bengali/README.md) establish
reproducible, normalized word-type partitions before B2 alignment or learning.
Google supplies 55,121 training types and Dakshina supplies 23,117 spelling
training types. Both exclude all Google and Dakshina dev/test types. The
protected-source loader and output lexicon guard are the entry points for B2.

The external audit found that Aksharantar includes the entire Dakshina test
vocabulary. This prevents treating their scores as independent confirmations.
BanglaTLit also shares many training words; later reports must separate full
and unseen-type performance. Neither supplies training references. Lyrics gold
is still T-0054; alignment ambiguities and artifact provenance are T-0053.

Validation: deterministic reconstruction from pinned Google and Dakshina inputs;
source-preserving Google loader returns 55,181 training pronunciation rows;
normalization, source-checksum, leakage, and output-lexicon guards are tested.
A deliberately injected held-out word with a recomputed manifest hash fails
semantic validation. Removing the exclusion assertion in a temporary copy makes
the regression test fail (`ValueError not raised`), confirming the guard. An
independent hash/partition recomputation also agrees with all frozen inventories.
The new offline isolation target is part of `make ci`.
There are no engine, rule, learned-model, or runtime-output changes in this PR.

Full `make ci` passed, including the frozen Hindi replay, accuracy benchmarks,
Bengali B1 gate and eight isolation tests. Repository pre-commit hooks also pass.

**Superseded figures (2026-10-06).** Training partitions were rebuilt (schema 2) and this
experiment's committed record regenerated; see the
[retrain record](2026-10-06-bengali-retrain.md) for current counts.
