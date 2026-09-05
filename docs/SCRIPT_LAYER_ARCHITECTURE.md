> **HISTORICAL DOCUMENT (design doc for the script layer, superseded by SPEC.md and the shipped code).** Statistics and plans below reflect the
> project as of 2025 and are NOT current. For today's status see `CLAUDE.md`;
> for the live backlog see `/TASKS.md` (via `./tools/tasks tree`); for every
> subsequent result and decision see `docs/reviews/`.

# Script Layer Architecture

## Overview

This document describes the layered architecture for transliteration, designed to scale to 50+ languages and multiple output schemes while keeping each component maintainable.

## Design Principles

1. **Separation of Concerns**: Each layer has a single responsibility
2. **Rules are Agnostic**: Rules don't know which schemes use them
3. **Schemes Select, Don't Define**: Schemes choose from language's rule catalog
4. **Scalability**: Adding a language or scheme is mechanical, not architectural

## Architecture Layers

```
┌─────────────────────────────────────────────────────────────────────┐
│                           CORE (core/)                              │
│                     Universal Mechanics                             │
├─────────────────────────────────────────────────────────────────────┤
│  Types:      Unit, Word, Position, Category, UnitType               │
│  Mechanics:  Rule, RuleEngine, RulePhase, RuleScope, RuleMode       │
│  Interfaces: Script, Language, Scheme, Parser, Renderer             │
│  Engine:     Orchestrates the pipeline                              │
└─────────────────────────────────────────────────────────────────────┘
                               │
         ┌─────────────────────┼─────────────────────┐
         ▼                     ▼                     ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│ SCRIPT          │  │ SCRIPT          │  │ SCRIPT          │
│ script/brahmic/ │  │ script/arabic/  │  │ script/latin/   │
├─────────────────┤  │ (future)        │  │ (future)        │
│ Parser          │  └─────────────────┘  └─────────────────┘
│ Renderer        │
│ PrepareWord     │
│ Script-specific │
│ types & cats    │
└─────────────────┘
         │
    ┌────┴────────┬──────────┬──────────┐
    ▼             ▼          ▼          ▼
┌───────┐    ┌───────┐  ┌───────┐  ┌───────┐
│LANGUAGE    │LANGUAGE│  │LANGUAGE│  │LANGUAGE│
│lang/hindi/ │marathi/│  │ tamil/ │  │bengali/│
├───────┤    └───────┘  └───────┘  └───────┘
│Symbols│
│Config │
│Rules  │ ← RuleCatalog (all rules for this language)
└───────┘
    │
    └──────────────┬──────────────┬──────────────┐
                   ▼              ▼              ▼
              ┌─────────┐   ┌─────────┐   ┌─────────┐
              │ SCHEME  │   │ SCHEME  │   │ SCHEME  │
              │colloq.  │   │  iast   │   │hunter.  │
              ├─────────┤   └─────────┘   └─────────┘
              │SelectRules(catalog) → []Rule        │
              └─────────────────────────────────────┘
```

## Layer Responsibilities

### Core (`core/`)

Universal types and mechanics that work for ANY script.

```go
// core/types.go
type Unit struct {
    Runes      []rune
    Type       UnitType
    BaseRom    string
    Prev, Next *Unit
    ScriptData interface{}  // Extension point for script-specific data
}

type Word struct {
    Units    []*Unit
    Original string
    Options  Options
}

// core/rule.go
type Rule struct {
    Name      string
    Phase     RulePhase
    Scope     RuleScope
    Priority  int
    Mode      RuleMode
    Condition func(*Unit, *Word) bool
    Action    func(*Unit, *Word)
}

type RulePhase int
const (
    PhaseSchwa     RulePhase = iota  // Schwa decisions
    PhaseConsonant                    // Consonant modifications
    PhaseVowel                        // Vowel modifications
    PhaseRender                       // Output transforms (aa→ā)
)
```

### Script (`script/`)

Script-family-specific parsing and rendering.

```go
// core/interfaces.go
type Script interface {
    Name() string
    NewParser(config interface{}) Parser
    NewRenderer() Renderer
    PrepareWord(word *Word)  // e.g., IdentifyRuns for Brahmic
}

type Parser interface {
    Parse(input string, symbols SymbolMap) *Word
}

type Renderer interface {
    Render(word *Word) string
}
```

**Brahmic Script** (`script/brahmic/`):
- Handles halant, nukta, matras
- Tracks SchwaState, AfterHalant
- Identifies ConsonantRuns for schwa coordination

