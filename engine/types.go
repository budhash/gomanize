// Package engine provides a multi-pass romanization engine for Indic scripts.
package engine

// Position tracks location in source text for debugging.
type Position struct {
	Offset int // Byte offset in original string
	Rune   int // Rune (character) index
}

// SchwaState tracks schwa deletion decisions for consonants.
type SchwaState int

const (
	SchwaPending SchwaState = iota // Not yet decided
	SchwaKeep                      // Definitely keep
	SchwaDelete                    // Definitely delete
)

func (s SchwaState) String() string {
	switch s {
	case SchwaPending:
		return "Pending"
	case SchwaKeep:
		return "Keep"
	case SchwaDelete:
		return "Delete"
	default:
		return "Unknown"
	}
}

// UnitType classifies parsed phonetic units.
type UnitType int

const (
	UnitVowel     UnitType = iota // Vowels and matras that replace inherent schwa
	UnitModifier                  // Modifiers that follow the vowel (anusvara, visarga, chandrabindu)
	UnitConsonant                 // Single consonant
	UnitConjunct                  // Multi-character conjunct (e.g., ज्ञ)
	UnitNumber                    // Devanagari numeral
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

	// Precomputed context
	AfterHalant bool // Was preceded by ्

	// State (for consonants/conjuncts)
	Schwa SchwaState

	// Bidirectional links
	Prev *Unit
	Next *Unit

	// Run membership (nil for vowels)
	Run      *ConsonantRun
	RunIndex int // Position within the run
}

// IsWordFinal returns true if this is the last unit in the word.
func (u *Unit) IsWordFinal() bool {
	return u.Next == nil
}

// IsWordInitial returns true if this is the first unit in the word.
func (u *Unit) IsWordInitial() bool {
	return u.Prev == nil
}

// String returns the original Devanagari characters as a string.
func (u *Unit) String() string {
	return string(u.Runes)
}

// ConsonantRun represents consecutive consonants between vowels.
// Used for coordinating schwa deletion decisions.
type ConsonantRun struct {
	Units     []*Unit // Consonants in this run
	PrevVowel *Unit   // Vowel before the run (nil if word-initial)
	NextVowel *Unit   // Vowel after the run (nil if word-final)
	DeletedAt int     // Index where schwa was deleted (-1 if none)
}

// NewConsonantRun creates a new run with DeletedAt initialized to -1.
func NewConsonantRun() *ConsonantRun {
	return &ConsonantRun{
		DeletedAt: -1,
	}
}

// HasDeletion returns true if a schwa has been deleted in this run.
func (r *ConsonantRun) HasDeletion() bool {
	return r.DeletedAt >= 0
}

// Word is the complete parsed representation of an input word.
type Word struct {
	Units    []*Unit         // All parsed units in order
	Runs     []*ConsonantRun // Consonant runs for schwa coordination
	Original string          // Original input string
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
