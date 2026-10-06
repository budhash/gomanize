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

// SchwaQuality separates the pronunciation of a retained inherent vowel from
// the keep/delete decision. Default preserves the configured script spelling.
type SchwaQuality int

const (
	SchwaDefault SchwaQuality = iota
	SchwaRaised
	SchwaOpen
)

// GeminationMode controls onset rendering without changing source identity or
// vowel ownership. None preserves historical rendering.
type GeminationMode int

const (
	GeminationNone GeminationMode = iota
	RepeatPrevious
	GeminateSelf
)

func (m GeminationMode) String() string {
	switch m {
	case GeminationNone:
		return "None"
	case RepeatPrevious:
		return "RepeatPrevious"
	case GeminateSelf:
		return "GeminateSelf"
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

	// IsMatra indicates a dependent vowel sign (matra), as opposed to an
	// independent vowel. Only a matra binds to the preceding consonant and
	// suppresses its inherent schwa; an independent vowel (e.g. ई in गई) starts
	// its own syllable, so the consonant keeps its schwa.
	IsMatra bool

	// Schwa state for consonants/conjuncts
	Schwa SchwaState

	// Quality of a retained inherent vowel; zero preserves existing behavior.
	SchwaQuality SchwaQuality

	// Gemination changes only consonant onsets, never the vowel decision.
	Gemination GeminationMode

	// NoInherentVowel marks an intrinsically dead consonant (e.g. khanda ta).
	// This structural property takes precedence over rules and vowel quality.
	NoInherentVowel bool

	// Nukta marks a letter parsed from base + nukta (decomposed form). It is
	// canonically one letter, equivalent to a precomposed nukta letter.
	Nukta bool

	// TrailingHalant marks the last unit of a word spelled with an explicit
	// final halant, which the parser otherwise consumes without a unit.
	TrailingHalant bool

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

// IsMatraUnit reports whether a unit is a dependent vowel sign (matra), as
// opposed to an independent vowel. Only a matra suppresses the preceding
// consonant's inherent schwa.
func IsMatraUnit(u *core.Unit) bool {
	bd := GetBrahmicData(u)
	return bd != nil && bd.IsMatra
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
	// Profile supplies script-specific rendering and shared-rule parameters.
	// Nil preserves the historical Devanagari defaults.
	Profile *ScriptProfile
	// VowellessConsonants lists intrinsically dead consonants. Nil preserves
	// the existing parser behavior. This does not change terminal virama handling.
	VowellessConsonants []rune
}

// Helper functions for working with BrahmicData through core.Unit

// IsAfterHalant reports a halant boundary, or an equivalent consonant boundary
// after a configured intrinsically dead consonant.
func IsAfterHalant(u *core.Unit) bool {
	bd := GetBrahmicData(u)
	if bd != nil && bd.AfterHalant {
		return true
	}
	// A consonant after an intrinsically dead consonant has the same cluster
	// boundary as after a halant for shared run lookaheads.
	if u != nil && IsConsonantOrConjunct(u) {
		prev := GetBrahmicData(u.Prev)
		return prev != nil && prev.NoInherentVowel
	}
	return false
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
	// Profile is resolved by the parser and retained by PrepareWord.
	// Treat it as read-only; a parser may share it between its words.
	Profile *ScriptProfile
	Runs    []*ConsonantRun
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
