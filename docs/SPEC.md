# Gomanize Architecture Specification

## Overview

Gomanize is a transliteration engine designed to convert scripts (Devanagari, Arabic, etc.) to romanized output. The architecture supports multiple scripts, languages, and output schemes while keeping each component maintainable and scalable.

**Goal**: Adding a new language or scheme should be a mechanical process - implement the required interfaces, no architectural rethinking needed.

## Design Principles

1. **Separation of Concerns**: Each layer has a single responsibility
2. **Rules are Agnostic**: Rules don't know which schemes use them
3. **Schemes Select, Don't Define**: Schemes choose from a language's rule catalog
4. **Scripts Provide Capabilities**: Languages decide how to use them
5. **Scalability**: 50 languages × 5 schemes should work without coupling

## Architecture Layers

```
┌─────────────────────────────────────────────────────────────────────┐
│                           CORE (core/)                              │
│                     Universal Mechanics                             │
├─────────────────────────────────────────────────────────────────────┤
│  Types:      Unit, Word, Position, Category, UnitType               │
│  Mechanics:  Rule, RuleEngine, RulePhase, RuleScope, RuleMode       │
│  Interfaces: Script, Language, Scheme, Parser, Renderer             │
│  Catalog:    RuleCatalog (organizes rules by category)              │
│  Engine:     Orchestrates the pipeline                              │
└─────────────────────────────────────────────────────────────────────┘
                               │
         ┌─────────────────────┼─────────────────────┐
         ▼                     ▼                     ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│ SCRIPT          │  │ SCRIPT          │  │ SCRIPT          │
│ script/brahmic/ │  │ script/arabic/  │  │ script/latin/   │
├─────────────────┤  │ (future)        │  │ (future)        │
│ • Parser        │  └─────────────────┘  └─────────────────┘
│ • Renderer      │
│ • PrepareWord   │
│ • ScriptData    │
│ • Categories    │
└─────────────────┘
         │
    ┌────┴────────┬──────────┬──────────┐
    ▼             ▼          ▼          ▼
┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐
│LANGUAGE │  │LANGUAGE │  │LANGUAGE │  │LANGUAGE │
│ hindi   │  │ marathi │  │  tamil  │  │ bengali │
├─────────┤  └─────────┘  └─────────┘  └─────────┘
│ Symbols │
│ Config  │
│ Rules() │ → RuleCatalog (ALL rules for this language)
└─────────┘
         │
         └──────────────┬──────────────┬──────────────┐
                        ▼              ▼              ▼
                   ┌─────────┐   ┌─────────┐   ┌─────────┐
                   │ SCHEME  │   │ SCHEME  │   │ SCHEME  │
                   │colloq.  │   │  iast   │   │hunter.  │
                   ├─────────┤   └─────────┘   └─────────┘
                   │SelectRules(catalog) → []Rule       │
                   └────────────────────────────────────┘
```

## Layer Specifications

### 1. Core (`core/`)

Universal types and mechanics that work for ANY script.

#### Types (`core/types.go`)

```go
// Position tracks location in source text for debugging
type Position struct {
    Offset int  // Byte offset in original string
    Rune   int  // Rune (character) index
}

// UnitType classifies parsed phonetic units
type UnitType int

const (
    UnitVowel     UnitType = iota  // Vowels and matras
    UnitModifier                    // Anusvara, visarga, chandrabindu
    UnitConsonant                   // Single consonant
    UnitConjunct                    // Multi-character conjunct (ज्ञ)
    UnitNumber                      // Numerals
    UnitSymbol                      // Punctuation, other
)

// Unit represents a single phonetic unit in the parsed word
type Unit struct {
    // Source tracking
    Runes []rune
    Start Position
    End   Position

    // Classification
    Type    UnitType
    BaseRom string  // Base romanization (modifiable by rules)

    // Navigation
    Prev *Unit
    Next *Unit

    // Script-specific extension point
    ScriptData interface{}
}

// Word is the complete parsed representation
type Word struct {
    Units    []*Unit
    Original string
    Options  Options
}

// Options configures transliteration behavior
type Options struct {
    LongVowels bool  // Output "aa" for all ा positions
}
```

