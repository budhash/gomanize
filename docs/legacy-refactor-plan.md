> **HISTORICAL DOCUMENT (the completed 2025 core-refactor task plan; all phases shipped. The live tracker is /TASKS.md).** Statistics and plans below reflect the
> project as of 2025 and are NOT current. For today's status see `CLAUDE.md`;
> for the live backlog see `/TASKS.md` (via `./tools/tasks tree`); for every
> subsequent result and decision see `docs/reviews/`.

# Architecture Refactoring Tasks

This document contains detailed tasks for refactoring gomanize to the new layered architecture.

## Current State

- `engine/` - Contains Brahmic-specific code mixed with universal mechanics
- `script/brahmic/` - New brahmic layer (proof of concept, working)
- `lang/hindi/` - Uses old `engine/` package
- `lang/hindi2/` - Uses new `script/brahmic/` package (proof of concept, working)
- No `core/` package yet
- No `scheme/` packages yet

## Phase 1: Create Core Package

**Goal**: Extract universal mechanics into `core/` package.

### Task 1.1: Create `core/types.go`
- [ ] Define `Position` struct
- [ ] Define `UnitType` enum (UnitVowel, UnitConsonant, etc.)
- [ ] Define `Unit` struct with `ScriptData interface{}` extension point
- [ ] Define `Word` struct
- [ ] Define `Options` struct
- [ ] Add helper methods (`Unit.IsWordFinal()`, `Unit.IsWordInitial()`, etc.)

### Task 1.2: Create `core/categories.go`
- [ ] Define base `Category` type
- [ ] Define `CatUnknown`, `CatVowel`, `CatConsonant`, `CatNumber`, `CatSymbol`
- [ ] Define `SymbolInfo` struct
- [ ] Define `SymbolMap` type

### Task 1.3: Create `core/rule.go`
- [ ] Define `RulePhase` enum (PhaseSchwa, PhaseConsonant, PhaseVowel, PhaseRender)
- [ ] Define `RuleScope` enum (ScopeUniversal, ScopeScript, ScopeLanguage, ScopeScheme)
- [ ] Define `RuleMode` enum (ModeExclusive, ModeAlways, ModeFallback)
- [ ] Define `Rule` struct
- [ ] Define `RuleEngine` struct
- [ ] Implement `NewRuleEngine(rules []Rule)`
- [ ] Implement `RuleEngine.Apply(word *Word)`
- [ ] Implement `RuleEngine.AddRule()` with priority conflict detection

### Task 1.4: Create `core/catalog.go`
- [ ] Define `RuleCatalog` struct with Schwa, Consonant, Vowel, Render slices
- [ ] Implement `RuleCatalog.AllRules() []Rule`
- [ ] Implement `RuleCatalog.FindByName(name string) *Rule`
- [ ] Implement helper `appendIfFound(rules []Rule, catalog []Rule, name string) []Rule`

### Task 1.5: Create `core/interfaces.go`
- [ ] Define `Parser` interface
- [ ] Define `Renderer` interface
- [ ] Define `Script` interface
- [ ] Define `Language` interface
- [ ] Define `Scheme` interface

### Task 1.6: Create `core/engine.go`
- [ ] Define `Engine` struct
- [ ] Implement `NewEngine(lang Language, scheme Scheme) *Engine`
- [ ] Implement `Engine.Transliterate(input string) string`
- [ ] Implement `Engine.TransliterateWithOptions(input string, opts Options) string`

### Task 1.7: Tests for core package
- [ ] Test RuleEngine priority sorting
- [ ] Test RuleEngine phase ordering
- [ ] Test RuleCatalog.FindByName
- [ ] Test Engine with mock Script/Language/Scheme

## Phase 2: Refactor script/brahmic to use core

**Goal**: Update `script/brahmic/` to implement `core.Script` interface.

### Task 2.1: Update `script/brahmic/types.go`
- [ ] Import `core` package
- [ ] Keep `SchwaState`, `ConsonantRun` (Brahmic-specific)
- [ ] Define `BrahmicData` struct for `Unit.ScriptData`
- [ ] Define `Config` struct with Halant, Nukta fields
- [ ] Add helper to get BrahmicData from Unit: `GetBrahmicData(u *core.Unit) *BrahmicData`

### Task 2.2: Update `script/brahmic/categories.go`
- [ ] Import `core` package
- [ ] Define Brahmic categories starting at 100+ to avoid collision
- [ ] Update `SymbolInfo` references to use `core.SymbolInfo`

### Task 2.3: Update `script/brahmic/parser.go`
- [ ] Change Parser to implement `core.Parser` interface
- [ ] Use `core.Unit` with `ScriptData` for BrahmicData
- [ ] Use `core.Word`
- [ ] Accept `core.SymbolMap` in Parse method

### Task 2.4: Update `script/brahmic/renderer.go`
- [ ] Change Renderer to implement `core.Renderer` interface
- [ ] Use `core.Unit` and `core.Word`
- [ ] Get SchwaState from `Unit.ScriptData`

### Task 2.5: Update `script/brahmic/runs.go`
- [ ] Use `core.Unit` and `core.Word`
- [ ] Store ConsonantRun in `Unit.ScriptData`

### Task 2.6: Create `script/brahmic/brahmic.go`
- [ ] Implement `Brahmic` struct
- [ ] Implement `Name() string`
- [ ] Implement `NewParser(config interface{}) core.Parser`
- [ ] Implement `NewRenderer() core.Renderer`
- [ ] Implement `PrepareWord(word *core.Word)`
- [ ] Implement `Categories() []core.Category`

