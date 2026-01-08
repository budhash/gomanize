package core

import (
	"testing"
)

func TestRulePhaseString(t *testing.T) {
	tests := []struct {
		phase    RulePhase
		expected string
	}{
		{PhaseSchwa, "Schwa"},
		{PhaseConsonant, "Consonant"},
		{PhaseVowel, "Vowel"},
		{PhaseRender, "Render"},
		{RulePhase(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.phase.String(); got != tt.expected {
			t.Errorf("RulePhase(%d).String() = %q, want %q", tt.phase, got, tt.expected)
		}
	}
}

func TestRuleScopeString(t *testing.T) {
	tests := []struct {
		scope    RuleScope
		expected string
	}{
		{ScopeUniversal, "Universal"},
		{ScopeScript, "Script"},
		{ScopeLanguage, "Language"},
		{ScopeScheme, "Scheme"},
		{RuleScope(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.scope.String(); got != tt.expected {
			t.Errorf("RuleScope(%d).String() = %q, want %q", tt.scope, got, tt.expected)
		}
	}
}

func TestRuleModeString(t *testing.T) {
	tests := []struct {
		mode     RuleMode
		expected string
	}{
		{ModeExclusive, "Exclusive"},
		{ModeAlways, "Always"},
		{ModeFallback, "Fallback"},
		{RuleMode(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.expected {
			t.Errorf("RuleMode(%d).String() = %q, want %q", tt.mode, got, tt.expected)
		}
	}
}

func TestRuleEffectivePriority(t *testing.T) {
	tests := []struct {
		name     string
		scope    RuleScope
		priority int
		expected int
	}{
		{"Universal P0", ScopeUniversal, 0, 0},
		{"Universal P50", ScopeUniversal, 50, 50},
		{"Script P0", ScopeScript, 0, 100},
		{"Script P25", ScopeScript, 25, 125},
		{"Language P0", ScopeLanguage, 0, 200},
		{"Language P99", ScopeLanguage, 99, 299},
		{"Scheme P0", ScopeScheme, 0, 300},
		{"Scheme P50", ScopeScheme, 50, 350},
	}

	for _, tt := range tests {
		r := Rule{Name: tt.name, Scope: tt.scope, Priority: tt.priority}
		if got := r.EffectivePriority(); got != tt.expected {
			t.Errorf("%s: EffectivePriority() = %d, want %d", tt.name, got, tt.expected)
		}
	}
}

func TestNewRuleEnginePanicsOnNilCondition(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewRuleEngine should panic on nil Condition")
		}
	}()

	rules := []Rule{
		{Name: "test", Condition: nil, Action: func(*Unit, *Word) {}},
	}
	NewRuleEngine(rules)
}

func TestNewRuleEnginePanicsOnNilAction(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewRuleEngine should panic on nil Action")
		}
	}()

	rules := []Rule{
		{Name: "test", Condition: func(*Unit, *Word) bool { return true }, Action: nil},
	}
	NewRuleEngine(rules)
}

func TestNewRuleEnginePanicsOnInvalidPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority int
	}{
		{"negative", -1},
		{"too high", 100},
		{"way too high", 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("NewRuleEngine should panic on priority %d", tt.priority)
				}
			}()

			rules := []Rule{
				{
					Name:      "test",
					Priority:  tt.priority,
					Condition: func(*Unit, *Word) bool { return true },
					Action:    func(*Unit, *Word) {},
				},
			}
			NewRuleEngine(rules)
		})
	}
}

func TestNewRuleEngineValidRules(t *testing.T) {
	rules := []Rule{
		{
			Name:      "rule1",
			Phase:     PhaseSchwa,
			Scope:     ScopeLanguage,
			Priority:  50,
			Mode:      ModeExclusive,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "rule2",
			Phase:     PhaseConsonant,
			Scope:     ScopeScript,
			Priority:  25,
			Mode:      ModeAlways,
			Condition: func(*Unit, *Word) bool { return false },
			Action:    func(*Unit, *Word) {},
		},
	}

	engine := NewRuleEngine(rules)
	if engine == nil {
		t.Fatal("NewRuleEngine returned nil")
	}
	if len(engine.Rules()) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(engine.Rules()))
	}
}

