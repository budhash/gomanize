# Romanization Engine Architecture Proposal

## Problem Statement

The current single-pass processing mixes **parsing** and **rendering**, which:
1. Prevents rules from seeing full word context
2. Makes rules interfere with each other
3. Limits extensibility (IAST, reverse transliteration, new languages)
4. Cannot coordinate decisions across positions (schwa deletion)

This proposal redesigns the core to be a proper **romanization engine** with clean separation of concerns.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                      ROMANIZATION ENGINE                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                    GENERIC CORE (engine/)                     │  │
│  │                                                               │  │
│  │  Data Structures    Rule System       Config                 │  │
│  │  ───────────────    ───────────       ──────                 │  │
│  │  • Unit             • Rule            • Language             │  │
│  │  • Word             • RulePhase       • Scheme               │  │
│  │  • ConsonantRun     • RuleScope       • Options              │  │
│  │  • Position         • RuleMode                               │  │
│  │  • SymbolMap        • RuleEngine                             │  │
│  │                                                               │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                              │                                      │
│                              ▼                                      │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                 LANGUAGE-SPECIFIC (lang/)                     │  │
│  │                                                               │  │
│  │  hindi/              marathi/           bengali/              │  │
│  │  ────────            ─────────          ─────────             │  │
│  │  • HindiSymbols      • MarathiSymbols   • BengaliSymbols     │  │
│  │  • HindiRules        • MarathiRules     • BengaliRules       │  │
│  │  • Hindi (Language)  • Marathi          • Bengali            │  │
│  │                                                               │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                              │                                      │
│                              ▼                                      │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                   SCHEMES (scheme/)                           │  │
│  │                                                               │  │
│  │  • Hunterian (default)                                       │  │
│  │  • IAST (scholarly with diacritics)                          │  │
│  │  • ISO-15919                                                 │  │
│  │                                                               │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Part 1: Core Data Structures

### Symbol Mapping (Generic)

```go
// Category classifies Devanagari characters
type Category int
const (
    Vowel Category = iota
    Consonant
    Conjunct
    Number
    Symbol
    Halant
)

// SymbolInfo holds romanization data for a character
type SymbolInfo struct {
    Category Category
    BaseRom  string    // Default romanization
    Alts     []string  // Alternates (for future use)
}

// SymbolMap is the generic type - works for any Indic script
type SymbolMap map[string]SymbolInfo

// Helper methods
func (m SymbolMap) Lookup(s string) (SymbolInfo, bool)
func (m SymbolMap) IsVowel(s string) bool
func (m SymbolMap) IsConsonant(s string) bool
```

### Unit (Parsed Element)

```go
// Position tracks location in source text
type Position struct {
    Offset int  // Byte offset in original string
    Rune   int  // Rune (character) index
}

// UnitType classifies parsed units
type UnitType int
const (
    UnitVowel UnitType = iota
    UnitConsonant
    UnitConjunct
    UnitSymbol
    UnitNumber
)

// SchwaState tracks schwa deletion decisions
type SchwaState int
const (
    SchwaPending SchwaState = iota  // Not yet decided
    SchwaKeep                        // Definitely keep
    SchwaDelete                      // Definitely delete
)

// Unit represents a single phonetic unit
type Unit struct {
    // Source tracking
    Runes    []rune    // Original characters (1-3 for conjuncts)
    Start    Position  // Start position in source
    End      Position  // End position in source

    // Classification
    Type     UnitType
    BaseRom  string    // Base romanization (modifiable by rules)

    // Precomputed context
    AfterHalant bool   // Was preceded by ्

    // State (for consonants)
    Schwa    SchwaState

    // Bidirectional links
    Prev     *Unit
    Next     *Unit

    // Run membership (nil for vowels)
    Run      *ConsonantRun
    RunIndex int
}

// Helper methods
func (u *Unit) IsWordFinal() bool { return u.Next == nil }
func (u *Unit) IsWordInitial() bool { return u.Prev == nil }
func (u *Unit) DebugString() string
```

### Word (Container)

```go
// Word is the complete parsed representation
type Word struct {
    Units    []*Unit
    Runs     []*ConsonantRun
    Original string
    Options  *Options  // Runtime config
}

func (w *Word) DebugDump() string
func (w *Word) Render(scheme *Scheme) string
```