### Task 2.7: Tests for refactored brahmic
- [ ] Verify Parser produces correct Units with ScriptData
- [ ] Verify Renderer reads SchwaState from ScriptData
- [ ] Verify IdentifyRuns populates ConsonantRun in ScriptData

## Phase 3: Refactor lang/hindi to use core

**Goal**: Update Hindi to implement `core.Language` with `RuleCatalog`.

### Task 3.1: Update `lang/hindi/symbols.go`
- [ ] Use `core.SymbolMap` and `core.SymbolInfo`
- [ ] Reference `brahmic.CatHalant`, etc. for categories

### Task 3.2: Create `lang/hindi/rules.go` with RuleCatalog
- [ ] Define individual rules as named variables (for scheme reference)
- [ ] Implement `rules() core.RuleCatalog`
- [ ] Organize rules into Schwa, Consonant, Vowel, Render categories
- [ ] Include alternative rules (e.g., `SchwaKeepAll` for scholarly schemes)
- [ ] Rules use `core.Unit`, `core.Word`
- [ ] Rules access BrahmicData via `Unit.ScriptData`

### Task 3.3: Update `lang/hindi/hindi.go`
- [ ] Implement `core.Language` interface
- [ ] `Name() string` returns "hindi"
- [ ] `Script() core.Script` returns `brahmic.Brahmic{}`
- [ ] `Symbols() core.SymbolMap` returns symbol map
- [ ] `ScriptConfig() interface{}` returns `brahmic.Config{Halant: "्", Nukta: "़"}`
- [ ] `Rules() core.RuleCatalog` returns rule catalog

### Task 3.4: Tests for refactored hindi
- [ ] Verify Hindi implements core.Language
- [ ] Verify RuleCatalog contains expected rules
- [ ] Verify rules work with core.Unit/ScriptData

## Phase 4: Create scheme/colloquial

**Goal**: Implement colloquial scheme that selects rules from catalog.

### Task 4.1: Create `scheme/colloquial/colloquial.go`
- [ ] Define `Colloquial` struct
- [ ] Implement `Name() string` returns "colloquial"
- [ ] Implement `SelectRules(catalog core.RuleCatalog) []core.Rule`
- [ ] Select schwa deletion rules
- [ ] Select va-to-wa rule
- [ ] Select long-aa-closed-final rule
- [ ] Skip render transform rules

### Task 4.2: Tests for colloquial scheme
- [ ] Verify SelectRules returns expected rules
- [ ] Verify missing rules are skipped gracefully

## Phase 5: Integration and Testing

**Goal**: Verify new architecture works end-to-end.

### Task 5.1: Create integration test
- [ ] Create engine with hindi.Hindi{} and colloquial.Colloquial{}
- [ ] Test basic transliteration
- [ ] Test schwa deletion
- [ ] Test conjuncts
- [ ] Test long vowels option

### Task 5.2: Parity test
- [ ] Compare output of new architecture vs old engine
- [ ] Run against Dakshina dataset
- [ ] Verify 82.5%+ accuracy maintained

### Task 5.3: Update gomanize.go public API
- [ ] Update `New(lang string)` to use new architecture
- [ ] Default to colloquial scheme
- [ ] Support scheme selection (future)

## Phase 6: Cleanup

**Goal**: Remove old code and update documentation.

### Task 6.1: Remove old packages
- [ ] Remove `engine/` package (after confirming all tests pass)
- [ ] Remove `lang/hindi2/` (merge into `lang/hindi/`)
- [ ] Remove old `script/brahmic/` types that moved to core

### Task 6.2: Update documentation
- [ ] Update CLAUDE.md with new architecture
- [ ] Remove `docs/SCRIPT_LAYER_ARCHITECTURE.md` (superseded by SPEC.md)
- [ ] Update README.md if needed

### Task 6.3: Final verification
- [ ] Run `make ci`
- [ ] Run `make test-dakshina`
- [ ] Verify accuracy threshold (82%+)

## Phase 7: Future Work (not in this refactor)

### Schemes
- [ ] Implement `scheme/iast/` for scholarly output
- [ ] Implement `scheme/hunterian/` for official India standard

### Languages
- [ ] Implement `lang/marathi/` (different schwa rules)
- [ ] Implement `lang/tamil/` (no schwa deletion)

### Scripts
- [ ] Design `script/arabic/` for future RTL support

## Task Dependencies

```
Phase 1 (core/)
    │
    ├──► Phase 2 (script/brahmic uses core)
    │        │
    │        └──► Phase 3 (lang/hindi uses core + brahmic)
    │                 │
    │                 └──► Phase 4 (scheme/colloquial)
    │                          │
    │                          └──► Phase 5 (integration)
    │                                   │
    │                                   └──► Phase 6 (cleanup)
    │
    └──► Can be developed in parallel with Phase 2
```

## Estimated Complexity

| Phase | Tasks | Complexity | Notes |
|-------|-------|------------|-------|
| 1 | 7 | Medium | New code, well-defined interfaces |
| 2 | 7 | Medium | Refactoring existing code |
| 3 | 4 | Low | Mostly reorganization |
| 4 | 2 | Low | Small, focused |
| 5 | 3 | Medium | Integration testing |
| 6 | 3 | Low | Cleanup |

## Definition of Done

- [ ] All tasks checked off
- [ ] `make ci` passes
- [ ] Dakshina accuracy ≥ 82%
- [ ] No Brahmic-specific code in `core/`
- [ ] Clean imports: core ← script ← lang ← scheme
- [ ] Documentation updated