func TestAddRuleValidation(t *testing.T) {
	engine := NewRuleEngine(nil)

	tests := []struct {
		name      string
		rule      Rule
		expectErr bool
	}{
		{
			name: "valid rule",
			rule: Rule{
				Name:      "valid",
				Priority:  50,
				Condition: func(*Unit, *Word) bool { return true },
				Action:    func(*Unit, *Word) {},
			},
			expectErr: false,
		},
		{
			name: "nil condition",
			rule: Rule{
				Name:      "invalid",
				Condition: nil,
				Action:    func(*Unit, *Word) {},
			},
			expectErr: true,
		},
		{
			name: "nil action",
			rule: Rule{
				Name:      "invalid",
				Condition: func(*Unit, *Word) bool { return true },
				Action:    nil,
			},
			expectErr: true,
		},
		{
			name: "priority too high",
			rule: Rule{
				Name:      "invalid",
				Priority:  100,
				Condition: func(*Unit, *Word) bool { return true },
				Action:    func(*Unit, *Word) {},
			},
			expectErr: true,
		},
		{
			name: "negative priority",
			rule: Rule{
				Name:      "invalid",
				Priority:  -1,
				Condition: func(*Unit, *Word) bool { return true },
				Action:    func(*Unit, *Word) {},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.AddRule(tt.rule)
			if (err != nil) != tt.expectErr {
				t.Errorf("AddRule() error = %v, expectErr = %v", err, tt.expectErr)
			}
		})
	}
}

func TestAddRulePriorityConflict(t *testing.T) {
	engine := NewRuleEngine(nil)

	rule1 := Rule{
		Name:      "rule1",
		Phase:     PhaseSchwa,
		Scope:     ScopeLanguage,
		Priority:  50,
		Condition: func(*Unit, *Word) bool { return true },
		Action:    func(*Unit, *Word) {},
	}
	rule2 := Rule{
		Name:      "rule2",
		Phase:     PhaseSchwa,
		Scope:     ScopeLanguage,
		Priority:  50,
		Condition: func(*Unit, *Word) bool { return true },
		Action:    func(*Unit, *Word) {},
	}

	if err := engine.AddRule(rule1); err != nil {
		t.Fatalf("First AddRule failed: %v", err)
	}

	if err := engine.AddRule(rule2); err == nil {
		t.Error("Second AddRule should fail due to priority conflict")
	}
}

func TestRuleEngineApply(t *testing.T) {
	var schwaRan, consonantRan, vowelRan, renderRan bool

	rules := []Rule{
		{
			Name:      "schwa-rule",
			Phase:     PhaseSchwa,
			Scope:     ScopeLanguage,
			Priority:  50,
			Mode:      ModeAlways,
			Condition: func(u *Unit, w *Word) bool { return u.Type == UnitConsonant },
			Action:    func(u *Unit, w *Word) { schwaRan = true },
		},
		{
			Name:      "consonant-rule",
			Phase:     PhaseConsonant,
			Scope:     ScopeLanguage,
			Priority:  50,
			Mode:      ModeAlways,
			Condition: func(u *Unit, w *Word) bool { return u.Type == UnitConsonant },
			Action:    func(u *Unit, w *Word) { consonantRan = true },
		},
		{
			Name:      "vowel-rule",
			Phase:     PhaseVowel,
			Scope:     ScopeLanguage,
			Priority:  50,
			Mode:      ModeAlways,
			Condition: func(u *Unit, w *Word) bool { return u.Type == UnitVowel },
			Action:    func(u *Unit, w *Word) { vowelRan = true },
		},
		{
			Name:      "render-rule",
			Phase:     PhaseRender,
			Scope:     ScopeLanguage,
			Priority:  50,
			Mode:      ModeAlways,
			Condition: func(u *Unit, w *Word) bool { return true },
			Action:    func(u *Unit, w *Word) { renderRan = true },
		},
	}

	engine := NewRuleEngine(rules)

	word := NewWord("test")
	word.AddUnit(&Unit{Type: UnitConsonant, BaseRom: "k"})
	word.AddUnit(&Unit{Type: UnitVowel, BaseRom: "a"})

	engine.Apply(word)

	if !schwaRan {
		t.Error("Schwa phase rule should have run")
	}
	if !consonantRan {
		t.Error("Consonant phase rule should have run")
	}
	if !vowelRan {
		t.Error("Vowel phase rule should have run")
	}
	if !renderRan {
		t.Error("Render phase rule should have run")
	}
}