### ConsonantRun (Schwa Coordination)

```go
// ConsonantRun represents consecutive consonants between vowels
type ConsonantRun struct {
    Units     []*Unit
    PrevVowel *Unit   // nil if word-initial
    NextVowel *Unit   // nil if word-final
    DeletedAt int     // Index where schwa was deleted (-1 if none)
}
```

---

## Part 2: Rule System

### Rule Definition

```go
// RulePhase determines when a rule executes
type RulePhase int
const (
    PhaseSchwa      RulePhase = iota  // Schwa decisions
    PhaseConsonant                     // Consonant modifications
    PhaseVowel                         // Vowel modifications
    PhaseRender                        // Final output
)

// RuleScope determines which languages/schemes a rule applies to
type RuleScope int
const (
    ScopeUniversal  RuleScope = iota  // All languages (priority base: 0)
    ScopeScript                        // Script-specific (priority base: 100)
    ScopeLanguage                      // Language-specific (priority base: 200)
    ScopeScheme                        // Scheme-specific (priority base: 300)
)

// RuleMode determines execution behavior
type RuleMode int
const (
    ModeExclusive RuleMode = iota  // First match wins (default for schwa)
    ModeAlways                      // Always run if condition matches
    ModeFallback                    // Only if no other rule acted
)

// Rule represents a single transliteration rule
type Rule struct {
    Name        string
    Description string
    Scope       RuleScope
    Priority    int         // 0-99 within scope
    Phase       RulePhase
    Mode        RuleMode
    Scripts     []string    // If ScopeScript
    Languages   []string    // If ScopeLanguage
    Schemes     []string    // If ScopeScheme
    Condition   func(*Unit, *Word) bool
    Action      func(*Unit, *Word)
}

// EffectivePriority calculates priority across scopes
func (r *Rule) EffectivePriority() int {
    return int(r.Scope)*100 + r.Priority
}
```

### Priority System

```
Effective priority = (Scope × 100) + Priority

┌────────────────────────────────────────────────────────────────┐
│                    PRIORITY CALCULATION                         │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Scope         Base    Priority Range   Example                │
│  ─────         ────    ──────────────   ───────                │
│  Scheme        300     300-399          IAST diacritics        │
│  Language      200     200-299          Hindi व→w              │
│  Script        100     100-199          Devanagari C+C+V       │
│  Universal     0       0-99             Word-final schwa       │
│                                                                 │
│  Higher effective priority = runs first                        │
│  Within same scope: higher Priority field = runs first         │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

### Rule Modes

```
┌────────────────────────────────────────────────────────────────┐
│                      RULE EXECUTION MODES                       │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  EXCLUSIVE (default for PhaseSchwa)                            │
│  ─────────────────────────────────                             │
│  First matching rule wins. Others skip if state changed.       │
│  Use for: Mutually exclusive decisions (schwa keep/delete)     │
│                                                                 │
│  ALWAYS (default for PhaseConsonant, PhaseVowel)               │
│  ───────────────────────────────────────────────               │
│  Runs regardless of current state. Rules chain in order.       │
│  Use for: Transformations (व→w), scheme overrides (aa→ā)       │
│                                                                 │
│  FALLBACK                                                      │
│  ────────                                                      │
│  Only runs if NO other rule in phase has acted on unit.        │
│  Use for: Defaults (if nothing matched, keep schwa)            │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

### Rule Sharing Across Languages/Schemes

```
┌────────────────────────────────────────────────────────────────┐
│                     RULE SHARING MODEL                          │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  UNIVERSAL RULES (ScopeUniversal)                              │
│  Apply to ALL Indic languages                                  │
│  • number-passthrough                                          │
│  • schwa-delete-word-final                                     │
│                                                                 │
│  SCRIPT RULES (ScopeScript)                                    │
│  Apply to all languages using a script                         │
│  • schwa-delete-ccv (Devanagari)                               │
│  • conjunct-detection (Devanagari)                             │
│                                                                 │
│  LANGUAGE RULES (ScopeLanguage)                                │
│  Apply to specific language only                               │
│  • va-to-wa-conjunct (Hindi only)                              │
│  • marathi-schwa-retention (Marathi only)                      │
│                                                                 │
│  SCHEME RULES (ScopeScheme)                                    │
│  Apply based on output scheme                                  │
│  • iast-long-aa (IAST: aa→ā)                                   │
│  • long-vowels-mode (--long-vowels flag)                       │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

### Rule Engine

```go
type RuleEngine struct {
    allRules []Rule              // All registered rules
    active   map[RulePhase][]Rule // Filtered and sorted for current config
    trace    *Trace              // nil if not tracing
}

