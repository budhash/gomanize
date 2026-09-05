> **HISTORICAL DOCUMENT (early ा→aa analysis; the definitive measured result is docs/reviews/2026-09-04-h2-vowel-length-experiments.md).** Statistics and plans below reflect the
> project as of 2025 and are NOT current. For today's status see `CLAUDE.md`;
> for the live backlog see `/TASKS.md` (via `./tools/tasks tree`); for every
> subsequent result and decision see `docs/reviews/`.

# 'aa' vs 'a' Pattern Analysis

Analysis of when Aksharantar dataset uses 'aa' vs 'a' for the ा (aa-matra) vowel sign.

## Key Finding 1: Agreement Rate

- **62.1% agreement** between gomanize and Aksharantar on 'aa' usage
- Gomanize is more conservative: uses 'a' when Aksharantar uses 'aa' in **27.6%** of cases
- Gomanize uses 'aa' when Aksharantar uses 'a' in only **10.2%** of cases

| Category | Count | % |
|----------|-------|---|
| Both use 'aa' (agreement) | 455 | 8.5% |
| Both use 'a' only (agreement) | 2,852 | 53.6% |
| Gomanize uses 'aa', Aksharantar uses 'a' | 544 | 10.2% |
| Gomanize uses 'a', Aksharantar uses 'aa' | 1,471 | 27.6% |

## Key Finding 2: Position Matters

| Position | Aksharantar uses 'aa' |
|----------|----------------------|
| Initial (start of word) | 24.5% |
| Medial (middle) | 34.6% |
| Final (end of word) | 37.5% |

**Pattern**: Aksharantar uses 'aa' more often at word-final and medial positions.

## Key Finding 3: What Follows ा

| Following Character | % with 'aa' |
|--------------------|-------------|
| Visarga (ः) | 80.0% |
| After ऐ/अ | 75-80% |
| Before ओ | 55.6% |
| Chandrabindu (ँ) | 50.0% |
| Anusvara (ं) | 43.7% |
| Consonant | 40.7% |
| Word-end | 33.1% |
| Before इ | 15.6% |
| Before उ | 8.6% |

## Key Finding 4: The `--long-vowels` Trade-off

- **1,388 words break** with `--long-vowels` (currently pass)
- **641 words fix** with `--long-vowels` (currently fail)
- **Net: -747 words** (worse overall on Aksharantar)

This confirms that Aksharantar has **inconsistent 'aa' conventions** - some annotators use 'aa' for all ा, others use 'a'.

## Future Flexibility Options

1. **Conservative (current default)**: ā→a everywhere except ा+C+END
2. **Liberal (`--long-vowels`)**: ā→aa everywhere
3. **Position-based (potential)**: ā→aa only in specific positions:
   - Before visarga (80% match)
   - At word-final position (37.5% match)
   - After specific consonants (h, k, y have highest correlation)

### Possible Implementation

```
--aa-style=conservative|liberal|balanced
```

Or position-aware rules that match higher-confidence patterns.

## Analysis Date

2025-01-06
