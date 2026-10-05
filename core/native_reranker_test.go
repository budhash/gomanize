package core

import "testing"

type nativeTestReranker struct {
	*mockLanguage
	calls, legacyCalls int
	hit, decline       bool
	seen               Options
}

func (l *nativeTestReranker) LexiconLookup(string) (string, bool) { return "lexicon", l.hit }
func (l *nativeTestReranker) RerankRomans([]string) string        { l.legacyCalls++; return "legacy" }
func (l *nativeTestReranker) RerankNative(input string, opts Options, render CandidateRenderer) (string, bool) {
	l.calls++
	l.seen = opts
	if l.decline {
		return "", false
	}
	// A newly available lexicon hit must not short-circuit candidate rendering.
	l.hit = true
	return render(l, opts), true
}

func TestNativeRerankerContract(t *testing.T) {
	var candidate Options
	rule := Rule{Name: "test.mutate", Phase: PhaseConsonant, Scope: ScopeLanguage, Priority: 1, DisabledDefault: true,
		Condition: func(*Unit, *Word) bool { return true }, Action: func(u *Unit, w *Word) { candidate = w.Options; u.BaseRom = "x" }}
	language := &nativeTestReranker{mockLanguage: &mockLanguage{script: &mockScript{parser: &mockParser{}, renderer: &mockRenderer{}}}}
	scheme := &mockScheme{rules: []Rule{rule}}
	e := NewEngine(language, scheme)
	opts := Options{Rerank: true, SchwaModel: true, Lexicon: true, Debug: true, LongVowels: true}
	e.RuleEngine().EnableRule(rule.Name)
	got, info := e.TransliterateDebug("a", opts)
	if got != "x" || info != nil || language.seen != opts {
		t.Fatalf("result/options/debug: %q %+v %+v", got, info, language.seen)
	}
	want := opts
	want.Rerank = false
	want.Lexicon = false
	want.Debug = false
	if candidate != want || language.calls != 1 || language.legacyCalls != 0 {
		t.Fatalf("candidate contract: %+v %+v", candidate, language)
	}
	// Runtime rule toggles and constructor disables survive candidate engine creation.
	language.hit = false
	e.RuleEngine().DisableRule(rule.Name)
	if got = e.TransliterateWithOptions("a", opts); got != "a" {
		t.Fatal(got)
	}
	if e.TransliterateWithOptions("a", opts) != "lexicon" || language.calls != 2 {
		t.Fatal("lexicon did not win first")
	}
	language.hit = false
	language.decline = true
	if got = e.TransliterateWithOptions("a", opts); got != "a" || language.legacyCalls != 0 {
		t.Fatal("decline fell through to legacy reranker")
	}
	language.decline = false
	language.hit = false
	e = NewEngine(language, &mockScheme{})
	if got = e.TransliterateWithOptions("a", opts); got != "a" {
		t.Fatal("candidate ignored scheme filtering")
	}
}