func NewRuleEngine(lang *Language, scheme *Scheme) *RuleEngine

func (e *RuleEngine) Apply(word *Word) {
    for _, phase := range []RulePhase{PhaseSchwa, PhaseConsonant, PhaseVowel, PhaseRender} {
        acted := make(map[*Unit]bool)

        // First pass: Exclusive and Always rules
        for _, rule := range e.active[phase] {
            if rule.Mode == ModeFallback {
                continue
            }
            for _, unit := range word.Units {
                if rule.Mode == ModeExclusive && acted[unit] {
                    continue
                }
                if rule.Condition(unit, word) {
                    e.traceEvent(rule, unit, "fired")
                    rule.Action(unit, word)
                    acted[unit] = true
                }
            }
        }

        // Second pass: Fallback rules
        for _, rule := range e.active[phase] {
            if rule.Mode != ModeFallback {
                continue
            }
            for _, unit := range word.Units {
                if acted[unit] {
                    continue
                }
                if rule.Condition(unit, word) {
                    e.traceEvent(rule, unit, "fallback")
                    rule.Action(unit, word)
                    acted[unit] = true
                }
            }
        }
    }
}

func (e *RuleEngine) ListRules() []Rule
func (e *RuleEngine) EnableTrace()
func (e *RuleEngine) TraceOutput() string
```

### Conflict Prevention

```go
// Conflicts are prevented by design:
// 1. Different scopes have different base priorities
// 2. Same-scope rules must have different Priority values
// 3. Registration-time validation catches errors

func (e *RuleEngine) AddRule(r Rule) error {
    for _, existing := range e.allRules {
        if existing.Phase == r.Phase &&
           existing.Mode == ModeAlways && r.Mode == ModeAlways &&
           existing.EffectivePriority() == r.EffectivePriority() {
            return fmt.Errorf("conflict: rules %q and %q have same effective priority %d",
                existing.Name, r.Name, r.EffectivePriority())
        }
    }
    e.allRules = append(e.allRules, r)
    return nil
}
```

---

## Part 3: Composition Model

### Why Composition Over Inheritance

The engine uses **composition** rather than base/derived classes:

```
┌────────────────────────────────────────────────────────────────┐
│                    COMPOSITION vs INHERITANCE                   │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  INHERITANCE (rejected)                                        │
│  ──────────────────────                                        │
│  type DevanagariBase struct { ... }                            │
│  type Hindi struct { DevanagariBase }  // embeds               │
│  type Marathi struct { DevanagariBase }                        │
│                                                                 │
│  Problems:                                                      │
│  • Script ≠ Language (Hindi/Marathi share script, differ in    │
│    schwa rules)                                                │
│  • Diamond problem if language uses multiple scripts           │
│  • Hard to override specific behaviors                         │
│  • Schemes don't fit hierarchy (IAST applies to all)           │
│                                                                 │
│  COMPOSITION (used)                                            │
│  ──────────────────                                            │
│  engine.AddRules(UniversalRules...)                            │
│  engine.AddRules(DevanagariRules...)                           │
│  engine.AddRules(HindiRules...)                                │
│  engine.AddRules(HunterianRules...)                            │
│                                                                 │
│  Benefits:                                                      │
│  • Each layer adds/overrides/disables independently            │
│  • Schemes orthogonal to languages                             │
│  • No inheritance complexity                                   │
│  • Easy to test layers in isolation                            │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

### Rule Layering Example: Hindi vs Marathi

