//go:build bengali_vowel_v1

package bengali

import _ "embed"

// Evaluation-only build (go run -tags bengali_vowel_v1): swaps in the
// simple-word vowel model shipped before T-0121 (main at 46c2967), so the
// conjunct retrain can be scored against it paired, word for word. Never part
// of a release build.
//
//go:embed testdata/vowel_tree_simple_v1.json
var vowelTreeSimpleV1 []byte

func init() { vowelTreeJSON = vowelTreeSimpleV1 }
