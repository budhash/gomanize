// Package brahmic provides types and logic for Brahmic script family transliteration.
// Brahmic scripts (Devanagari, Bengali, Tamil, etc.) share common features:
// - Inherent vowel (schwa) in consonants
// - Halant/virama to suppress inherent vowel
// - Consonant clusters via halant
// - Matras (dependent vowel signs)
package brahmic

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

// Unit represents a single phonetic unit in a parsed Brahmic script word.
// Extends the base engine.Unit concept with Brahmic-specific fields.
type Unit struct {
	// Source tracking
	Runes []rune // Original characters (1-3 for conjuncts)
	Start int    // Rune index start
	End   int    // Rune index end

	// Classification
	Type    UnitType
	BaseRom string // Base romanization (modifiable by rules)

	// Brahmic-specific context
	AfterHalant bool // Was preceded by halant (part of conjunct)

	// Schwa state (for consonants/conjuncts)
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

// String returns the original characters as a string.
func (u *Unit) String() string {
	return string(u.Runes)
}

// IsInConjunct returns true if this unit is part of a multi-consonant sequence.
func (u *Unit) IsInConjunct() bool {
	return u.AfterHalant
}

// IsRunInitial returns true if this is the first consonant in its run.
func (u *Unit) IsRunInitial() bool {
	return u.Run != nil && u.RunIndex == 0
}

// IsRunFinal returns true if this is the last consonant in its run.
func (u *Unit) IsRunFinal() bool {
	return u.Run != nil && u.RunIndex == len(u.Run.Units)-1
}

// PrevInRun returns the previous consonant in the same run, or nil.
func (u *Unit) PrevInRun() *Unit {
	if u.Run == nil || u.RunIndex == 0 {
		return nil
	}
	return u.Run.Units[u.RunIndex-1]
}

// NextInRun returns the next consonant in the same run, or nil.
func (u *Unit) NextInRun() *Unit {
	if u.Run == nil || u.RunIndex >= len(u.Run.Units)-1 {
		return nil
	}
	return u.Run.Units[u.RunIndex+1]
}

// UnitType classifies parsed phonetic units in Brahmic scripts.
type UnitType int

const (
	UnitVowel     UnitType = iota // Vowels and matras that replace inherent schwa
	UnitModifier                  // Modifiers that follow (anusvara, visarga, chandrabindu)
	UnitConsonant                 // Single consonant
	UnitConjunct                  // Multi-character conjunct (e.g., Devanagari ज्ञ)
	UnitNumber                    // Script numerals
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

// Word is the complete parsed representation of a Brahmic script input word.
type Word struct {
	Units    []*Unit         // All parsed units in order
	Runs     []*ConsonantRun // Consonant runs for schwa coordination
	Original string          // Original input string
	Options  Options         // Transliteration options
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
	// LongVowels outputs "aa" for all aa-matra positions.
	LongVowels bool
}

// DefaultOptions returns the default transliteration options.
func DefaultOptions() Options {
	return Options{LongVowels: false}
}
