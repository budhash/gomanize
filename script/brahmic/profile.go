package brahmic

import (
	"fmt"
	"strings"

	"github.com/budhash/gomanize/core"
)

// ScriptProfile supplies the script-specific constants used by the shared
// renderer and schwa rules. A nil Config.Profile retains Devanagari behavior;
// a non-nil profile must set every field (NewParser panics otherwise).
// NewParser copies the profile and its slices; treat a parsed word's profile
// as read-only. An explicitly empty rune slice disables that membership check.
type ScriptProfile struct {
	InherentVowel         string
	RaisedVowel           string
	BareVowelRunes        []rune
	AaMatra               rune
	SonorousRunes         []rune
	IndependentVowelRange [2]rune
}

// DevanagariProfile returns an independent copy of the historical shared-layer
// constants. Hindi supplies it explicitly; legacy callers receive it by default.
func DevanagariProfile() ScriptProfile {
	return ScriptProfile{
		InherentVowel: "a", RaisedVowel: "o",
		BareVowelRunes:        []rune{0x0905, 0x0904},
		AaMatra:               0x093E,
		SonorousRunes:         []rune{'र', 'य', 'व'},
		IndependentVowelRange: [2]rune{0x0905, 0x0914},
	}
}

var defaultProfile = DevanagariProfile()

func (c Config) normalize() Config {
	p := DevanagariProfile()
	if c.Profile != nil {
		// A supplied profile must be complete: silently filling gaps with
		// Devanagari values would turn a new script's checks into no-ops.
		if missing := c.Profile.missingFields(); len(missing) > 0 {
			panic(fmt.Sprintf("brahmic: incomplete ScriptProfile, unset fields: %s", strings.Join(missing, ", ")))
		}
		p = *c.Profile
		p.BareVowelRunes = append([]rune{}, c.Profile.BareVowelRunes...)
		p.SonorousRunes = append([]rune{}, c.Profile.SonorousRunes...)
	}
	c.Profile = &p
	return c
}

// missingFields lists unset profile fields. Nil rune slices are unset; an
// explicitly empty slice disables that membership check.
func (p *ScriptProfile) missingFields() []string {
	var missing []string
	if p.InherentVowel == "" {
		missing = append(missing, "InherentVowel")
	}
	if p.RaisedVowel == "" {
		missing = append(missing, "RaisedVowel")
	}
	if p.BareVowelRunes == nil {
		missing = append(missing, "BareVowelRunes")
	}
	if p.AaMatra == 0 {
		missing = append(missing, "AaMatra")
	}
	if p.SonorousRunes == nil {
		missing = append(missing, "SonorousRunes")
	}
	if p.IndependentVowelRange == [2]rune{} {
		missing = append(missing, "IndependentVowelRange")
	}
	return missing
}

func profileFor(w *core.Word) *ScriptProfile {
	if wd := GetWordBrahmicData(w); wd != nil && wd.Profile != nil {
		return wd.Profile
	}
	return &defaultProfile
}

func containsRune(rs []rune, r rune) bool {
	for _, candidate := range rs {
		if candidate == r {
			return true
		}
	}
	return false
}
