package core

import (
	"fmt"
	"sort"
	"strings"
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
	ScopeScript                     // Base 100: script-specific (Brahmic)
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
	Name            string
	Phase           RulePhase
	Scope           RuleScope
	Priority        int // 0-99 within scope
	Mode            RuleMode
	DisabledDefault bool // If true, rule is disabled by default (must be explicitly enabled)
	Condition       func(*Unit, *Word) bool
	Action          func(*Unit, *Word)
}

// RuleStatus represents a rule and its current enabled/disabled state.
type RuleStatus struct {
	Name     string
	Phase    RulePhase
	Scope    RuleScope
	Priority int
	Enabled  bool
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
	disabled map[string]bool      // Disabled rule names

	// Debug support
	traces       []RuleTrace
	debugEnabled bool
	debugMeta    func(*Unit) string // Script-specific metadata extractor
}

// NewRuleEngine creates a new rule engine with the given rules.
// Panics if any rule has nil Condition or Action functions.
func NewRuleEngine(rules []Rule) *RuleEngine {
	// Validate all rules have required functions
	for i, r := range rules {
		if r.Condition == nil {
			panic(fmt.Sprintf("core.NewRuleEngine: rule %d (%q) has nil Condition", i, r.Name))
		}
		if r.Action == nil {
			panic(fmt.Sprintf("core.NewRuleEngine: rule %d (%q) has nil Action", i, r.Name))
		}
		if r.Priority < 0 || r.Priority > 99 {
			panic(fmt.Sprintf("core.NewRuleEngine: rule %d (%q) has invalid Priority %d (must be 0-99)", i, r.Name, r.Priority))
		}
	}

	e := &RuleEngine{
		allRules: append([]Rule{}, rules...),
		active:   make(map[RulePhase][]Rule),
		disabled: make(map[string]bool),
	}

	// Apply default enabled/disabled states
	// Rules with DisabledDefault=true are disabled by default
	for _, r := range e.allRules {
		if r.DisabledDefault {
			e.disabled[r.Name] = true
		}
	}

	e.rebuildActive()
	return e
}