#### Categories (`core/categories.go`)

```go
// Category classifies symbols for parsing
type Category int

const (
    CatUnknown Category = iota
    CatVowel
    CatConsonant
    CatNumber
    CatSymbol
    // Script-specific categories added by scripts
)

// SymbolInfo holds romanization info for a symbol
type SymbolInfo struct {
    Category Category
    BaseRom  string
}

// SymbolMap maps script characters to their info
type SymbolMap map[string]SymbolInfo
```

#### Rules (`core/rule.go`)

```go
// RulePhase determines when a rule executes
type RulePhase int

const (
    PhaseSchwa     RulePhase = iota  // Schwa deletion/retention
    PhaseConsonant                    // Consonant modifications (व→w)
    PhaseVowel                        // Vowel modifications
    PhaseRender                       // Output transforms (aa→ā)
)

// RuleScope determines priority tier
type RuleScope int

const (
    ScopeUniversal RuleScope = iota  // Base 0
    ScopeScript                       // Base 100
    ScopeLanguage                     // Base 200
    ScopeScheme                       // Base 300
)

// RuleMode determines execution behavior
type RuleMode int

const (
    ModeExclusive RuleMode = iota  // First match wins
    ModeAlways                      // Always run if condition matches
    ModeFallback                    // Only if no other rule acted
)

// Rule represents a single transliteration rule
type Rule struct {
    Name      string
    Phase     RulePhase
    Scope     RuleScope
    Priority  int  // 0-99 within scope
    Mode      RuleMode
    Condition func(*Unit, *Word) bool
    Action    func(*Unit, *Word)
}

// RuleEngine applies rules to parsed words
type RuleEngine struct {
    rules map[RulePhase][]Rule  // Sorted by priority
}
```

#### Rule Catalog (`core/catalog.go`)

```go
// RuleCatalog organizes rules by category
// Languages provide complete catalogs; schemes select from them
type RuleCatalog struct {
    Schwa     []Rule  // All schwa-related rules
    Consonant []Rule  // All consonant rules
    Vowel     []Rule  // All vowel rules
    Render    []Rule  // All output transform rules
}

// AllRules returns all rules in the catalog as a flat slice
func (c RuleCatalog) AllRules() []Rule

// FindByName finds a rule by name, returns nil if not found
func (c RuleCatalog) FindByName(name string) *Rule
```

#### Interfaces (`core/interfaces.go`)

```go
// Parser converts script input into Units
type Parser interface {
    Parse(input string, symbols SymbolMap) *Word
}

// Renderer converts Units into romanized output
type Renderer interface {
    Render(word *Word) string
}

// Script defines a script family (Brahmic, Arabic, Latin)
type Script interface {
    Name() string
    NewParser(config interface{}) Parser
    NewRenderer() Renderer
    PrepareWord(word *Word)  // Script-specific processing (e.g., IdentifyRuns)
    Categories() []Category  // Script-specific categories
}

// Language defines language-specific behavior
type Language interface {
    Name() string
    Script() Script
    Symbols() SymbolMap
    ScriptConfig() interface{}  // Script-specific config (halant, nukta)
    Rules() RuleCatalog         // ALL rules this language knows
}

// Scheme selects which rules to use
type Scheme interface {
    Name() string
    SelectRules(catalog RuleCatalog) []Rule
}
```

#### Engine (`core/engine.go`)

