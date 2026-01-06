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
