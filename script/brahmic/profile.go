package brahmic

import "github.com/budhash/gomanize/core"

// ScriptProfile supplies the script-specific constants used by the shared
// renderer and schwa rules. A nil Config.Profile retains Devanagari behavior.
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
		supplied := c.Profile
		if supplied.InherentVowel != "" {
			p.InherentVowel = supplied.InherentVowel
		}
		if supplied.RaisedVowel != "" {
			p.RaisedVowel = supplied.RaisedVowel
		}
		if supplied.BareVowelRunes != nil {
			p.BareVowelRunes = append([]rune{}, supplied.BareVowelRunes...)
		}
		if supplied.AaMatra != 0 {
			p.AaMatra = supplied.AaMatra
		}
		if supplied.SonorousRunes != nil {
			p.SonorousRunes = append([]rune{}, supplied.SonorousRunes...)
		}
		if supplied.IndependentVowelRange != [2]rune{} {
			p.IndependentVowelRange = supplied.IndependentVowelRange
		}
	}
	c.Profile = &p
	return c
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
