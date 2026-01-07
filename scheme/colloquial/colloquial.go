// Package colloquial provides a colloquial romanization scheme.
//
// Colloquial romanization aims to produce output that:
// - Matches how native speakers would casually write Hindi in Roman script
// - Is easy to read and pronounce for both Hindi speakers and non-speakers
// - Uses common spellings without diacritics (aa not ā, kh not ḵẖ)
//
// This scheme selects all standard rules from the language catalog.
package colloquial

import "github.com/budhash/gomanize/core"

// Colloquial implements the core.Scheme interface.
type Colloquial struct{}

// Name returns the scheme identifier.
func (c Colloquial) Name() string {
	return "colloquial"
}

// SelectRules selects rules from the language's catalog for colloquial output.
// This scheme uses all standard rules - it's the default scheme.
func (c Colloquial) SelectRules(catalog core.RuleCatalog) []core.Rule {
	// Colloquial uses all rules from the catalog
	return catalog.AllRules()
}
