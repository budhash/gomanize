# Bengali CLI, npm/WASM and browser integration

T-0069 exposes the existing Bengali engine through CLI language selection;
T-0070 adds the same choice to npm/WASM and the browser. Neither increment
changes the core rules, learned artifacts or quality thresholds. Bengali
remains experimental; the independent lyrics-reference review is still open.

## Contract

- CLI: `--language=bengali` and `--language bengali`; omitting the option selects
  Hindi. Selection applies to every input/debug/test/rule mode. Full language
  names are case-insensitive. Missing, empty and unsupported values fail.
- JS: `translit(text, {language: "bengali"})`; each call independently defaults
  to Hindi when language is absent. Both ESM and CJS forward the option. Invalid
  values throw an Error without terminating the WASM runtime. The TypeScript
  `Language` union exposes the canonical lowercase names.
- Browser: a labeled selector updates input language, label, placeholder, sample
  and rendered output. Existing text survives language changes. Checkbox values
  are remembered separately per language for the current page session. Hindi
  retains its previous Long vowels default; Bengali starts with all flags off.
- Bengali reranking requires the vowel model. Hindi style toggles are hidden
  for Bengali: they do not implement Bengali rendering styles, and passing them
  through the API bypasses Bengali lexicon/reranking under the existing contract.
- npm and web assets include code/embedded-data attribution via `NOTICE.md`.
  The package version is unchanged and no package or website is published.

## Validation

CLI subprocess tests run the actual entry point and compare both languages with
the library. Coverage includes positional and piped input, input files, TSV
comparison, debug, rule listing/overrides, omitted language and malformed flags.
A sanity-revert to the hard-coded Hindi constructor fails the Bengali tests.
Full local `make ci` passed for CLI commit `3e3665e` in an isolated checkout.

`make npm-test` now builds the CLI as well as WASM. Both ESM and CJS test nine
option profiles for each language against the CLI, plus defaults after Bengali
calls, option reset, mixed scripts, whitespace/punctuation and recovery after
invalid languages. A sanity-revert that ignores the WASM language option fails
its Bengali assertion; the restored implementation passes. TypeScript strict
checking accepts both languages and rejects an unsupported one. An npm pack
dry run confirms the entry points, declarations, WASM, runtime and notice ship.

Local Safari inspection confirms initial Hindi selection, successful Bengali
sample rendering, text preservation during switching, restored Hindi defaults
and remembered Bengali learned options. A final page check confirms that only
supported learned controls appear for Bengali. The second full local `make ci`
result is recorded in the npm/browser PR. Hosted full CI remains configured for
main-targeting PRs, so these stacked-branch full runs are local.

## Review notes (2026-10-05)

Independent stack review; language/option isolation, ESM/CJS parity, textContent
output and per-language UI state verified.

- **Flag types are validated in JS.** Only `language` was type-checked, so a
  BigInt flag (`{longVowels: 1n}`) panicked inside the Go runtime and every later
  call failed with "Go program has already exited"; string flags were truthy
  (`"false"` meant true). Both wrappers now reject non-object options and
  non-boolean flags with `TypeError`; the smoke test asserts each case and that
  the instance survives.
- **Licensing.** `NOTICE.md` lists every embedded data file (Hindi lexicon, schwa
  tree, n-grams; Bengali lexicon, vowel tree, selector) with source, license and
  changes. `package.json` declares `MIT AND CC-BY-SA-4.0 AND CC-BY-4.0`. GoReleaser
  archives now include `NOTICE.md`; the npm README links it absolutely.
- **Web.** The language select opts out of browser form restoration so a restored
  value cannot desynchronize the UI; structured data lists `bn`.
- Tracked: unknown option keys are ignored; npm Bengali `rerank` without
  `schwaModel` is silent (the CLI warns); the About dialog still describes the
  Hindi components only.