func TestRuleEngineExclusiveMode(t *testing.T) {
	var rule1Ran, rule2Ran bool

	rules := []Rule{
		{
			Name:      "high-priority",
			Phase:     PhaseSchwa,
			Scope:     ScopeScheme,
			Priority:  50,
			Mode:      ModeExclusive,
			Condition: func(u *Unit, w *Word) bool { return true },
			Action:    func(u *Unit, w *Word) { rule1Ran = true },
		},
		{
			Name:      "low-priority",
			Phase:     PhaseSchwa,
			Scope:     ScopeUniversal,
			Priority:  50,
			Mode:      ModeExclusive,
			Condition: func(u *Unit, w *Word) bool { return true },
			Action:    func(u *Unit, w *Word) { rule2Ran = true },
		},
	}

	engine := NewRuleEngine(rules)

	word := NewWord("test")
	word.AddUnit(&Unit{Type: UnitConsonant, BaseRom: "k"})

	engine.Apply(word)

	if !rule1Ran {
		t.Error("High priority rule should have run")
	}
	if rule2Ran {
		t.Error("Low priority exclusive rule should NOT have run (unit already acted on)")
	}
}

func TestRuleEngineFallbackMode(t *testing.T) {
	var exclusiveRan, fallbackRan bool

	rules := []Rule{
		{
			Name:      "exclusive",
			Phase:     PhaseSchwa,
			Scope:     ScopeLanguage,
			Priority:  50,
			Mode:      ModeExclusive,
			Condition: func(u *Unit, w *Word) bool { return u.BaseRom == "k" },
			Action:    func(u *Unit, w *Word) { exclusiveRan = true },
		},
		{
			Name:      "fallback",
			Phase:     PhaseSchwa,
			Scope:     ScopeUniversal,
			Priority:  50,
			Mode:      ModeFallback,
			Condition: func(u *Unit, w *Word) bool { return true },
			Action:    func(u *Unit, w *Word) { fallbackRan = true },
		},
	}

	engine := NewRuleEngine(rules)

	word := NewWord("test")
	word.AddUnit(&Unit{Type: UnitConsonant, BaseRom: "k"}) // Matches exclusive
	word.AddUnit(&Unit{Type: UnitConsonant, BaseRom: "g"}) // Only matches fallback

	engine.Apply(word)

	if !exclusiveRan {
		t.Error("Exclusive rule should have run for 'k'")
	}
	if !fallbackRan {
		t.Error("Fallback rule should have run for 'g' (not acted on by exclusive)")
	}
}

func TestRuleEngineDebugTraces(t *testing.T) {
	rules := []Rule{
		{
			Name:      "test-rule",
			Phase:     PhaseSchwa,
			Scope:     ScopeLanguage,
			Priority:  50,
			Mode:      ModeAlways,
			Condition: func(u *Unit, w *Word) bool { return true },
			Action:    func(u *Unit, w *Word) { u.BaseRom = "changed" },
		},
	}

	engine := NewRuleEngine(rules)
	engine.EnableDebug(true)

	word := NewWord("test")
	word.AddUnit(&Unit{Type: UnitConsonant, BaseRom: "original", Runes: []rune("क")})

	engine.Apply(word)

	traces := engine.Traces()
	if len(traces) != 1 {
		t.Fatalf("Expected 1 trace, got %d", len(traces))
	}

	trace := traces[0]
	if trace.Rule != "test-rule" {
		t.Errorf("Trace.Rule = %q, want %q", trace.Rule, "test-rule")
	}
	if trace.Before != "original" {
		t.Errorf("Trace.Before = %q, want %q", trace.Before, "original")
	}
	if trace.After != "changed" {
		t.Errorf("Trace.After = %q, want %q", trace.After, "changed")
	}
	if trace.Phase != "Schwa" {
		t.Errorf("Trace.Phase = %q, want %q", trace.Phase, "Schwa")
	}
}