**Arabic Script** (future):
- Handles RTL, harakat (vowel points)
- Different categories and parsing logic

### Language (`lang/`)

Language-specific symbols and rule catalog.

```go
// core/interfaces.go
type Language interface {
    Name() string
    Script() Script
    Symbols() SymbolMap
    ScriptConfig() interface{}
    Rules() RuleCatalog  // ALL rules this language knows
}

// core/catalog.go
type RuleCatalog struct {
    Schwa     []Rule  // All schwa-related rules
    Consonant []Rule  // All consonant rules
    Vowel     []Rule  // All vowel rules
    Render    []Rule  // All output transform rules
}
```

Each language provides a **complete catalog** of all its rules, organized by category.

### Scheme (`scheme/`)

Selects which rules to use from a language's catalog.

```go
// core/interfaces.go
type Scheme interface {
    Name() string
    SelectRules(catalog RuleCatalog) []Rule
}
```

Schemes are **language-agnostic** - they select by rule name/category, not by language.

## How It Works

### Engine Creation

```go
func NewEngine(lang Language, scheme Scheme) *Engine {
    // Get language's complete rule catalog
    catalog := lang.Rules()

    // Scheme selects which rules to use
    selectedRules := scheme.SelectRules(catalog)

    return &Engine{
        script:     lang.Script(),
        symbols:    lang.Symbols(),
        config:     lang.ScriptConfig(),
        ruleEngine: NewRuleEngine(selectedRules),
        renderer:   lang.Script().NewRenderer(),
    }
}
```

### Execution Pipeline

```go
func (e *Engine) Transliterate(input string) string {
    // 1. Parse (Script handles script-specific parsing)
    parser := e.script.NewParser(e.config)
    word := parser.Parse(input, e.symbols)

    // 2. Prepare (Script-specific, e.g., IdentifyRuns)
    e.script.PrepareWord(word)

    // 3. Apply rules (selected by Scheme from Language catalog)
    e.ruleEngine.Apply(word)

    // 4. Render (Script handles script-specific rendering)
    return e.renderer.Render(word)
}
```

## Example: Hindi Language

```go
// lang/hindi/rules.go
func (h Hindi) Rules() core.RuleCatalog {
    return core.RuleCatalog{
        Schwa: []core.Rule{
            // Multiple options - schemes will choose
            {Name: "schwa-delete-word-final", Phase: PhaseSchwa, ...},
            {Name: "schwa-delete-ccv", Phase: PhaseSchwa, ...},
            {Name: "schwa-keep-internal-conjunct", Phase: PhaseSchwa, ...},
            {Name: "schwa-keep-all", Phase: PhaseSchwa, ...},  // For scholarly
        },
        Consonant: []core.Rule{
            {Name: "va-to-wa-conjunct", Phase: PhaseConsonant, ...},
            {Name: "va-keep-v", Phase: PhaseConsonant, ...},  // Alternative
        },
        Vowel: []core.Rule{
            {Name: "long-aa-closed-final", Phase: PhaseVowel, ...},
            {Name: "long-aa-all", Phase: PhaseVowel, ...},  // For --long-vowels
        },
        Render: []core.Rule{
            {Name: "render-aa-macron", Phase: PhaseRender, ...},  // aa→ā
            {Name: "render-sh-acute", Phase: PhaseRender, ...},   // sh→ś
        },
    }
}
```

## Example: Schemes

```go
// scheme/colloquial/colloquial.go
type Colloquial struct{}

func (c Colloquial) Name() string { return "colloquial" }

func (c Colloquial) SelectRules(catalog core.RuleCatalog) []core.Rule {
    var rules []core.Rule

    // Select colloquial schwa rules
    rules = append(rules, findByName(catalog.Schwa, "schwa-delete-word-final"))
    rules = append(rules, findByName(catalog.Schwa, "schwa-delete-ccv"))
    rules = append(rules, findByName(catalog.Schwa, "schwa-keep-internal-conjunct"))

    // Select consonant rules
    rules = append(rules, findByName(catalog.Consonant, "va-to-wa-conjunct"))

    // Select vowel rules
    rules = append(rules, findByName(catalog.Vowel, "long-aa-closed-final"))

    // No render transforms for colloquial (output as-is)

    return filterNil(rules)
}
```