```go
// Both share Devanagari script rules
engine := NewRuleEngine()
engine.AddRules(UniversalRules...)
engine.AddRules(DevanagariRules...)  // schwa-delete-ccv, etc.

// Hindi adds its own rules
if lang == "hindi" {
    engine.AddRules(HindiRules...)   // va-to-wa-conjunct
}

// Marathi overrides schwa behavior
if lang == "marathi" {
    engine.AddRules(MarathiRules...) // schwa-retain-epenthetic
}
```

### Concrete Example: Schwa Deletion Differences

```
Word: अगरतला (Agartala)

┌─────────────────────────────────────────────────────────────────┐
│ HINDI                                                            │
├─────────────────────────────────────────────────────────────────┤
│ Rules fired (in priority order):                                │
│   1. [Script:50] schwa-delete-ccv → deletes at र               │
│   2. [Universal:0] schwa-keep-default → keeps ग, त, ल          │
│ Result: agartala                                                │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ MARATHI (hypothetical - retains more schwas)                    │
├─────────────────────────────────────────────────────────────────┤
│ Rules fired (in priority order):                                │
│   1. [Language:60] schwa-retain-clusters → keeps र              │
│   2. [Script:50] schwa-delete-ccv → SKIPPED (higher rule acted) │
│   3. [Universal:0] schwa-keep-default → keeps ग, त, ल          │
│ Result: agaratala                                               │
└─────────────────────────────────────────────────────────────────┘
```

### Scheme Independence

Schemes are completely orthogonal to languages:

```go
// Same parsed word, different output
word := engine.Parse("आकाश")
engine.Apply(word)

hunterianOutput := word.Render(Hunterian)  // "akash"
iastOutput := word.Render(IAST)            // "ākāś"
iso15919Output := word.Render(ISO15919)    // "ākāśa"
```

### Layer Override Mechanism

Higher-scope rules can explicitly disable lower-scope rules:

```go
// Language rule can disable a script rule for specific cases
var MarathiRules = []Rule{
    {
        Name:        "schwa-retain-clusters",
        Description: "Marathi retains schwa in consonant clusters",
        Scope:       ScopeLanguage,
        Languages:   []string{"marathi"},
        Priority:    60,  // Higher than Script:50 schwa-delete-ccv
        Phase:       PhaseSchwa,
        Mode:        ModeExclusive,
        Condition: func(u *Unit, w *Word) bool {
            // Marathi-specific cluster detection
            return u.Type == UnitConsonant &&
                   u.Schwa == SchwaPending &&
                   isMarathiCluster(u)
        },
        Action: func(u *Unit, w *Word) { u.Schwa = SchwaKeep },
    },
}
```

---

## Part 4: Language and Scheme Configuration

### Language Definition

```go
type Language struct {
    Name      string
    Script    string        // "devanagari", "bengali", etc.
    Symbols   SymbolMap
    MultiChar []string      // Conjuncts to try first: ["क्ष", "ज्ञ", "क़"]
    Rules     []Rule        // Language-specific rules
}

// Example
var Hindi = Language{
    Name:   "hindi",
    Script: "devanagari",
    Symbols: SymbolMap{
        "क": {Category: Consonant, BaseRom: "k"},
        "ख": {Category: Consonant, BaseRom: "kh"},
        // ... 100+ mappings
    },
    MultiChar: []string{"ज्ञ", "क्ष", "त्र", "क़", "ख़", "ग़", "ज़", "ड़", "ढ़", "फ़"},
    Rules:     HindiRules,
}
```

### Scheme Definition

```go
type Scheme struct {
    Name      string
    SymbolMap map[string]string  // 1:1 symbol overrides (applied at render)
    Rules     []Rule             // Context-dependent transforms
}
```

### Scheme Design: Hybrid Approach

Schemes use a **hybrid** of symbol overrides and rules:

| Mechanism | Purpose | Applied At |
|-----------|---------|------------|
| `SymbolMap` | Unconditional 1:1 mappings | Render time |
| `Rules` | Context-dependent transforms | Rule engine |

**Why hybrid?**
- IAST needs ~30 symbol overrides (श→ś, आ→ā, etc.)
- Writing 30 trivial rules is boilerplate
- But some transforms ARE context-dependent (ज्ञ→gya vs jña)

### Scheme Examples

