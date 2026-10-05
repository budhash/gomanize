# Claude handover: Bengali review and staged merges

Prepared 2026-10-05. Read this before reviewing or merging the Bengali stack.
The user wants **each stacked PR reviewed**, followed by a controlled merge
process. This handover does not merge anything, mark PRs ready, enable auto-merge,
release a package, or deploy a site.

## 1. Where we are

The planned forward Bengali implementation is ready for review: generalized
Brahmic infrastructure, Bengali rules, learned vowel model, optional spelling
lexicon and native reranker, evaluation tooling, CLI, npm/WASM and browser
selection. This is **experimental Bengali**, not a claim of Hindi-level quality.

All 16 PRs (#111–#126) were verified open and draft, with auto-merge disabled.
Every parent branch is an ancestor of its child. Local feature branches were
pushed and synchronized at handover. No main merge has occurred.

Snapshot before this documentation-only addition:

- Main, local and remote: `c683292abef3ee20a0d62b8aa617e6f2517a8f52` (#110).
- Implementation stack tip: `f6f9e657043c47da1c9f47db937301d2049ecedb`
  on `feature/bengali-web-npm` (#126).
- This note is subsequently committed/pushed on that top branch. Head hashes in
  the table are the implementation snapshot, not a promise that refs never move.
- Workspace: `/System/Volumes/Data/data/byte/code/gomanize`.
- The workspace is switched to `main` for handoff, with an identical **untracked
  copy of this note** at the repository root. No commit is made directly to main.
  Move that local copy to `/tmp` before checking out a branch that tracks it;
  otherwise Git may refuse to overwrite the untracked file.
- On main, most Bengali source/docs/tasks below do not exist yet. Read them from
  the PR branch or use `git show origin/feature/bengali-web-npm:PATH`.
  Main's old TASKS.md is not the completed stack's task state.

Start with `CLAUDE.md`, `docs/PROCESS.md`, then this note and the individual PR
bodies. Update TASKS.md **only through `./tools/tasks`**, never manually. Implement
review fixes on feature branches, not directly on main.

## 2. Exact stack and review order

Each row targets the branch in the preceding row, except #111 targets main.
Review each PR against its **actual base**, not the entire accumulated diff from
main. The cumulative implementation spans 180 files; many are frozen data and
reports. Each incremental PR is the useful review unit.

| Order / PR | Head branch | Current base | Snapshot head | Main review focus |
|---|---|---|---|---|
| [111](https://github.com/budhash/gomanize/pull/111) | `feature/bengali-design` | `main` | `85fadb9` | Design v4, scope, defaults, isolation/evaluation contracts |
| [112](https://github.com/budhash/gomanize/pull/112) | `feature/brahmic-profiles` | `feature/bengali-design` | `90b5ee5` | Shared ScriptProfile/retained vowels; frozen Hindi compatibility |
| [113](https://github.com/budhash/gomanize/pull/113) | `feature/bengali-b0` | `feature/brahmic-profiles` | `963a646` | Bengali symbols/parser configuration, API registration, pinned baseline fixtures |
| [114](https://github.com/budhash/gomanize/pull/114) | `feature/bengali-b1-prototype` | `feature/bengali-b0` | `6a82b6a` | Gemination/phala prototypes, hao/howa, khanda-ta, metadata-only debug traces |
| [115](https://github.com/budhash/gomanize/pull/115) | `feature/bengali-evaluation-gate` | `feature/bengali-b1-prototype` | `3ee9ce2` | Default-o and curation evidence; preregistered B1 dev gate |
| [116](https://github.com/budhash/gomanize/pull/116) | `feature/bengali-b1-rules` | `feature/bengali-evaluation-gate` | `83d2bae` | B1 rule behavior, gate improvement, external sources and sentence evaluation |
| [117](https://github.com/budhash/gomanize/pull/117) | `feature/bengali-b2-data` | `feature/bengali-b1-rules` | `e38db9e` | Frozen normalized training partitions and cross-source exclusions |
| [118](https://github.com/budhash/gomanize/pull/118) | `feature/bengali-b2-vowels` | `feature/bengali-b2-data` | `0af5994` | Train-only alignment and three-class absent/O/o vowel model |
| [119](https://github.com/budhash/gomanize/pull/119) | `feature/bengali-b2-lexicon` | `feature/bengali-b2-vowels` | `48d2235` | 8,980 attested spelling entries, Unicode keys, style fallback and exclusion guards |
| [120](https://github.com/budhash/gomanize/pull/120) | `feature/bengali-b2-rerank` | `feature/bengali-b2-lexicon` | `3f8aa0b` | **Rejected** character-only reranker experiment; no runtime promotion |
| [121](https://github.com/budhash/gomanize/pull/121) | `feature/bengali-b2-native-selector` | `feature/bengali-b2-rerank` | `b0a0130` | **Rejected** initial native-conditioned selector; dev-only decision |
| [122](https://github.com/budhash/gomanize/pull/122) | `feature/bengali-b2-crossfit` | `feature/bengali-b2-native-selector` | `d35a603` | Expanded candidates, out-of-fold training, fixed dev selection, held-out validation |
| [123](https://github.com/budhash/gomanize/pull/123) | `feature/bengali-b2-runtime-rerank` | `feature/bengali-b2-crossfit` | `8701a9b` | Opt-in NativeReranker contract, offline/runtime parity, style fallbacks, external results |
| [124](https://github.com/budhash/gomanize/pull/124) | `feature/bengali-b3-lyrics-seed` | `feature/bengali-b2-runtime-rerank` | `1465d6a` | Source-pinned lyrics pilot, **unreviewed references**, scoring and exposure audit |
| [125](https://github.com/budhash/gomanize/pull/125) | `feature/bengali-cli` | `feature/bengali-b3-lyrics-seed` | `3e3665e` | Language flag through every CLI path; Hindi default compatibility |
| [126](https://github.com/budhash/gomanize/pull/126) | `feature/bengali-web-npm` | `feature/bengali-cli` | `f6f9e65` | JS/WASM language dispatch, browser controls, option isolation, types and notices |

#120/#121 are intentional negative research results, not unfinished runtime
features. Later tooling builds on their work. Do not drop them from the stack
merely because their candidate models were rejected. Do not promote the rejected
models during cleanup. The successful replacement is #122/#123.

## 3. What each major layer now does

### Shared engine and Hindi compatibility

`script/brahmic` takes a script profile instead of hard-coding all Devanagari
assumptions. Bengali owns its symbol mappings and language rules; the colloquial
scheme is reused. Hindi's existing behavior is protected by a committed snapshot
of **3,432,832 exact outputs: 264,064 inputs × 13 profiles**. Aggregate accuracy
alone is insufficient to approve a refactor: different wrong outputs could
cancel each other numerically.

Review parser/renderer ownership, source runes/positions and unit metadata;
canonical Unicode equivalence; option/rule propagation; equal-priority checks;
and debug versus non-debug output. In particular, renderer behavior for
`SchwaPending` must remain compatible: `GetSchwa(unit) != SchwaDelete` includes
pending units. Replacing it with an explicit Keep-only check is not equivalent.

The separate Hindi second-unit `cccc`-final index correction is **not included**.
It is T-0058, deliberately separated from the structural refactor because it
would change Hindi output. Do not slip it into a conflict resolution or refresh
the frozen snapshot to conceal differences.

### Bengali rules and learned components

Rules handle scoped conjunct/phala pronunciation, gemination, final-cluster
vowels, anusvara/visarga, khanda-ta and হও/হওয়া contexts. Default inherent vowel
style is `o`; the model predicts absence versus retained O/o quality. Learned
components are opt-in; selecting Bengali does not automatically enable them.

The spelling lexicon is constructed from isolated Dakshina training spellings,
with normalized keys, sufficient attestation and a unique winning spelling.
It contains 8,980 entries. It is not a source of held-out answers.

The successful native reranker needs **both `SchwaModel` and `Rerank`**. It uses
the frozen cross-fitted selector at threshold 0.6 and considers up to eight
single-slot flips. Ties preserve earlier candidate order; ranking is deterministic.
Lexicon hits take precedence. Unsupported inputs and malformed-artifact paths
fall back safely. `InherentVowelA`, `LongVowels`, `SimpleNasals` and
`KeepMedialSchwa` bypass Bengali lexicon/reranking under the existing policy.

Core's optional native-reranker provider takes precedence over the legacy Hindi
path, including provider decline. Candidate rendering must retain the caller's
scheme and rule enable/disable selections, while preventing recursive reranking,
repeat lexicon application and candidate debug accumulation. Candidate closures
must not share mutable flip state. Review concurrency, declined-candidate debug
behavior and nil debug handling.

The runtime selector is byte-identical to the accepted experiment artifact:
`37215f8f26586d831e32d6c2cb84cca80526a8d1ed6edfc23180f6d2b8f7d5d2`.
See [runtime report](docs/reviews/2026-10-05-bengali-b2-runtime.md),
[selector contract](lang/bengali/RERANKER.md) and
[lexicon contract](lang/bengali/LEXICON.md).

### Frontends

- Go: `gomanize.New("bengali")`, or `NewWithOptions` for learned components.
- CLI: `--language=bengali` or `--language bengali`. Omission keeps Hindi.
  Positional input, stdin, input files, TSV tests, debug, rule listing and rule
  overrides all use the selected engine. Empty/missing/unsupported names fail.
- npm ESM/CJS: `g.translit(text, { language: "bengali" })`. Default language is
  Hindi **on each call**, not whichever language was used most recently. Full
  names are case-insensitive at runtime; TypeScript exposes lowercase names.
  Unsupported language values throw without terminating WASM. Wrappers reject
  unsupported JS types before crossing syscall/js, including BigInt/Symbol.
- WASM owns one engine per language and resets options on every invocation.
  A Bengali call must not leak flags or language into the next Hindi call.
- Browser: explicit selector, Bengali sample and labels, rerender on selection,
  input preserved across switches, options remembered separately per language
  for the page session. Hindi keeps its old Long vowels checkbox default;
  Bengali starts with all flags off.
- **Hindi style toggles are not Bengali rendering styles.** They are hidden when
  Bengali is selected; API callers supplying them retain the documented
  lexicon/reranker bypass behavior. The Bengali UI explains that reranking needs
  the vowel model. Do not present these Hindi toggles as completed Bengali style
  features or enable models silently.

See [frontend review record](docs/reviews/2026-10-05-bengali-frontends.md).
Package version is unchanged. No npm release or website deployment was performed.

## 4. Quality evidence — keep the denominators straight

These are final recorded results, not targets to tune against during review.

| Evaluation / profile | Result | Interpretation |
|---|---|---|
| Hindi curated, pure | 1,147/1,330 strict (86.2%); 1,236/1,330 match-any (92.9%) | Unchanged compatibility/accuracy reference |
| Bengali held-out Dakshina B1 | 1,413/2,500 match-any (56.52%) | Rules-only, no lexicon |
| Bengali held-out vowel model | 1,569/2,500 match-any (62.76%); 914 strict | Learned model, no lexicon |
| Bengali model + native reranker | 1,576/2,500 match-any (63.04%); 916 strict | 21 changed words, 8 match-any wins / 1 loss; minCER 0.07572118 |
| Bengali lexicon on held-out vocabularies | Zero coverage | Isolation guard, not a failure to improve those excluded words |
| Aksharantar pure model → reranker | 2,200 → 2,205 matches / 7,134 types | Small gain; overlaps Dakshina test, not independent evidence |
| BanglaTLit pure model → reranker | 12 → 12 exact / 2,500 sentences; macro CER 0.31286250 → 0.31286312 | Slight CER regression; do not hide it or call all external results improvements |
| Lyrics pilot pure B1 | 43/82 exact; macro line CER 0.021638 | Agreement with one assistant draft per line |
| Lyrics pilot pure model + reranker | 52/82 exact; macro line CER 0.017013 | **Provisional**, not independently attested accuracy |
| Lyrics pilot model + reranker + lexicon | 46/82 exact; macro line CER 0.025234 | Lexicon reduces agreement on this sample |

Hindi/Bengali datasets, curation and reference conventions differ. Do not compare
these percentages as if they used one common difficulty distribution. The good
lyrics pilot numbers do not establish Hindi parity. Report pure results first;
lexicon-assisted results separately. Retain both wins and losses, CER, strict
versus match-any definitions and full/unseen slices.

## 5. Data isolation, licensing and reference traps

Read [frozen training policy](training/data/bengali/README.md),
[source audit](docs/reference/bengali-repositories.md),
[B1 gate](docs/reviews/2026-10-04-bengali-evaluation-gate.md), and
[research record](docs/RESEARCH.md).

- Native keys normalize Unicode (Cf removal + NFC), including canonical nukta
  aliases. Exclude the union of Google and Dakshina dev/test native types from
  **both** training pools before fitting models or constructing lexicons.
- Frozen Google train/dev/test: 55,121 / 3,184 / 3,023 types. Frozen Dakshina train:
  23,117 types; dev and test: 2,500 each. Guards pin keys and semantic exclusions,
  not just editable manifest hashes. Changing a checksum is not an acceptable
  fix for an isolation failure.
- Dev selects models under the recorded protocol. Test and external evaluation
  sources must not become rule-tuning, threshold-selection or exception-mining
  inputs. Repeated review runs should verify behavior, not optimize their scores.
- The B1 gate is preregistered in #115, before B1 production changes in #116.
  Its enforcement intentionally fails at the preregistration stage. Use targets
  appropriate to the PR; do not rewrite earlier evidence to make every future
  target pass retroactively.
- **BanglaTLit's large upstream `train` CSV contains every official dev/test pair.**
  Only the pinned official test fixture is imported here, for evaluation. Do not
  feed that large pool into a trainer on the strength of its filename.
- Aksharantar includes Dakshina test words. External-source status does not mean
  no training-word overlap. Keep the explicit overlap audits and unseen slices.
- The user supplied bntranslit, bengali-romanizer, BanglaTLit and Romabangla as
  viable references. Their pinned revisions and useful roles are documented.
  bengali-romanizer's pronunciation/lexicon plumbing remains useful despite a
  different output style. No upstream runtime, pretrained network, or GPL code
  was copied. Romabangla remains a conceptual/separate-comparator reference.
- Code is MIT; data is not all MIT. Google pronunciation data/model attribution
  is CC BY 4.0; Dakshina-derived spelling data and the accepted Bengali selector
  are CC BY-SA 4.0. BanglaTLit test attribution is retained. `npm/NOTICE.md`
  carries code/data notices and the MIT text, is packaged in npm, and is copied
  to `web/NOTICE.md` by `make wasm`. Preserve this through release packaging.

### Lyrics review is a different review from PR approval

[The pilot](benchmark/data/bengali_lyrics/README.md) contains the first four
complete songs in the 1913 Gitanjali edition, selected in source order before
prediction: 82 occurrences / 71 unique native lines. It is a narrow Tagore seed,
not representative coverage of every Bengali lyrics style.

Original poems are identified as public domain by the source; the Wikisource
transcription/adapted dataset retains CC BY-SA 4.0 attribution. Source revisions
can transclude changing pages/templates: **committed HTML captures and hashes**
are the reproducibility anchor, not oldid URLs alone. Preserve native punctuation,
old spelling and the documented normalization rules.

All Roman references are explicitly `assistant_draft_unreviewed`, drafted from
native text and frozen before engine predictions. They are not human/native-
speaker attested, and the assistant's prior exposure to famous songs is unknown.
142/213 native token types overlap training vocabularies; only seven complete
lines are all-token-unseen. Scoring includes punctuation and spaces, and reports
all-line, unique-line, per-song and equal-song-weight metrics.

**T-0068 requires a fluent Bengali reviewer** to check pronunciation, poetic
forms, nasals, conventions and acceptable variants, record identity/date and
per-row corrections, and attest the reference set. Start from source + worksheet,
without consulting engine predictions. Preserve v1 evidence when correcting the
references. An LLM code review or another pass over my own drafts must not be
silently labeled independent human gold. T-0054 stays open pending that review
and the B3 coverage/completion decision. These tasks do not block reviewing the
code as experimental support, but block claiming validated lyrics gold.

## 6. Validation evidence and practical commands

Full local `make ci` runs were completed before publishing the implementation
PRs; detailed phase results live in PR bodies and dated review records. Recent
runs: #124 frozen root replay 689.700 seconds, #125 660.324 seconds in an isolated
checkout of `3e3665e`, #126 782.056 seconds. Both frontend PRs passed full CI.
The final task-link commits and this handover are documentation-only additions;
pre-commit hooks validate those, without repeating a 10–14 minute replay.

**Hosted CI gap:** `.github/workflows/ci.yml` runs on pushes and PRs targeting
main. Most stacked PRs target feature branches, so do not confuse absent hosted
full checks with a passed full run. #111's hosted CI was successful when checked.
Retargeting each next PR to main should trigger hosted validation; inspect the
actual run and commit rather than relying on an earlier green badge.

On the implementation branch (not pre-stack main):

```sh
export GOCACHE=/tmp/gomanize-review-go-cache
export GOLANGCI_LINT_CACHE=/tmp/gomanize-review-lint-cache
make ci
make npm-test
GOOS=js GOARCH=wasm go vet ./cmd/gomanize-wasm
./tools/tasks validate
```

`make ci` includes format, lint, build, race/coverage, accuracy and Bengali guards.
`make npm-test` builds CLI and WASM, then checks actual ESM and CJS execution:
nine profiles per language against CLI, defaults after Bengali calls, option
reset, invalid-value recovery and mixed-script/whitespace behavior. WASM glue is
excluded from ordinary native Go builds, so **make ci alone is insufficient for
#126**. Strict TypeScript checking and npm pack dry-run were also verified.
Local Safari verification covered switching, samples, text preservation,
Hindi defaults, remembered Bengali options and hidden Hindi-only controls.

Useful targeted checks at the final tip:

```sh
go test ./cmd/gomanize -count=1
make test-bengali
make test-bengali-b1-gate
make test-bengali-data
make test-bengali-vowels
make test-bengali-lexicon
make test-bengali-rerank
make test-bengali-native-selector
make test-bengali-crossfit
make test-bengali-runtime
make test-bengali-lyrics
python3 tools/bengali/evaluate_lyrics.py --output /tmp/bengali-lyrics-review.json
```

Do not rewrite committed reference/report artifacts automatically when a test
fails. Inspect the cause and preserve independent evidence. Negative controls
already exercised include corrupted data/rehashed semantic violations, restoring
the CLI's hard-coded Hindi constructor, and ignoring the WASM language option;
those controls failed as intended, then restored implementations passed.

Operational notes:

- Race/atomic-coverage replay can take 10–14 minutes with little new output.
  Current Go test timeout is 20 minutes. Check process activity before declaring
  a hang; do not repeatedly restart the full suite. Extra worktrees can miss Go
  caches because paths differ. Use focused checks while editing and full CI on
  the final reviewed change.
- Prior disk exhaustion was resolved by the user; no personal files were deleted.
  Check `df -h .` before large regenerations. Frozen outputs/data can be large.
- Sandbox may deny default Go caches/module metadata writes. The `/tmp` caches
  above worked. Some builds printed a nonfatal `/apps/go/pkg/mod/...` stat-cache
  warning yet exited successfully. Distinguish that from a failing command.
- `.gitignore` has an unanchored `gomanize` entry, which can hide **new files under
  cmd/gomanize/** from `rg --files` and `git add`. Existing tracked files still
  exist. #125's new test was explicitly force-added; check `git ls-files` when
  reviewing that directory, and use narrowly scoped `git add -f` if necessary.
- Make regex end anchors need `$$` (Make escaping); a single `$` can break shell
  quoting. Existing new targets already account for this.
- Generated root CLI, web WASM/loader/notice, npm/dist, coverage and caches can
  survive checkout. **Rebuild after switching branches**; a binary left in the
  working tree does not prove main contains the feature. Do not commit binaries.
- Temporary raw sources under `/tmp` and local logs are conveniences, not durable
  dependencies. Some complete experiment reproductions require the pinned Google
  lexicon or Aksharantar download. Follow manifest URLs/hashes if the scratch
  files have gone. Frozen fixtures and normal CI should not depend on my temp
  clone paths. The temporary CLI-validation worktree and local preview server
  created for #125/#126 were removed/stopped.

## 7. Review and merge procedure — important gotchas

**Review all PRs; merge from #111 upward. Do not merge a child into its feature
parent by accidentally leaving its base unchanged. Do not merge the top PR as a
shortcut around individual reviews.**

Repository settings verified at handover: merge commits, squash and rebase merges
are all enabled; automatic branch deletion is false. For this existing linear
stack, **ordinary merge commits are the simplest way to preserve ancestry**.
Squash/rebase merging changes ancestor identities and requires deliberate
restacking of descendants; a base change alone may then replay old changes.

Suggested sequence for each PR:

1. Fetch and inspect the live PR base, head SHA, diff and checks. Record the
   reviewed SHA. Review the appropriate dated report, source, artifacts and
   tests; categorize findings as fixed, rejected with reason, or tracked deferral.
2. Make required fixes on that feature branch. Recompute relevant evidence,
   including pure accuracy before/after for a core behavior change. A regression
   test needs a sanity-revert proving it catches the defect. Do not refresh the
   Hindi snapshot or tune on held-out references just to make CI green.
3. After its parent has landed, retarget this PR to main explicitly and integrate
   updated main into the child as needed. Parent review fixes made after the child
   fork are **not automatically in the child's working tree**. With merge-commit
   ancestry, a normal merge from updated main into the child propagates them.
   Resolve conflicts by retaining intended child changes and reviewed parent fixes.
4. Inspect the new main-targeting diff and rerun required tests on the actual new
   head; #126 needs npm/WASM tests too. Review any newly exposed conflict changes.
5. Mark ready and merge only as part of the user's review/merge process, then
   fetch the resulting main state and continue with the next child. Keep feature
   parents until descendants have been retargeted and their bases verified.
6. After all intended PRs land, verify main with full CI + npm tests, check task
   statuses/docs/versioning and distinguish merged experimental support from a
   tagged package release and from human-attested lyrics quality.

Do not mass-force-push the stack. If squash/rebase merges are chosen, save the
old parent boundaries and plan descendant `rebase --onto` operations carefully,
one layer at a time, rather than blindly rebasing every branch onto main.

Inspection and remote-ref refresh commands:

```sh
git status --short --branch
git fetch origin
gh pr list --state open --limit 100 --json number,title,baseRefName,headRefName,isDraft
gh pr view 111 --json baseRefName,headRefName,headRefOid,isDraft,statusCheckRollup
gh pr diff 111
git diff origin/feature/bengali-design...origin/feature/brahmic-profiles
```

### Merging main also deploys the website

`.github/workflows/pages.yml` deploys on **every push to main**, independently of
the main CI workflow, as well as manual dispatch. Thus staged merges can publish
intermediate web builds; merging #126 exposes the Bengali selector on the public
site. Surface this to the user before the first merge if they expect review-only
or deferred deployment. Do not assume a merge is isolated from deployment.

Go releases trigger on `v*.*.*` tags; npm publishing triggers on `v*` tags or
manual workflow dispatch. Do not create release tags or dispatch publishing as
an incidental part of PR merging. npm version remains 1.2.1; release preparation
must align package/version metadata with the intended tag before publishing.

## 8. Remaining work and the expected outcome

| Item | Status / expectation |
|---|---|
| Core + frontend implementation | Implemented and tested on the stack; awaiting this independent PR review |
| T-0067 lyrics pilot | Done; source/measurement infrastructure, not gold |
| T-0068 reference attestation | Open; fluent Bengali reference review required |
| T-0054 B3 gold completion | Open; depends on pilot + attestation and coverage assessment |
| T-0058 Hindi index correction | Open separate investigation; not part of this structural compatibility change |
| F-0012 / T-0059 Roman → Hindi/Bengali | Future design/candidate-ranking work; no reverse API implemented |
| Release | Not performed; preserve experimental designation and explicit opt-ins |

Some earlier design/source-audit reports describe their historical stage (for
example, B2 not yet implemented). Do not interpret that historical text as the
current implementation state. Read the later phase reports, top-branch TASKS.md
and this handover together. The original design mentions Gitabitan and wider
poet coverage; the implemented pilot is explicitly the first four songs from a
1913 Gitanjali edition, and the overall gold task remains open.

Claude's expected deliverable is an independently reviewed stack, explicit
resolution of findings, validated parent-to-child propagation, and a controlled
bottom-up merge record. Preserve provenance, rejected experiments, Hindi
compatibility and honest Bengali quality labels throughout. Code review can
approve experimental implementation; it cannot manufacture missing human
reference attestation.