```go
// scheme/iast/iast.go
type IAST struct{}

func (i IAST) Name() string { return "iast" }

func (i IAST) SelectRules(catalog core.RuleCatalog) []core.Rule {
    var rules []core.Rule

    // IAST keeps all schwas (scholarly)
    rules = append(rules, findByName(catalog.Schwa, "schwa-keep-all"))

    // No consonant changes
    rules = append(rules, findByName(catalog.Consonant, "va-keep-v"))

    // Render transforms for diacritics
    rules = append(rules, findByName(catalog.Render, "render-aa-macron"))
    rules = append(rules, findByName(catalog.Render, "render-sh-acute"))

    return filterNil(rules)
}
```

## Scaling Analysis

### Adding a New Language (e.g., Marathi)

1. Create `lang/marathi/symbols.go` - character mappings
2. Create `lang/marathi/rules.go` - rule catalog (Marathi-specific schwa rules)
3. Create `lang/marathi/marathi.go` - implement `Language` interface
4. **Done** - all existing schemes work automatically

### Adding a New Scheme (e.g., Hunterian)

1. Create `scheme/hunterian/hunterian.go`
2. Implement `SelectRules()` - choose rules by name
3. **Done** - works with all existing languages

### Adding a New Script (e.g., Arabic)

1. Create `script/arabic/parser.go` - RTL parsing, harakat
2. Create `script/arabic/renderer.go` - Arabic rendering
3. Create `script/arabic/types.go` - Arabic-specific data
4. Create `script/arabic/arabic.go` - implement `Script` interface
5. **Done** - can now add Arabic languages

## Directory Structure

```
gomanize/
├── core/                    # Universal mechanics
│   ├── types.go            # Unit, Word, Position
│   ├── rule.go             # Rule, RuleEngine
│   ├── catalog.go          # RuleCatalog
│   ├── interfaces.go       # Script, Language, Scheme
│   └── engine.go           # Orchestrator
│
├── script/                  # Script family implementations
│   └── brahmic/
│       ├── brahmic.go      # Script interface impl
│       ├── parser.go       # Brahmic parser
│       ├── renderer.go     # Brahmic renderer
│       ├── types.go        # SchwaState, ConsonantRun
│       ├── categories.go   # Brahmic categories
│       └── runs.go         # IdentifyRuns
│
├── lang/                    # Language implementations
│   ├── hindi/
│   │   ├── hindi.go        # Language interface impl
│   │   ├── symbols.go      # Symbol mappings
│   │   └── rules.go        # Rule catalog
│   ├── marathi/
│   └── tamil/
│
├── scheme/                  # Scheme implementations
│   ├── colloquial/
│   │   └── colloquial.go   # SelectRules for colloquial
│   ├── iast/
│   │   └── iast.go         # SelectRules for IAST
│   └── hunterian/
│       └── hunterian.go    # SelectRules for Hunterian
│
└── gomanize.go             # Public API
```

## Key Design Decisions

### 1. Rules are Agnostic
Rules don't know which schemes use them. They're building blocks.

### 2. Languages Own Rules
Each language defines ALL its possible rules in a RuleCatalog. This keeps language-specific logic together.

### 3. Schemes Select, Don't Define
Schemes choose from the language's catalog by rule name. This decouples schemes from languages.

### 4. Script Provides Extension Point
`Unit.ScriptData interface{}` allows scripts to attach their own data (SchwaState, etc.) without core knowing about it.

### 5. Missing Rules are OK
If a scheme requests a rule that doesn't exist in a language's catalog (e.g., Tamil has no "schwa-delete-word-final"), it's simply skipped.

## Migration Path

1. **Phase 1** (done): Create `script/brahmic/` and `lang/hindi2/` as proof of concept
2. **Phase 2**: Create `core/` package with interfaces
3. **Phase 3**: Refactor `script/brahmic/` to implement `core.Script`
4. **Phase 4**: Refactor `lang/hindi2/` to implement `core.Language` with `RuleCatalog`
5. **Phase 5**: Create `scheme/colloquial/` implementing `core.Scheme`
6. **Phase 6**: Update `gomanize.go` to use new architecture
7. **Phase 7**: Remove old `engine/` and `lang/hindi/`

## Success Criteria

1. All existing tests pass
2. Dakshina accuracy maintained (82.5%+)
3. Adding a new language requires only `lang/X/` files
4. Adding a new scheme requires only `scheme/X/` files
5. No coupling between schemes and specific languages
