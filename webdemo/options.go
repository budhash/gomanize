// Package webdemo holds the host-testable pieces shared by the WebAssembly
// build of gomanize (cmd/gomanize-wasm). The syscall/js glue only compiles for
// GOOS=js, so anything worth unit-testing lives here where the normal test
// toolchain — and `make ci` — can reach it.
package webdemo

import "github.com/budhash/gomanize"

// Flag names accepted by Options. These are exactly the keys the browser UI
// puts in the options object it passes to gomanizeTranslit; keeping them as
// constants keeps the JS and Go sides from drifting and lets FlagNames expose
// the full set to both the glue and the tests.
const (
	FlagLongVowels      = "longVowels"
	FlagSimpleNasals    = "simpleNasals"
	FlagKeepMedialSchwa = "keepMedialSchwa"
	FlagSchwaModel      = "schwaModel"
	FlagLexicon         = "lexicon"
	FlagRerank          = "rerank"
)

// FlagNames returns every flag Options understands, in display order. The glue
// iterates this to read the JS options object, so a new flag is added in one
// place.
func FlagNames() []string {
	return []string{
		FlagLongVowels,
		FlagSimpleNasals,
		FlagKeepMedialSchwa,
		FlagSchwaModel,
		FlagLexicon,
		FlagRerank,
	}
}

// Options builds gomanize.Options from the demo's boolean UI flags. Absent or
// unknown keys default to false (the gomanize defaults). This is the single
// mapping point between the page's checkboxes and the engine options.
func Options(flags map[string]bool) gomanize.Options {
	o := gomanize.NewOptions()
	o.LongVowels = flags[FlagLongVowels]
	o.SimpleNasals = flags[FlagSimpleNasals]
	o.KeepMedialSchwa = flags[FlagKeepMedialSchwa]
	o.SchwaModel = flags[FlagSchwaModel]
	o.Lexicon = flags[FlagLexicon]
	o.Rerank = flags[FlagRerank]
	return o
}
