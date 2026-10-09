// Package core provides universal types and mechanics for transliteration.
// This package is script-agnostic - it knows nothing about Brahmic, Arabic, etc.
package core

// Position tracks location in source text for debugging.
type Position struct {
	Offset int // Byte offset in original string
	Rune   int // Rune (character) index
}

// UnitType classifies parsed phonetic units.
type UnitType int

const (
	UnitVowel     UnitType = iota // Vowels and matras that replace inherent schwa
	UnitModifier                  // Modifiers that follow the vowel (anusvara, visarga, chandrabindu)
	UnitConsonant                 // Single consonant
	UnitConjunct                  // Multi-character conjunct (e.g., ज्ञ)
	UnitNumber                    // Numerals
	UnitSymbol                    // Other symbols (punctuation, etc.)
)

func (t UnitType) String() string {
	switch t {
	case UnitVowel:
		return "Vowel"
	case UnitModifier:
		return "Modifier"
	case UnitConsonant:
		return "Consonant"
	case UnitConjunct:
		return "Conjunct"
	case UnitNumber:
		return "Number"
	case UnitSymbol:
		return "Symbol"
	default:
		return "Unknown"
	}
}

// Unit represents a single phonetic unit in the parsed word.
type Unit struct {
	// Source tracking
	Runes []rune   // Original characters (1-3 for conjuncts)
	Start Position // Start position in source
	End   Position // End position in source

	// Classification
	Type    UnitType
	BaseRom string // Base romanization (modifiable by rules)

	// Navigation (bidirectional links)
	Prev *Unit
	Next *Unit

	// Script-specific extension point
	// Scripts store their own data here (e.g., SchwaState for Brahmic)
	ScriptData interface{}
}

// IsWordFinal returns true if this is the last unit in the word.
func (u *Unit) IsWordFinal() bool {
	return u.Next == nil
}

// IsWordInitial returns true if this is the first unit in the word.
func (u *Unit) IsWordInitial() bool {
	return u.Prev == nil
}

// String returns the original characters as a string.
func (u *Unit) String() string {
	return string(u.Runes)
}

// Word is the complete parsed representation of an input word.
type Word struct {
	Units    []*Unit // All parsed units in order
	Original string  // Input word in canonical form (see Canonicalizer)
	Options  Options // Transliteration options

	runes []rune // decoded Original, cached by Runes
}

// MaxLearnedWordRunes bounds the word length that learned per-word components
// (vowel/schwa models, candidate selectors) process; longer tokens fall back to
// the rules. Real words are far shorter (dataset maximum 53); the bound keeps
// pathological input from costing quadratic time.
const MaxLearnedWordRunes = 64

// Runes returns Original decoded as runes, computed once per word. Callers must
// not modify the returned slice, and Original must not change after the first
// call (only NewWord sets it).
func (w *Word) Runes() []rune {
	if w.runes == nil {
		w.runes = []rune(w.Original)
	}
	return w.runes
}

// NewWord creates a new empty Word.
func NewWord(original string) *Word {
	return &Word{
		Original: original,
	}
}

// AddUnit appends a unit and maintains bidirectional links.
func (w *Word) AddUnit(u *Unit) {
	if len(w.Units) > 0 {
		prev := w.Units[len(w.Units)-1]
		prev.Next = u
		u.Prev = prev
	}
	w.Units = append(w.Units, u)
}

// Options configures transliteration behavior.
type Options struct {
	// InherentVowelA renders a default-quality inherent vowel as "a" instead
	// of the script profile's default. Raised/open qualities are unchanged.
	// This renderer option does not rewrite explicit vowels or lexicon hits.
	// It has no effect on Hindi output; language support for other styles is
	// independent of this option (it is not a full academic scheme).
	InherentVowelA bool
	// LongVowels outputs "aa" for all ā (aa-matra) positions.
	LongVowels bool
	// SimpleNasals uses simplified nasal endings (करें→karen instead of karein).
	SimpleNasals bool
	// KeepMedialSchwa disables CCV schwa deletion, retaining schwa in more positions.
	// This produces output closer to some datasets (जनता→janata instead of janta).
	KeepMedialSchwa bool
	// SchwaModel uses the learned decision-tree schwa classifier for inherent-schwa
	// decisions instead of the hand-written heuristic schwa rules. Bengali uses
	// a three-class model on supported simple words and retains B1 rules elsewhere.
	// See lang/{hindi,bengali} and docs/reviews for held-out evaluations.
	SchwaModel bool
	// Lexicon consults the language's high-confidence romanization lexicon first;
	// known words return the attested human spelling, unknown words fall through
	// to the rule engine. Requires LexiconProvider or OptionsLexiconProvider;
	// the latter can decline lookup when the requested style is unsupported.
	Lexicon bool
	// Rerank enables the language's optional candidate selector (NativeReranker
	// or legacy Reranker). Bengali also requires SchwaModel and default style;
	// Hindi uses its character LM. A lexicon hit (with Lexicon) wins first.
	Rerank bool
	// Debug enables debug output showing rule applications.
	Debug bool
}

// DefaultStyle reports whether opts requests the default rendering style:
// every option other than the component switches (SchwaModel, Lexicon, Rerank)
// and Debug is unset. Components evaluated only for the default style (e.g.
// Bengali's lexicon and native selector) must decline otherwise. Any style
// option added later is non-default automatically.
func (o Options) DefaultStyle() bool {
	o.SchwaModel, o.Lexicon, o.Rerank, o.Debug = false, false, false, false
	return o == Options{}
}

// DefaultOptions returns the default transliteration options.
func DefaultOptions() Options {
	return Options{LongVowels: false, SimpleNasals: false, Debug: false}
}

// RuleTrace records a single rule application for debugging.
type RuleTrace struct {
	Phase    string // Phase name (Schwa, Consonant, Vowel, Render)
	Rule     string // Rule name
	Unit     string // Unit characters (e.g., "क")
	UnitIdx  int    // Unit index in word
	Before   string // BaseRom before rule
	After    string // BaseRom after rule (empty if unchanged)
	Metadata string // Additional info (e.g., schwa state)
}

// DebugInfo contains debugging information from transliteration.
type DebugInfo struct {
	Input  string      // Input in canonical form; unit RunePos indexes it
	Output string      // Final output
	Units  []UnitDebug // Parsed units
	Traces []RuleTrace // Rule applications
}

// UnitDebug contains debug info for a single unit.
type UnitDebug struct {
	Index    int    // Position in word
	Chars    string // Original characters
	Type     string // Unit type
	BaseRom  string // Base romanization
	RunePos  int    // Rune position in original string
	Metadata string // Script-specific info
}