```go
// Colloquial: Default for Hindi (song lyrics, casual use)
// - No diacritics (aa not ā)
// - Phonetic conjuncts (ज्ञ → gya)
// - Aggressive schwa deletion
var Colloquial = Scheme{
    Name:      "colloquial",
    SymbolMap: nil,  // Uses Language.Symbols BaseRom as-is
    Rules:     nil,  // Default schwa rules
}

// Hunterian: Official India government standard
// - Macrons for long vowels (ā, ī, ū)
// - Traditional conjunct spellings
// - Less schwa deletion
var Hunterian = Scheme{
    Name: "hunterian",
    SymbolMap: map[string]string{
        // Long vowels with macron
        "आ": "ā", "ई": "ī", "ऊ": "ū",
        "ा": "ā", "ी": "ī", "ू": "ū",
        // Diphthongs
        "ऐ": "ai", "औ": "au",
        "ै": "ai", "ौ": "au",
    },
    Rules: []Rule{
        // ज्ञ → jñ (not gya)
        {
            Name:  "hunterian-jnya",
            Phase: PhaseConsonant,
            Condition: func(u *Unit, w *Word) bool {
                return string(u.Runes) == "ज्ञ"
            },
            Action: func(u *Unit, w *Word) { u.BaseRom = "jñ" },
        },
    },
}

// IAST: International scholarly standard
// - Full diacritics (ś, ṣ, ṛ, etc.)
// - Preserves more schwas
var IAST = Scheme{
    Name: "iast",
    SymbolMap: map[string]string{
        // Vowels
        "आ": "ā", "इ": "i", "ई": "ī", "उ": "u", "ऊ": "ū",
        "ऋ": "ṛ", "ए": "e", "ऐ": "ai", "ओ": "o", "औ": "au",
        // Matras
        "ा": "ā", "ि": "i", "ी": "ī", "ु": "u", "ू": "ū",
        "े": "e", "ै": "ai", "ो": "o", "ौ": "au",
        // Sibilants
        "श": "ś", "ष": "ṣ", "स": "s",
        // Retroflex
        "ट": "ṭ", "ठ": "ṭh", "ड": "ḍ", "ढ": "ḍh", "ण": "ṇ",
        // Anusvara/Visarga
        "ं": "ṃ", "ः": "ḥ",
    },
    Rules: []Rule{
        // ज्ञ → jña
        {Name: "iast-jnya", ...},
        // Preserve final schwa in certain contexts
        {Name: "iast-final-schwa", ...},
    },
}
```

### Rendering with Scheme

```go
func (w *Word) Render(scheme *Scheme) string {
    var out strings.Builder
    for _, u := range w.Units {
        // 1. Check scheme symbol override first
        if scheme.SymbolMap != nil {
            if override, ok := scheme.SymbolMap[string(u.Runes)]; ok {
                out.WriteString(override)
                continue  // Skip to next unit
            }
        }

        // 2. Use BaseRom (possibly modified by rules)
        switch u.Type {
        case UnitConsonant, UnitConjunct:
            out.WriteString(u.BaseRom)
            if u.Next == nil || u.Next.Type != UnitVowel {
                if u.Schwa == SchwaKeep {
                    out.WriteString("a")
                }
            }
        case UnitVowel, UnitNumber, UnitSymbol:
            out.WriteString(u.BaseRom)
        }
    }
    return out.String()
}
```

### Language Default Scheme

Each language specifies its default scheme:

```go
type Language struct {
    Name          string
    Script        string
    Symbols       SymbolMap
    MultiChar     []string
    Rules         []Rule
    DefaultScheme string    // "colloquial" for Hindi
    Schemes       []string  // Available: ["colloquial", "hunterian", "iast"]
}

var Hindi = Language{
    Name:          "hindi",
    Script:        "devanagari",
    DefaultScheme: "colloquial",
    Schemes:       []string{"colloquial", "hunterian", "iast"},
    // ...
}
```

### CLI Usage

```bash
# Default scheme (colloquial for Hindi)
gomanize "भारत"                      # bharat

# Explicit scheme
gomanize --scheme hunterian "भारत"   # bhārat
gomanize --scheme iast "भारत"        # bhārata

# List available schemes
gomanize --list-schemes              # colloquial*, hunterian, iast
```