func TestRulesForPhase(t *testing.T) {
	rules := []Rule{
		{
			Name:      "schwa1",
			Phase:     PhaseSchwa,
			Scope:     ScopeLanguage,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "schwa2",
			Phase:     PhaseSchwa,
			Scope:     ScopeScript,
			Priority:  25,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "consonant1",
			Phase:     PhaseConsonant,
			Scope:     ScopeLanguage,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
	}

	engine := NewRuleEngine(rules)

	schwaRules := engine.RulesForPhase(PhaseSchwa)
	if len(schwaRules) != 2 {
		t.Errorf("Expected 2 schwa rules, got %d", len(schwaRules))
	}

	// Check sorted by priority (highest first)
	if schwaRules[0].Name != "schwa1" {
		t.Errorf("First schwa rule should be schwa1 (higher priority), got %q", schwaRules[0].Name)
	}

	consonantRules := engine.RulesForPhase(PhaseConsonant)
	if len(consonantRules) != 1 {
		t.Errorf("Expected 1 consonant rule, got %d", len(consonantRules))
	}

	vowelRules := engine.RulesForPhase(PhaseVowel)
	if len(vowelRules) != 0 {
		t.Errorf("Expected 0 vowel rules, got %d", len(vowelRules))
	}
}

// =============================================================================
// Rule Enable/Disable Tests
// =============================================================================

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		match   bool
	}{
		// Exact matches
		{"schwa.delete.ccv", "schwa.delete.ccv", true},
		{"schwa.delete.ccv", "schwa.delete.word-final", false},
		{"schwa.keep.default", "schwa.keep.default", true},

		// Wildcard * (matches all)
		{"schwa.delete.ccv", "*", true},
		{"anything.here", "*", true},

		// Prefix glob patterns
		{"schwa.delete.ccv", "schwa.*", true},
		{"schwa.delete.word-final", "schwa.*", true},
		{"schwa.keep.default", "schwa.*", true},
		{"consonant.va-to-wa.conjunct", "schwa.*", false},

		// More specific prefix patterns
		{"schwa.delete.ccv", "schwa.delete.*", true},
		{"schwa.delete.word-final", "schwa.delete.*", true},
		{"schwa.keep.default", "schwa.delete.*", false},

		// Edge cases
		{"schwa", "schwa", true},
		{"schwa", "schwa.*", false},
		{"schwa.delete", "schwa.delete.*", false},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.pattern, func(t *testing.T) {
			got := matchPattern(tt.name, tt.pattern)
			if got != tt.match {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.name, tt.pattern, got, tt.match)
			}
		})
	}
}