```go
// Engine orchestrates the transliteration pipeline
type Engine struct {
    script     Script
    symbols    SymbolMap
    config     interface{}
    ruleEngine *RuleEngine
    renderer   Renderer
}

// NewEngine creates an engine for a language + scheme combination
func NewEngine(lang Language, scheme Scheme) *Engine {
    catalog := lang.Rules()
    selectedRules := scheme.SelectRules(catalog)

    return &Engine{
        script:     lang.Script(),
        symbols:    lang.Symbols(),
        config:     lang.ScriptConfig(),
        ruleEngine: NewRuleEngine(selectedRules),
        renderer:   lang.Script().NewRenderer(),
    }
}

// Transliterate converts script text to romanized form
func (e *Engine) Transliterate(input string) string {
    // 1. Parse
    parser := e.script.NewParser(e.config)
    word := parser.Parse(input, e.symbols)

    // 2. Prepare (script-specific)
    e.script.PrepareWord(word)

    // 3. Apply rules
    e.ruleEngine.Apply(word)

    // 4. Render
    return e.renderer.Render(word)
}

// TransliterateWithOptions allows custom options
func (e *Engine) TransliterateWithOptions(input string, opts Options) string
```

### 2. Script (`script/`)

Script-family-specific parsing and rendering.

#### Brahmic Script (`script/brahmic/`)

Handles Devanagari, Bengali, Tamil, Gujarati, etc.

**Script-Specific Types** (`script/brahmic/types.go`):

```go
// SchwaState tracks schwa deletion decisions
type SchwaState int

const (
    SchwaPending SchwaState = iota
    SchwaKeep
    SchwaDelete
)

// ConsonantRun represents consecutive consonants
type ConsonantRun struct {
    Units     []*core.Unit
    PrevVowel *core.Unit
    NextVowel *core.Unit
    DeletedAt int  // -1 if none
}

// BrahmicData is stored in Unit.ScriptData
type BrahmicData struct {
    Schwa       SchwaState
    AfterHalant bool
    Run         *ConsonantRun
    RunIndex    int
}

// Config for Brahmic languages
type Config struct {
    Halant string  // e.g., "्" for Devanagari
    Nukta  string  // e.g., "़" for Devanagari
}
```

**Script-Specific Categories** (`script/brahmic/categories.go`):

```go
const (
    CatHalant      core.Category = 100 + iota
    CatAnusvara
    CatVisarga
    CatChandrabindu
    CatNukta
    CatMatra
    CatConjunct
)
```

**Script Interface Implementation** (`script/brahmic/brahmic.go`):

```go
type Brahmic struct{}

func (b Brahmic) Name() string { return "brahmic" }

func (b Brahmic) NewParser(config interface{}) core.Parser {
    cfg := config.(Config)
    return &Parser{halant: cfg.Halant, nukta: cfg.Nukta}
}

func (b Brahmic) NewRenderer() core.Renderer {
    return &Renderer{}
}

func (b Brahmic) PrepareWord(word *core.Word) {
    IdentifyRuns(word)
}

func (b Brahmic) Categories() []core.Category {
    return []core.Category{
        CatHalant, CatAnusvara, CatVisarga,
        CatChandrabindu, CatNukta, CatMatra, CatConjunct,
    }
}
```

### 3. Language (`lang/`)

Language-specific symbols and rule catalog.

#### Hindi Example (`lang/hindi/`)

**Symbols** (`lang/hindi/symbols.go`):

```go
var Symbols = core.SymbolMap{
    "क": {Category: brahmic.CatConsonant, BaseRom: "k"},
    "ख": {Category: brahmic.CatConsonant, BaseRom: "kh"},
    // ... all Devanagari mappings
    "ज्ञ": {Category: brahmic.CatConjunct, BaseRom: "gy"},
}

var MultiChar = []string{"ज्ञ", "क़", "ख़", ...}
```

**Rules** (`lang/hindi/rules.go`):