### Current Gomanize = Colloquial

The current gomanize implementation is the "colloquial" scheme:

| Aspect | Colloquial (current) | Hunterian | IAST |
|--------|---------------------|-----------|------|
| ज्ञ | gya | jñ | jña |
| आ | aa | ā | ā |
| Schwa | Aggressive deletion | Less deletion | Minimal deletion |
| Target | Song lyrics, casual | Official docs | Scholarly |

### Options (Runtime Flags)

```go
type Options struct {
    Scheme     string  // --scheme flag (default from Language)
    LongVowels bool    // --long-vowels flag
    Trace      bool    // --trace flag
}
```

### Transliterator (Combined)

```go
type Transliterator struct {
    lang    *Language
    scheme  *Scheme
    engine  *RuleEngine
    options *Options
}

func NewTransliterator(lang *Language, scheme *Scheme, opts *Options) *Transliterator

func (t *Transliterator) Transliterate(input string) string {
    word := t.parse(input)       // Uses lang.Symbols, lang.MultiChar
    word.Options = t.options
    word.identifyRuns()
    t.engine.Apply(word)         // Uses filtered rules (including scheme rules)
    return word.Render(t.scheme) // Applies scheme SymbolMap overrides
}
```

---

## Part 5: Debugging and Tracing

### Trace Events

```go
type TraceEvent struct {
    Phase    RulePhase
    Rule     string
    Unit     *Unit
    Position Position
    Action   string  // "fired", "skipped", "fallback"
    Before   string
    After    string
}

type Trace struct {
    Events []TraceEvent
}
```

### Schwa and Rendering

Schwa rules only decide state for consonants NOT followed by explicit vowels. The rendering logic (shown in Part 4) handles:
- **Consonant + Vowel** → consonant only (vowel provides sound)
- **Consonant + Consonant** → schwa rules decide (Keep/Delete)
- **Consonant + END** → schwa-delete-word-final rule fires

### Trace Output Example

```
$ gomanize --trace "अगरतला"

=== PARSE ===
Unit[0] अ       @ 0   Vowel      "a"
Unit[1] ग       @ 1   Consonant  "g"   Schwa:Pending
Unit[2] र       @ 2   Consonant  "r"   Schwa:Pending
Unit[3] त       @ 3   Consonant  "t"   Schwa:Pending
Unit[4] ल       @ 4   Consonant  "l"   Schwa:Pending
Unit[5] ा       @ 5   Vowel      "a"

=== RUNS ===
Run[0]: [ग र त ल] → ा  (DeletedAt: -1)

=== RULES (PhaseSchwa) ===
[Exclusive] "schwa-keep-initial-conjunct" @ ग: skipped
[Exclusive] "schwa-delete-ccv" @ र: FIRED (Pending→Delete, Run.DeletedAt=1)
[Exclusive] "schwa-delete-ccv" @ त: skipped (already acted)
[Fallback]  "schwa-keep-default" @ ग: FIRED (Pending→Keep)
[Fallback]  "schwa-keep-default" @ त: FIRED (Pending→Keep)

=== RENDER ===
Unit[0] अ  → "a"                    (vowel)
Unit[1] ग  → "g" + "a" = "ga"       (schwa=Keep, not followed by vowel)
Unit[2] र  → "r"                    (schwa=Delete)
Unit[3] त  → "t" + "a" = "ta"       (schwa=Keep, not followed by vowel)
Unit[4] ल  → "l"                    (followed by vowel, no schwa added)
Unit[5] ा  → "a"                    (vowel provides sound for ल)

Output: "a" + "ga" + "r" + "ta" + "l" + "a" = "agartala"
```

---

## Part 6: Example Rules

### Universal Rules

```go
var UniversalRules = []Rule{
    {
        Name:        "number-passthrough",
        Description: "Pass numbers through unchanged",
        Scope:       ScopeUniversal,
        Priority:    50,
        Phase:       PhaseRender,
        Mode:        ModeAlways,
        Condition:   func(u *Unit, w *Word) bool { return u.Type == UnitNumber },
        Action:      func(u *Unit, w *Word) { /* use BaseRom */ },
    },
    {
        Name:        "schwa-delete-word-final",
        Description: "Delete schwa at word end",
        Scope:       ScopeUniversal,
        Priority:    10,
        Phase:       PhaseSchwa,
        Mode:        ModeExclusive,
        Condition: func(u *Unit, w *Word) bool {
            return u.Type == UnitConsonant &&
                   u.Schwa == SchwaPending &&
                   u.Next == nil
        },
        Action: func(u *Unit, w *Word) { u.Schwa = SchwaDelete },
    },
}
```