func TestDisableRuleExactMatch(t *testing.T) {
	rules := []Rule{
		{
			Name:      "schwa.delete.ccv",
			Phase:     PhaseSchwa,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "schwa.delete.word-final",
			Phase:     PhaseSchwa,
			Priority:  40,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
	}

	engine := NewRuleEngine(rules)

	// Both rules should be active initially
	if len(engine.RulesForPhase(PhaseSchwa)) != 2 {
		t.Fatalf("Expected 2 active rules initially")
	}

	// Disable one rule by exact name
	count := engine.DisableRule("schwa.delete.ccv")
	if count != 1 {
		t.Errorf("DisableRule() returned %d, want 1", count)
	}

	// Verify it's disabled
	if !engine.IsDisabled("schwa.delete.ccv") {
		t.Error("schwa.delete.ccv should be disabled")
	}
	if engine.IsDisabled("schwa.delete.word-final") {
		t.Error("schwa.delete.word-final should NOT be disabled")
	}

	// Verify active rules
	active := engine.RulesForPhase(PhaseSchwa)
	if len(active) != 1 {
		t.Errorf("Expected 1 active rule, got %d", len(active))
	}
	if active[0].Name != "schwa.delete.word-final" {
		t.Errorf("Remaining active rule should be schwa.delete.word-final, got %q", active[0].Name)
	}
}

func TestDisableRuleGlobPattern(t *testing.T) {
	rules := []Rule{
		{
			Name:      "schwa.delete.ccv",
			Phase:     PhaseSchwa,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "schwa.delete.word-final",
			Phase:     PhaseSchwa,
			Priority:  40,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "schwa.keep.default",
			Phase:     PhaseSchwa,
			Priority:  30,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "consonant.va-to-wa.conjunct",
			Phase:     PhaseConsonant,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
	}

	engine := NewRuleEngine(rules)

	// Disable all schwa.delete.* rules
	count := engine.DisableRule("schwa.delete.*")
	if count != 2 {
		t.Errorf("DisableRule(schwa.delete.*) returned %d, want 2", count)
	}

	// Verify both delete rules are disabled
	if !engine.IsDisabled("schwa.delete.ccv") {
		t.Error("schwa.delete.ccv should be disabled")
	}
	if !engine.IsDisabled("schwa.delete.word-final") {
		t.Error("schwa.delete.word-final should be disabled")
	}
	// Keep rule should not be disabled
	if engine.IsDisabled("schwa.keep.default") {
		t.Error("schwa.keep.default should NOT be disabled")
	}
	// Consonant rule should not be disabled
	if engine.IsDisabled("consonant.va-to-wa.conjunct") {
		t.Error("consonant.va-to-wa.conjunct should NOT be disabled")
	}

	// Verify active rules
	schwaRules := engine.RulesForPhase(PhaseSchwa)
	if len(schwaRules) != 1 {
		t.Errorf("Expected 1 active schwa rule, got %d", len(schwaRules))
	}
}

func TestDisableRuleAllWithWildcard(t *testing.T) {
	rules := []Rule{
		{
			Name:      "schwa.delete.ccv",
			Phase:     PhaseSchwa,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "consonant.va-to-wa.conjunct",
			Phase:     PhaseConsonant,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
	}

	engine := NewRuleEngine(rules)

	// Disable all rules with *
	count := engine.DisableRule("*")
	if count != 2 {
		t.Errorf("DisableRule(*) returned %d, want 2", count)
	}

	if len(engine.RulesForPhase(PhaseSchwa)) != 0 {
		t.Error("Expected 0 active schwa rules")
	}
	if len(engine.RulesForPhase(PhaseConsonant)) != 0 {
		t.Error("Expected 0 active consonant rules")
	}
}

func TestEnableRule(t *testing.T) {
	rules := []Rule{
		{
			Name:      "schwa.delete.ccv",
			Phase:     PhaseSchwa,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "schwa.delete.word-final",
			Phase:     PhaseSchwa,
			Priority:  40,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
	}

	engine := NewRuleEngine(rules)

	// Disable both
	engine.DisableRule("schwa.*")
	if len(engine.RulesForPhase(PhaseSchwa)) != 0 {
		t.Fatal("Expected 0 active rules after disabling")
	}

	// Enable one back
	count := engine.EnableRule("schwa.delete.ccv")
	if count != 1 {
		t.Errorf("EnableRule() returned %d, want 1", count)
	}

	// Verify it's enabled
	if engine.IsDisabled("schwa.delete.ccv") {
		t.Error("schwa.delete.ccv should be enabled")
	}

	// Verify active rules
	active := engine.RulesForPhase(PhaseSchwa)
	if len(active) != 1 {
		t.Errorf("Expected 1 active rule, got %d", len(active))
	}
}

func TestEnableRuleGlobPattern(t *testing.T) {
	rules := []Rule{
		{
			Name:      "schwa.delete.ccv",
			Phase:     PhaseSchwa,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "schwa.delete.word-final",
			Phase:     PhaseSchwa,
			Priority:  40,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "schwa.keep.default",
			Phase:     PhaseSchwa,
			Priority:  30,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
	}

	engine := NewRuleEngine(rules)

	// Disable all schwa rules
	engine.DisableRule("schwa.*")

	// Enable all schwa.delete.* rules
	count := engine.EnableRule("schwa.delete.*")
	if count != 2 {
		t.Errorf("EnableRule(schwa.delete.*) returned %d, want 2", count)
	}

	// Verify
	if engine.IsDisabled("schwa.delete.ccv") {
		t.Error("schwa.delete.ccv should be enabled")
	}
	if engine.IsDisabled("schwa.delete.word-final") {
		t.Error("schwa.delete.word-final should be enabled")
	}
	if !engine.IsDisabled("schwa.keep.default") {
		t.Error("schwa.keep.default should still be disabled")
	}

	active := engine.RulesForPhase(PhaseSchwa)
	if len(active) != 2 {
		t.Errorf("Expected 2 active rules, got %d", len(active))
	}
}

func TestEnableRuleNoOpWhenAlreadyEnabled(t *testing.T) {
	rules := []Rule{
		{
			Name:      "schwa.delete.ccv",
			Phase:     PhaseSchwa,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
	}

	engine := NewRuleEngine(rules)

	// Try to enable an already-enabled rule
	count := engine.EnableRule("schwa.delete.ccv")
	if count != 0 {
		t.Errorf("EnableRule() on already-enabled rule returned %d, want 0", count)
	}
}

func TestDisableRuleNoOpWhenAlreadyDisabled(t *testing.T) {
	rules := []Rule{
		{
			Name:      "schwa.delete.ccv",
			Phase:     PhaseSchwa,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
	}

	engine := NewRuleEngine(rules)

	// Disable once
	engine.DisableRule("schwa.delete.ccv")

	// Try to disable again
	count := engine.DisableRule("schwa.delete.ccv")
	if count != 0 {
		t.Errorf("DisableRule() on already-disabled rule returned %d, want 0", count)
	}
}

func TestListRulesAll(t *testing.T) {
	rules := []Rule{
		{
			Name:      "schwa.delete.ccv",
			Phase:     PhaseSchwa,
			Scope:     ScopeLanguage,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "consonant.va-to-wa.conjunct",
			Phase:     PhaseConsonant,
			Scope:     ScopeScript,
			Priority:  40,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
	}

	engine := NewRuleEngine(rules)

	// List all rules
	statuses := engine.ListRules("")
	if len(statuses) != 2 {
		t.Errorf("ListRules(\"\") returned %d rules, want 2", len(statuses))
	}

	// Verify status fields
	for _, s := range statuses {
		if !s.Enabled {
			t.Errorf("Rule %q should be enabled by default", s.Name)
		}
	}
}

func TestListRulesWithPattern(t *testing.T) {
	rules := []Rule{
		{
			Name:      "schwa.delete.ccv",
			Phase:     PhaseSchwa,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "schwa.delete.word-final",
			Phase:     PhaseSchwa,
			Priority:  40,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "consonant.va-to-wa.conjunct",
			Phase:     PhaseConsonant,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
	}

	engine := NewRuleEngine(rules)

	// List schwa.delete.* rules
	statuses := engine.ListRules("schwa.delete.*")
	if len(statuses) != 2 {
		t.Errorf("ListRules(schwa.delete.*) returned %d rules, want 2", len(statuses))
	}

	// List consonant.* rules
	statuses = engine.ListRules("consonant.*")
	if len(statuses) != 1 {
		t.Errorf("ListRules(consonant.*) returned %d rules, want 1", len(statuses))
	}
}

func TestListRulesShowsDisabledStatus(t *testing.T) {
	rules := []Rule{
		{
			Name:      "schwa.delete.ccv",
			Phase:     PhaseSchwa,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "schwa.keep.default",
			Phase:     PhaseSchwa,
			Priority:  30,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
	}

	engine := NewRuleEngine(rules)

	// Disable one rule
	engine.DisableRule("schwa.delete.ccv")

	statuses := engine.ListRules("")
	for _, s := range statuses {
		if s.Name == "schwa.delete.ccv" && s.Enabled {
			t.Error("schwa.delete.ccv should show as disabled")
		}
		if s.Name == "schwa.keep.default" && !s.Enabled {
			t.Error("schwa.keep.default should show as enabled")
		}
	}
}

func TestDisabledDefault(t *testing.T) {
	rules := []Rule{
		{
			Name:            "vowel.long-aa.all",
			Phase:           PhaseVowel,
			Priority:        50,
			DisabledDefault: true, // Disabled by default
			Condition:       func(*Unit, *Word) bool { return true },
			Action:          func(*Unit, *Word) {},
		},
		{
			Name:      "vowel.long-aa.closed-final",
			Phase:     PhaseVowel,
			Priority:  40,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
	}

	engine := NewRuleEngine(rules)

	// Rule with DisabledDefault should be disabled
	if !engine.IsDisabled("vowel.long-aa.all") {
		t.Error("vowel.long-aa.all should be disabled by default")
	}

	// Regular rule should be enabled
	if engine.IsDisabled("vowel.long-aa.closed-final") {
		t.Error("vowel.long-aa.closed-final should be enabled by default")
	}

	// Verify active rules
	active := engine.RulesForPhase(PhaseVowel)
	if len(active) != 1 {
		t.Errorf("Expected 1 active vowel rule, got %d", len(active))
	}
	if active[0].Name != "vowel.long-aa.closed-final" {
		t.Errorf("Expected active rule to be vowel.long-aa.closed-final, got %q", active[0].Name)
	}
}

func TestEnableDisabledDefaultRule(t *testing.T) {
	rules := []Rule{
		{
			Name:            "vowel.long-aa.all",
			Phase:           PhaseVowel,
			Priority:        50,
			DisabledDefault: true,
			Condition:       func(*Unit, *Word) bool { return true },
			Action:          func(*Unit, *Word) {},
		},
	}

	engine := NewRuleEngine(rules)

	// Initially disabled
	if !engine.IsDisabled("vowel.long-aa.all") {
		t.Fatal("Rule should be disabled by default")
	}

	// Enable it
	count := engine.EnableRule("vowel.long-aa.all")
	if count != 1 {
		t.Errorf("EnableRule() returned %d, want 1", count)
	}

	// Now it should be enabled
	if engine.IsDisabled("vowel.long-aa.all") {
		t.Error("Rule should be enabled after EnableRule()")
	}

	// And active
	active := engine.RulesForPhase(PhaseVowel)
	if len(active) != 1 {
		t.Errorf("Expected 1 active rule, got %d", len(active))
	}
}

func TestDisabledRulesNotApplied(t *testing.T) {
	var ruleRan bool

	rules := []Rule{
		{
			Name:      "test-rule",
			Phase:     PhaseSchwa,
			Priority:  50,
			Mode:      ModeAlways,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) { ruleRan = true },
		},
	}

	engine := NewRuleEngine(rules)

	// Disable the rule
	engine.DisableRule("test-rule")

	// Apply rules
	word := NewWord("test")
	word.AddUnit(&Unit{Type: UnitConsonant, BaseRom: "k"})
	engine.Apply(word)

	// Rule should not have run
	if ruleRan {
		t.Error("Disabled rule should not have run")
	}
}

func TestEnabledRulesAreApplied(t *testing.T) {
	var ruleRan bool

	rules := []Rule{
		{
			Name:            "test-rule",
			Phase:           PhaseSchwa,
			Priority:        50,
			Mode:            ModeAlways,
			DisabledDefault: true,
			Condition:       func(*Unit, *Word) bool { return true },
			Action:          func(*Unit, *Word) { ruleRan = true },
		},
	}

	engine := NewRuleEngine(rules)

	// Enable the rule (was disabled by default)
	engine.EnableRule("test-rule")

	// Apply rules
	word := NewWord("test")
	word.AddUnit(&Unit{Type: UnitConsonant, BaseRom: "k"})
	engine.Apply(word)

	// Rule should have run
	if !ruleRan {
		t.Error("Enabled rule should have run")
	}
}