```go
func (h Hindi) Rules() core.RuleCatalog {
    return core.RuleCatalog{
        Schwa: []core.Rule{
            SchwaDeleteWordFinal,
            SchwaDeleteCCV,
            SchwaKeepInternalConjunct,
            SchwaKeepSonorousFinal,
            SchwaKeepIyaSuffix,
            SchwaKeepAll,  // For scholarly schemes
            SchwaKeepDefault,
        },
        Consonant: []core.Rule{
            VaToWaConjunct,
            VaKeepV,  // Alternative for scholarly
        },
        Vowel: []core.Rule{
            LongAaClosedFinal,
            LongAaAll,
        },
        Render: []core.Rule{
            RenderAaMacron,  // aa→ā
            RenderShAcute,   // sh→ś
            // ... other IAST transforms
        },
    }
}
```

**Language Implementation** (`lang/hindi/hindi.go`):

```go
type Hindi struct{}

func (h Hindi) Name() string { return "hindi" }

func (h Hindi) Script() core.Script { return brahmic.Brahmic{} }

func (h Hindi) Symbols() core.SymbolMap { return Symbols }

func (h Hindi) ScriptConfig() interface{} {
    return brahmic.Config{Halant: "्", Nukta: "़"}
}

func (h Hindi) Rules() core.RuleCatalog { return rules() }
```

### 4. Scheme (`scheme/`)

Selects which rules to use from a language's catalog.

#### Colloquial (`scheme/colloquial/`)

```go
type Colloquial struct{}

func (c Colloquial) Name() string { return "colloquial" }

func (c Colloquial) SelectRules(catalog core.RuleCatalog) []core.Rule {
    var rules []core.Rule

    // Schwa rules for colloquial
    rules = appendIfFound(rules, catalog.Schwa, "schwa-delete-word-final")
    rules = appendIfFound(rules, catalog.Schwa, "schwa-delete-ccv")
    rules = appendIfFound(rules, catalog.Schwa, "schwa-keep-internal-conjunct")
    rules = appendIfFound(rules, catalog.Schwa, "schwa-keep-sonorous-final")
    rules = appendIfFound(rules, catalog.Schwa, "schwa-keep-iya-suffix")
    rules = appendIfFound(rules, catalog.Schwa, "schwa-keep-default")

    // Consonant rules
    rules = appendIfFound(rules, catalog.Consonant, "va-to-wa-conjunct")

    // Vowel rules
    rules = appendIfFound(rules, catalog.Vowel, "long-aa-closed-final")

    // No render transforms (output as-is)

    return rules
}
```

#### IAST (`scheme/iast/`)

```go
type IAST struct{}

func (i IAST) Name() string { return "iast" }

func (i IAST) SelectRules(catalog core.RuleCatalog) []core.Rule {
    var rules []core.Rule

    // IAST keeps all schwas
    rules = appendIfFound(rules, catalog.Schwa, "schwa-keep-all")

    // Keep व as v
    rules = appendIfFound(rules, catalog.Consonant, "va-keep-v")

    // Render transforms for diacritics
    rules = appendIfFound(rules, catalog.Render, "render-aa-macron")
    rules = appendIfFound(rules, catalog.Render, "render-sh-acute")
    // ... other IAST transforms

    return rules
}
```

## Execution Pipeline

```
Input: "नमस्ते"
         │
         ▼
┌─────────────────────────────────────────────────────────────┐
│ 1. PARSE (Script.Parser)                                    │
│    "न" → Unit{Type:Consonant, BaseRom:"n", ScriptData:...}  │
│    "म" → Unit{Type:Consonant, BaseRom:"m", ScriptData:...}  │
│    "स्" → (halant tracked)                                  │
│    "ते" → Unit{Type:Consonant, BaseRom:"t"} + matra "e"     │
└─────────────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. PREPARE (Script.PrepareWord)                             │
│    IdentifyRuns() → ConsonantRun{न, म, स्त}                 │
│    Set Schwa=SchwaPending for consonants                    │
└─────────────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. APPLY RULES (RuleEngine)                                 │
│    PhaseSchwa:                                              │
│      schwa-keep-internal-conjunct → स्त keeps schwa         │
│      schwa-delete-word-final → त deletes schwa              │
│    PhaseConsonant:                                          │
│      (no changes)                                           │
│    PhaseVowel:                                              │
│      (no changes)                                           │
│    PhaseRender:                                             │
│      (no changes for colloquial)                            │
└─────────────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. RENDER (Script.Renderer)                                 │
│    Concatenate BaseRom, respecting SchwaState               │
│    Output: "namaste"                                        │
└─────────────────────────────────────────────────────────────┘
```