### Devanagari Script Rules

```go
var DevanagariRules = []Rule{
    {
        Name:        "schwa-delete-ccv",
        Description: "Delete schwa in C+C+V pattern",
        Scope:       ScopeScript,
        Scripts:     []string{"devanagari"},
        Priority:    50,
        Phase:       PhaseSchwa,
        Mode:        ModeExclusive,
        Condition: func(u *Unit, w *Word) bool {
            return u.Type == UnitConsonant &&
                   u.Schwa == SchwaPending &&
                   u.Run != nil &&
                   u.Run.DeletedAt < 0 &&
                   u.RunIndex > 0 &&
                   u.Next != nil && u.Next.Type == UnitConsonant &&
                   u.Next.Next != nil && u.Next.Next.Type == UnitVowel
        },
        Action: func(u *Unit, w *Word) {
            u.Schwa = SchwaDelete
            u.Run.DeletedAt = u.RunIndex
        },
    },
    {
        Name:        "schwa-keep-initial-conjunct",
        Description: "Keep schwa for word-initial conjuncts",
        Scope:       ScopeScript,
        Scripts:     []string{"devanagari"},
        Priority:    80,
        Phase:       PhaseSchwa,
        Mode:        ModeExclusive,
        Condition: func(u *Unit, w *Word) bool {
            return u.Type == UnitConsonant &&
                   u.Schwa == SchwaPending &&
                   u.AfterHalant &&
                   u.RunIndex <= 1
        },
        Action: func(u *Unit, w *Word) { u.Schwa = SchwaKeep },
    },
}
```

### Hindi-Specific Rules

```go
var HindiRules = []Rule{
    {
        Name:        "va-to-wa-conjunct",
        Description: "व→w after स, श, द, ख",
        Scope:       ScopeLanguage,
        Languages:   []string{"hindi"},
        Priority:    50,
        Phase:       PhaseConsonant,
        Mode:        ModeAlways,
        Condition: func(u *Unit, w *Word) bool {
            if u.BaseRom != "v" || u.Prev == nil {
                return false
            }
            prev := u.Prev.BaseRom
            return u.AfterHalant &&
                   (prev == "s" || prev == "sh" || prev == "d" || prev == "kh")
        },
        Action: func(u *Unit, w *Word) { u.BaseRom = "w" },
    },
}
```

### Scheme-Specific Rules

```go
var SchemeRules = []Rule{
    {
        Name:        "iast-long-aa",
        Description: "Transform aa→ā for IAST",
        Scope:       ScopeScheme,
        Schemes:     []string{"iast"},
        Priority:    50,
        Phase:       PhaseVowel,
        Mode:        ModeAlways,
        Condition: func(u *Unit, w *Word) bool {
            return u.BaseRom == "aa"
        },
        Action: func(u *Unit, w *Word) { u.BaseRom = "ā" },
    },
    {
        Name:        "long-vowels-all-aa",
        Description: "ा→aa everywhere (--long-vowels)",
        Scope:       ScopeScheme,
        Schemes:     []string{"long-vowels"},
        Priority:    90,
        Phase:       PhaseVowel,
        Mode:        ModeAlways,
        Condition: func(u *Unit, w *Word) bool {
            return w.Options != nil &&
                   w.Options.LongVowels &&
                   len(u.Runes) == 1 && u.Runes[0] == 'ा'
        },
        Action: func(u *Unit, w *Word) { u.BaseRom = "aa" },
    },
}
```

### Fallback Rules

