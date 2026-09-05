// Types and state for Brahmic script family transliteration.
// Brahmic scripts (Devanagari, Bengali, Tamil, etc.) share common features:
// - Inherent vowel (schwa) in consonants
// - Halant/virama to suppress inherent vowel
// - Consonant clusters via halant
// - Matras (dependent vowel signs)

package brahmic

import "github.com/budhash/gomanize/core"

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
	Units     []*core.Unit // Consonants in this run
	PrevVowel *core.Unit   // Vowel before the run (nil if word-initial)
	NextVowel *core.Unit   // Vowel after the run (nil if word-final)
	DeletedAt int          // Index where schwa was deleted (-1 if none)
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

// BrahmicData holds Brahmic-specific data for a core.Unit.
// Stored in Unit.ScriptData.
type BrahmicData struct {
	// AfterHalant indicates this unit followed a halant (part of conjunct)
	AfterHalant bool

	// Schwa state for consonants/conjuncts
	Schwa SchwaState

	// Run membership (nil for vowels)
	Run      *ConsonantRun
	RunIndex int // Position within the run

	// WordData holds word-level data (only on first unit)
	WordData *WordBrahmicData
}

// GetBrahmicData extracts BrahmicData from a core.Unit.
// Returns nil if ScriptData is not BrahmicData.
func GetBrahmicData(u *core.Unit) *BrahmicData {
	if u == nil || u.ScriptData == nil {
		return nil
	}
	if bd, ok := u.ScriptData.(*BrahmicData); ok {
		return bd
	}
	return nil
}

// SetBrahmicData sets BrahmicData on a core.Unit.
func SetBrahmicData(u *core.Unit, bd *BrahmicData) {
	u.ScriptData = bd
}

// NewBrahmicData creates BrahmicData with default values.
func NewBrahmicData() *BrahmicData {
	return &BrahmicData{
		Schwa: SchwaPending,
	}
}

// Config holds Brahmic script configuration.
type Config struct {
	Halant    string   // Halant/virama character (e.g., "्" for Devanagari)
	Nukta     string   // Nukta character (e.g., "़" for Devanagari)
	MultiChar []string // Multi-character sequences to match first (e.g., "ज्ञ")
}

// Helper functions for working with BrahmicData through core.Unit

// IsAfterHalant returns true if the unit followed a halant.
func IsAfterHalant(u *core.Unit) bool {
	bd := GetBrahmicData(u)
	return bd != nil && bd.AfterHalant
}

// GetSchwa returns the schwa state for a unit.
func GetSchwa(u *core.Unit) SchwaState {
	bd := GetBrahmicData(u)
	if bd == nil {
		return SchwaPending
	}
	return bd.Schwa
}

// SetSchwa sets the schwa state for a unit.
func SetSchwa(u *core.Unit, state SchwaState) {
	bd := GetBrahmicData(u)
	if bd != nil {
		bd.Schwa = state
	}
}

// GetRun returns the consonant run for a unit.
func GetRun(u *core.Unit) *ConsonantRun {
	bd := GetBrahmicData(u)
	if bd == nil {
		return nil
	}
	return bd.Run
}

// GetRunIndex returns the run index for a unit.
func GetRunIndex(u *core.Unit) int {
	bd := GetBrahmicData(u)
	if bd == nil {
		return -1
	}
	return bd.RunIndex
}

// IsConsonantOrConjunct returns true if the unit is a consonant or conjunct.
func IsConsonantOrConjunct(u *core.Unit) bool {
	return u.Type == core.UnitConsonant || u.Type == core.UnitConjunct
}

// WordBrahmicData holds Brahmic-specific data for a core.Word.
// Stored in the first unit's BrahmicData.WordData field.
type WordBrahmicData struct {
	Runs []*ConsonantRun
}

// GetWordBrahmicData retrieves word-level Brahmic data.
// We store this in the word's first unit's ScriptData as a map.
func GetWordBrahmicData(w *core.Word) *WordBrahmicData {
	if len(w.Units) == 0 {
		return nil
	}
	// Check if first unit has a map with word data
	bd := GetBrahmicData(w.Units[0])
	if bd == nil {
		return nil
	}
	// Word data is stored in the first unit's BrahmicData.WordData
	return bd.WordData
}

// SetWordBrahmicData sets word-level Brahmic data.
func SetWordBrahmicData(w *core.Word, wd *WordBrahmicData) {
	if len(w.Units) == 0 {
		return
	}
	bd := GetBrahmicData(w.Units[0])
	if bd == nil {
		bd = NewBrahmicData()
		SetBrahmicData(w.Units[0], bd)
	}
	bd.WordData = wd
}