## Scaling Analysis

### Adding a New Language (e.g., Marathi)

1. Create `lang/marathi/symbols.go` - character mappings
2. Create `lang/marathi/rules.go` - rule catalog with Marathi-specific schwa rules
3. Create `lang/marathi/marathi.go` - implement `Language` interface
4. **Done** - all existing schemes work automatically

### Adding a New Scheme (e.g., Hunterian)

1. Create `scheme/hunterian/hunterian.go`
2. Implement `SelectRules()` to choose appropriate rules by name
3. **Done** - works with all existing languages

### Adding a New Script (e.g., Arabic)

1. Create `script/arabic/types.go` - Arabic-specific data structures
2. Create `script/arabic/parser.go` - RTL parsing, harakat handling
3. Create `script/arabic/renderer.go` - Arabic rendering
4. Create `script/arabic/arabic.go` - implement `Script` interface
5. **Done** - can now add Arabic languages

## Key Design Decisions

### 1. Rules are Agnostic
Rules don't know which schemes use them. They're building blocks that schemes compose.

### 2. Languages Own Rules
Each language defines ALL its possible rules in a RuleCatalog. This keeps language-specific logic together and enables schemes to select without knowing implementation details.

### 3. Schemes Select, Don't Define
Schemes choose from the language's catalog by rule name. If a rule doesn't exist (e.g., Tamil has no "schwa-delete-word-final"), it's skipped.

### 4. Scripts Provide Extension Points
`Unit.ScriptData interface{}` allows scripts to attach their own data (SchwaState, ConsonantRun) without core knowing about it.

### 5. PrepareWord Hook
Scripts can perform pre-rule processing (like IdentifyRuns) that rules depend on.

## Directory Structure

```
gomanize/
├── core/                    # Universal mechanics
│   ├── types.go            # Unit, Word, Position, Options
│   ├── categories.go       # Category, SymbolMap, SymbolInfo
│   ├── rule.go             # Rule, RuleEngine, phases, scopes
│   ├── catalog.go          # RuleCatalog
│   ├── interfaces.go       # Script, Language, Scheme, Parser, Renderer
│   └── engine.go           # Engine orchestrator
│
├── script/                  # Script family implementations
│   └── brahmic/
│       ├── brahmic.go      # Script interface impl
│       ├── types.go        # SchwaState, ConsonantRun, BrahmicData
│       ├── categories.go   # Brahmic-specific categories
│       ├── parser.go       # Brahmic parser
│       ├── renderer.go     # Brahmic renderer
│       └── runs.go         # IdentifyRuns
│
├── lang/                    # Language implementations
│   ├── hindi/
│   │   ├── hindi.go        # Language interface impl
│   │   ├── symbols.go      # Symbol mappings
│   │   └── rules.go        # Rule catalog
│   ├── marathi/            # (future)
│   └── tamil/              # (future)
│
├── scheme/                  # Scheme implementations
│   ├── colloquial/
│   │   └── colloquial.go   # SelectRules for colloquial
│   ├── iast/               # (future)
│   │   └── iast.go
│   └── hunterian/          # (future)
│       └── hunterian.go
│
└── gomanize.go             # Public API
```

## Success Criteria

1. All existing tests pass
2. Dakshina accuracy maintained (82.5%+)
3. Adding a new language requires only `lang/X/` files
4. Adding a new scheme requires only `scheme/X/` files
5. No coupling between schemes and specific languages
6. Clean separation: core knows nothing about Brahmic concepts