```go
var FallbackRules = []Rule{
    {
        Name:        "schwa-keep-default",
        Description: "Default: keep schwa if undecided",
        Scope:       ScopeUniversal,
        Priority:    0,
        Phase:       PhaseSchwa,
        Mode:        ModeFallback,
        Condition: func(u *Unit, w *Word) bool {
            return u.Type == UnitConsonant && u.Schwa == SchwaPending
        },
        Action: func(u *Unit, w *Word) { u.Schwa = SchwaKeep },
    },
}
```

---

## Part 7: Implementation Plan

### Phase 1: Core Data Structures
1. Create `engine/` package with Unit, Word, ConsonantRun, Position
2. Implement parse() with multi-character support
3. Implement identifyRuns()
4. Add comprehensive unit tests

### Phase 2: Rule System
1. Implement Rule, RulePhase, RuleScope, RuleMode types
2. Implement RuleEngine with filtering and priority sorting
3. Implement Apply() with mode-aware execution
4. Add rule registration validation

### Phase 3: Migration
1. Extract current rules into Rule structs
2. Implement TransliterateV2() using new engine
3. Run parallel: verify V2 matches V1 output
4. Run full Dakshina test suite

### Phase 4: Fix Schwa Issues
1. Implement Run.DeletedAt tracking
2. Fix अगरतला and similar cases
3. Measure accuracy improvement
4. Tune rule priorities

### Phase 5: Cleanup and Extensions
1. Remove old single-pass code
2. Add --trace CLI flag
3. Add IAST scheme
4. Update documentation

---

## Part 8: Design Notes and Limitations

### DeletedAt Tracks One Deletion Per Run

`ConsonantRun.DeletedAt` tracks the index where schwa was deleted. This design assumes **at most one schwa deletion per consonant run**, which holds for standard Hindi patterns:

- C+C+V: delete at first C (e.g., समझ → samjh)
- C+C+C+V: delete at middle C (e.g., अगरतला → agartala)

For exotic cases with multiple deletions needed, the design would need extension (e.g., `DeletedPositions []int`). Not needed for current Hindi accuracy goals.

### Conjuncts Have Schwa Too

`UnitConjunct` (e.g., ज्ञ "gya") can have word-final schwa that needs deletion. The schwa rules should check `UnitConsonant || UnitConjunct` for completeness:

```go
Condition: func(u *Unit, w *Word) bool {
    return (u.Type == UnitConsonant || u.Type == UnitConjunct) &&
           u.Schwa == SchwaPending &&
           u.Next == nil
}
```

### Run Membership for Conjuncts

Conjuncts count as single units in runs. A word like ज्ञान (gyan) parses as:
- ज्ञ (conjunct, RunIndex=0)
- न (consonant, RunIndex=1)
- ा (vowel, NextVowel for the run)

---

## Part 9: Summary

### Core Primitives

| Primitive | Purpose |
|-----------|---------|
| `Unit` | Parsed phonetic element with bidirectional links |
| `Word` | Container with units, runs, options |
| `ConsonantRun` | Groups consonants for schwa coordination |
| `Position` | Source location tracking |
| `SymbolMap` | Generic character → romanization mapping |
| `Rule` | Condition + action with scope, phase, mode, priority |
| `RuleEngine` | Applies rules with filtering and tracing |
| `Language` | Symbols + rules for a language |
| `Scheme` | Output style overrides |
| `Options` | Runtime configuration |

### Key Design Decisions

1. **Parse first, decide later** - Full word context before any output
2. **Bidirectional links** - Rules can look forward AND backward
3. **Runs as first-class** - Schwa deletion coordinated per run
4. **Explicit rule priority** - Scope×100 + Priority
5. **Three execution modes** - Exclusive, Always, Fallback
6. **Conflict prevention** - Registration-time validation
7. **Rule sharing** - Universal → Script → Language → Scheme
8. **Full tracing** - Debug any word step by step

### Benefits

| Current | Proposed |
|---------|----------|
| Rules buried in if/else | Rules as enumerable data |
| Implicit priority | Explicit scope + priority |
| Can't test rules alone | `engine.TestRule()` |
| No debugging | `--trace` shows everything |
| Single scheme | Same parse, multiple renders |
| Hindi only | Language-agnostic core |
| Can't extend | `engine.AddRule()` |

This architecture directly addresses the schwa deletion issues we encountered AND provides a foundation for future enhancements (IAST, reverse transliteration, new languages).