// AddRule adds a rule to the engine.
// Returns error if there's a priority conflict or invalid rule.
func (e *RuleEngine) AddRule(r Rule) error {
	// Validate rule
	if r.Condition == nil {
		return fmt.Errorf("rule %q has nil Condition", r.Name)
	}
	if r.Action == nil {
		return fmt.Errorf("rule %q has nil Action", r.Name)
	}
	if r.Priority < 0 || r.Priority > 99 {
		return fmt.Errorf("rule %q has invalid Priority %d (must be 0-99)", r.Name, r.Priority)
	}

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

// rebuildActive sorts rules by priority for each phase, excluding disabled rules.
func (e *RuleEngine) rebuildActive() {
	e.active = make(map[RulePhase][]Rule)
	for _, r := range e.allRules {
		// Skip disabled rules
		if e.disabled[r.Name] {
			continue
		}
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
	for i := range rules {
		rule := &rules[i]
		if rule.Mode == ModeFallback {
			continue
		}
		for idx, unit := range word.Units {
			// Skip if already acted on and this is Exclusive mode
			if rule.Mode == ModeExclusive && acted[unit] {
				continue
			}
			if rule.Condition(unit, word) {
				before := unit.BaseRom
				rule.Action(unit, word)
				acted[unit] = true
				e.traceRule(phase, rule, unit, idx, before)
			}
		}
	}

	// Second pass: Fallback rules (only for units not acted on)
	for i := range rules {
		rule := &rules[i]
		if rule.Mode != ModeFallback {
			continue
		}
		for idx, unit := range word.Units {
			if acted[unit] {
				continue
			}
			if rule.Condition(unit, word) {
				before := unit.BaseRom
				rule.Action(unit, word)
				acted[unit] = true
				e.traceRule(phase, rule, unit, idx, before)
			}
		}
	}
}

// Rules returns a copy of all registered rules.
func (e *RuleEngine) Rules() []Rule {
	return append([]Rule{}, e.allRules...)
}

// RulesForPhase returns a copy of rules for a specific phase, sorted by priority.
func (e *RuleEngine) RulesForPhase(phase RulePhase) []Rule {
	return append([]Rule{}, e.active[phase]...)
}

// DisableRule disables rules matching the given pattern.
// Pattern can be an exact name or a glob pattern using '*' as wildcard.
// Examples: "schwa.delete.ccv", "schwa.*", "schwa.delete.*"
// Returns the count of rules matched and disabled.
func (e *RuleEngine) DisableRule(pattern string) int {
	count := 0
	for _, r := range e.allRules {
		if matchPattern(r.Name, pattern) && !e.disabled[r.Name] {
			e.disabled[r.Name] = true
			count++
		}
	}
	if count > 0 {
		e.rebuildActive()
	}
	return count
}

// EnableRule enables rules matching the given pattern.
// Pattern can be an exact name or a glob pattern using '*' as wildcard.
// Examples: "vowel.long-aa.all", "vowel.*", "vowel.long-aa.*"
// Returns the count of rules matched and enabled.
func (e *RuleEngine) EnableRule(pattern string) int {
	count := 0
	for _, r := range e.allRules {
		if matchPattern(r.Name, pattern) && e.disabled[r.Name] {
			delete(e.disabled, r.Name)
			count++
		}
	}
	if count > 0 {
		e.rebuildActive()
	}
	return count
}

// IsDisabled returns true if the rule with the given name is currently disabled.
func (e *RuleEngine) IsDisabled(name string) bool {
	return e.disabled[name]
}

// ListRules returns all rules with their enabled/disabled status.
// If pattern is non-empty, only rules matching the pattern are returned.
func (e *RuleEngine) ListRules(pattern string) []RuleStatus {
	var result []RuleStatus
	for _, r := range e.allRules {
		if pattern != "" && !matchPattern(r.Name, pattern) {
			continue
		}
		result = append(result, RuleStatus{
			Name:     r.Name,
			Phase:    r.Phase,
			Scope:    r.Scope,
			Priority: r.Priority,
			Enabled:  !e.disabled[r.Name],
		})
	}
	return result
}

// matchPattern checks if name matches the given pattern.
// Pattern supports '*' as a wildcard that matches any suffix.
// Examples:
//   - "schwa.delete.ccv" matches "schwa.delete.ccv" (exact)
//   - "schwa.*" matches "schwa.delete.ccv", "schwa.keep.default"
//   - "schwa.delete.*" matches "schwa.delete.ccv", "schwa.delete.word-final"
func matchPattern(name, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(name, prefix)
	}
	return name == pattern
}

// SetDebugMetaExtractor sets a function to extract script-specific metadata.
func (e *RuleEngine) SetDebugMetaExtractor(fn func(*Unit) string) {
	e.debugMeta = fn
}

// EnableDebug enables debug trace collection.
func (e *RuleEngine) EnableDebug(enabled bool) {
	e.debugEnabled = enabled
	if enabled {
		e.traces = nil // Reset traces
	}
}

// Traces returns collected debug traces (call after Apply).
func (e *RuleEngine) Traces() []RuleTrace {
	return e.traces
}

// traceRule records a rule application if debugging is enabled.
func (e *RuleEngine) traceRule(phase RulePhase, rule *Rule, unit *Unit, unitIdx int, before string) {
	if !e.debugEnabled {
		return
	}
	meta := ""
	if e.debugMeta != nil {
		meta = e.debugMeta(unit)
	}
	trace := RuleTrace{
		Phase:    phase.String(),
		Rule:     rule.Name,
		Unit:     string(unit.Runes),
		UnitIdx:  unitIdx,
		Before:   before,
		After:    unit.BaseRom,
		Metadata: meta,
	}
	// Only record if something changed or it's a schwa rule
	if before != unit.BaseRom || phase == PhaseSchwa {
		e.traces = append(e.traces, trace)
	}
}
