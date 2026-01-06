package engine

import (
	"fmt"
	"sort"
)

// RulePhase determines when a rule executes in the pipeline.
type RulePhase int

const (
	PhaseSchwa     RulePhase = iota // Schwa deletion/retention decisions
	PhaseConsonant                  // Consonant modifications (e.g., व→w)
	PhaseVowel                      // Vowel modifications (e.g., aa→ā)
	PhaseRender                     // Final output adjustments
)

func (p RulePhase) String() string {
	switch p {
	case PhaseSchwa:
		return "Schwa"
	case PhaseConsonant:
		return "Consonant"
	case PhaseVowel:
		return "Vowel"
	case PhaseRender:
		return "Render"
	default:
		return "Unknown"
	}
}

// RuleScope determines the priority tier for a rule.
// Higher scopes have higher effective priority.
type RuleScope int

const (
	ScopeUniversal RuleScope = iota // Base 0: applies to all languages
	ScopeScript                     // Base 100: script-specific (Devanagari)
	ScopeLanguage                   // Base 200: language-specific (Hindi)
	ScopeScheme                     // Base 300: scheme-specific (IAST)
)

func (s RuleScope) String() string {
	switch s {
	case ScopeUniversal:
		return "Universal"
	case ScopeScript:
		return "Script"
	case ScopeLanguage:
		return "Language"
	case ScopeScheme:
		return "Scheme"
	default:
		return "Unknown"
	}
}

// RuleMode determines execution behavior.
type RuleMode int

const (
	ModeExclusive RuleMode = iota // First match wins for this unit
	ModeAlways                    // Always run if condition matches
	ModeFallback                  // Only if no other rule acted on unit
)

func (m RuleMode) String() string {
	switch m {
	case ModeExclusive:
		return "Exclusive"
	case ModeAlways:
		return "Always"
	case ModeFallback:
		return "Fallback"
	default:
		return "Unknown"
	}
}

// Rule represents a single transliteration rule.
type Rule struct {
	Name      string
	Phase     RulePhase
	Scope     RuleScope
	Priority  int // 0-99 within scope
	Mode      RuleMode
	Condition func(*Unit, *Word) bool
	Action    func(*Unit, *Word)
}

// EffectivePriority calculates the overall priority across scopes.
// Higher values run first.
func (r *Rule) EffectivePriority() int {
	return int(r.Scope)*100 + r.Priority
}

// RuleEngine applies rules to parsed words.
type RuleEngine struct {
	allRules []Rule
	active   map[RulePhase][]Rule // Filtered and sorted per phase
}

// NewRuleEngine creates a new empty rule engine.
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		active: make(map[RulePhase][]Rule),
	}
}

// AddRule adds a rule to the engine.
// Returns error if there's a priority conflict.
func (e *RuleEngine) AddRule(r Rule) error {
	// Check for priority conflicts within same phase
	for _, existing := range e.allRules {
		if existing.Phase == r.Phase &&
			existing.EffectivePriority() == r.EffectivePriority() {
			return fmt.Errorf("priority conflict: rules %q and %q both have effective priority %d in phase %s",
				existing.Name, r.Name, r.EffectivePriority(), r.Phase)
		}
	}
	e.allRules = append(e.allRules, r)
	e.rebuildActive()
	return nil
}

// AddRules adds multiple rules to the engine.
func (e *RuleEngine) AddRules(rules []Rule) error {
	for _, r := range rules {
		if err := e.AddRule(r); err != nil {
			return err
		}
	}
	return nil
}

// rebuildActive sorts rules by priority for each phase.
func (e *RuleEngine) rebuildActive() {
	e.active = make(map[RulePhase][]Rule)
	for _, r := range e.allRules {
		e.active[r.Phase] = append(e.active[r.Phase], r)
	}
	// Sort each phase by effective priority (highest first)
	for phase := range e.active {
		sort.Slice(e.active[phase], func(i, j int) bool {
			return e.active[phase][i].EffectivePriority() > e.active[phase][j].EffectivePriority()
		})
	}
}

// Apply executes all rules on the word in phase order.
func (e *RuleEngine) Apply(word *Word) {
	phases := []RulePhase{PhaseSchwa, PhaseConsonant, PhaseVowel, PhaseRender}

	for _, phase := range phases {
		e.applyPhase(phase, word)
	}
}

// applyPhase executes rules for a single phase.
func (e *RuleEngine) applyPhase(phase RulePhase, word *Word) {
	rules := e.active[phase]
	if len(rules) == 0 {
		return
	}

	// Track which units have been acted on (for Exclusive mode)
	acted := make(map[*Unit]bool)

	// First pass: Exclusive and Always rules (highest priority first)
	for _, rule := range rules {
		if rule.Mode == ModeFallback {
			continue
		}
		for _, unit := range word.Units {
			// Skip if already acted on and this is Exclusive mode
			if rule.Mode == ModeExclusive && acted[unit] {
				continue
			}
			if rule.Condition(unit, word) {
				rule.Action(unit, word)
				acted[unit] = true
			}
		}
	}

	// Second pass: Fallback rules (only for units not acted on)
	for _, rule := range rules {
		if rule.Mode != ModeFallback {
			continue
		}
		for _, unit := range word.Units {
			if acted[unit] {
				continue
			}
			if rule.Condition(unit, word) {
				rule.Action(unit, word)
				acted[unit] = true
			}
		}
	}
}

// Rules returns all registered rules.
func (e *RuleEngine) Rules() []Rule {
	return e.allRules
}

// RulesForPhase returns rules for a specific phase, sorted by priority.
func (e *RuleEngine) RulesForPhase(phase RulePhase) []Rule {
	return e.active[phase]
}
