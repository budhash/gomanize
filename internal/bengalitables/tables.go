// Package bengalitables holds the entry tables behind the Bengali table-driven
// rules. It is module-internal so that evaluation tools in this repository can
// ablate single entries (tools/bengali/render_profiles -drop-entry) without a
// test-only knob in the public API. Library code must treat the tables as
// read-only.
package bengalitables

// LoanCodas lists word-final simple clusters (C + hasant + C) whose final
// inherent vowel is deleted (schwa.bengali.loan-coda). Keys join the two
// letters without the hasant: "নট" is ন্ট. Each entry passed a train-only
// leave-one-out ablation (docs/reviews/2026-10-20-bengali-suffix-coda.md);
// ন্ড ঞ্জ ক্স (2026-10-14) failed it and were removed.
var LoanCodas = map[string]bool{
	"কট": true, "নট": true, "নস": true, "মপ": true, "রক": true,
	"রচ": true, "রস": true, "লট": true, "লড": true, "সক": true,
}

// Suffixes lists word-final nominal suffixes before which a stem-final bare
// consonant loses its inherent vowel (schwa.bengali.suffix-boundary). The
// longest matching suffix wins. These are the survivors of the train-only
// selection over 19 preregistered candidates (same record); কে টি তে দের রা
// and the others failed the loss bound, mostly under the vowel model.
var Suffixes = map[string]bool{"খানা": true, "টাকে": true, "টিকে": true}

// Table returns the named table ("coda" or "suf") or nil.
func Table(name string) map[string]bool {
	switch name {
	case "coda":
		return LoanCodas
	case "suf":
		return Suffixes
	}
	return nil
}
