// Command final_clusters reports, for JSON-encoded Bengali words read one per
// line, the word-final simple cluster (C + hasant + C, the left consonant not
// itself after a hasant) and the engine's decision on that final consonant's
// inherent vowel in the rules-only (B1) and vowel-model profiles. Each output
// line is [key, b1, model]: key joins the two letters without the hasant ("নট"
// for ন্ট) or is "" when the word does not end in a simple cluster; b1 and
// model are the final unit's schwa state (Keep, Delete). -coda replaces the
// loan-coda table (comma-separated keys) so the decision reflects a given
// engine, e.g. the table on main, rather than the table under selection.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/internal/bengalitables"
	"github.com/budhash/gomanize/lang/bengali"
	"github.com/budhash/gomanize/scheme/colloquial"
)

// singleLetter is one consonant letter, optionally with a nukta.
func singleLetter(chars string) bool {
	n := utf8.RuneCountInString(chars)
	return n == 1 || (n == 2 && strings.HasSuffix(chars, "়"))
}

func consonant(u core.UnitDebug) bool { return u.Type == core.UnitConsonant.String() }

func main() {
	coda := flag.String("coda", "", "comma-separated loan-coda keys replacing the compiled table")
	flag.Parse()
	if *coda != "" {
		for k := range bengalitables.LoanCodas {
			delete(bengalitables.LoanCodas, k)
		}
		for _, k := range strings.Split(*coda, ",") {
			bengalitables.LoanCodas[k] = true
		}
	}
	// Final units are outside the suffix rule's reach; disabling it keeps the
	// engine identical to the pre-suffix runtime for the decision reported.
	e := core.NewEngine(bengali.Bengali{}, colloquial.Colloquial{}, core.WithDisabledRules("schwa.bengali.suffix-boundary"))
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	encoder := json.NewEncoder(os.Stdout)
	for scanner.Scan() {
		var native string
		if err := json.Unmarshal(scanner.Bytes(), &native); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		out := []string{"", "", ""}
		for i, opts := range []core.Options{{}, {SchwaModel: true}} {
			_, info := e.TransliterateDebug(native, opts)
			units := info.Units
			if len(units) < 2 {
				continue
			}
			last, prev := units[len(units)-1], units[len(units)-2]
			if consonant(last) && consonant(prev) && singleLetter(last.Chars) && singleLetter(prev.Chars) &&
				strings.Contains(last.Metadata, "after-halant") && !strings.Contains(prev.Metadata, "after-halant") &&
				!strings.Contains(prev.Metadata, "no-inherent-vowel") && !strings.Contains(last.Metadata, "no-inherent-vowel") {
				out[0] = prev.Chars + last.Chars
				out[i+1] = strings.Fields(last.Metadata)[0]
			}
		}
		if err := encoder.Encode(out); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
