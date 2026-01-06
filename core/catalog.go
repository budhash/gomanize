package core

// RuleCatalog organizes rules by category.
// Languages provide complete catalogs; schemes select from them.
type RuleCatalog struct {
	Schwa     []Rule // All schwa-related rules
	Consonant []Rule // All consonant rules
	Vowel     []Rule // All vowel rules
	Render    []Rule // All output transform rules
}

// AllRules returns all rules in the catalog as a flat slice.
func (c RuleCatalog) AllRules() []Rule {
	var all []Rule
	all = append(all, c.Schwa...)
	all = append(all, c.Consonant...)
	all = append(all, c.Vowel...)
	all = append(all, c.Render...)
	return all
}

// FindByName finds a rule by name across all categories.
// Returns nil if not found.
func (c RuleCatalog) FindByName(name string) *Rule {
	for i := range c.Schwa {
		if c.Schwa[i].Name == name {
			return &c.Schwa[i]
		}
	}
	for i := range c.Consonant {
		if c.Consonant[i].Name == name {
			return &c.Consonant[i]
		}
	}
	for i := range c.Vowel {
		if c.Vowel[i].Name == name {
			return &c.Vowel[i]
		}
	}
	for i := range c.Render {
		if c.Render[i].Name == name {
			return &c.Render[i]
		}
	}
	return nil
}

// FindInSlice finds a rule by name in a slice of rules.
// Returns nil if not found.
func FindInSlice(rules []Rule, name string) *Rule {
	for i := range rules {
		if rules[i].Name == name {
			return &rules[i]
		}
	}
	return nil
}

// AppendIfFound appends a rule to the slice if found by name.
// Returns the (possibly extended) slice.
func AppendIfFound(rules []Rule, catalog []Rule, name string) []Rule {
	if r := FindInSlice(catalog, name); r != nil {
		return append(rules, *r)
	}
	return rules
}
