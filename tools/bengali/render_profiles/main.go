// Command render_profiles streams the Bengali B1, model, lexicon and
// model+lexicon profiles for JSON-encoded words read one per line.
// -disable takes comma-separated rule patterns to disable in every profile
// (rule ablations, without code edits). -drop-entry and -add-entry take
// comma-separated table:entry items (tables "coda" and "suf") and edit the
// module-internal rule tables for this process only (entry ablations).
// -list-tables prints the current tables as JSON and exits.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	gomanize "github.com/budhash/gomanize"
	"github.com/budhash/gomanize/internal/bengalitables"
)

// editTables applies table:entry items; an unknown table, a dropped entry that
// is absent or an added entry that is present is an error, never a no-op.
func editTables(items string, add bool) error {
	if items == "" {
		return nil
	}
	for _, item := range strings.Split(items, ",") {
		name, entry, ok := strings.Cut(item, ":")
		table := bengalitables.Table(name)
		if !ok || table == nil || entry == "" {
			return fmt.Errorf("bad table entry %q", item)
		}
		if table[entry] == add {
			return fmt.Errorf("table %s: entry %q already %s", name, entry, map[bool]string{true: "present", false: "absent"}[add])
		}
		if add {
			table[entry] = true
		} else {
			delete(table, entry)
		}
	}
	return nil
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func main() {
	disable := flag.String("disable", "", "comma-separated rule patterns to disable")
	drop := flag.String("drop-entry", "", "comma-separated table:entry items to remove (tables coda, suf)")
	add := flag.String("add-entry", "", "comma-separated table:entry items to add (tables coda, suf)")
	list := flag.Bool("list-tables", false, "print the rule tables as JSON and exit")
	flag.Parse()
	for _, edit := range []struct {
		items string
		add   bool
	}{{*drop, false}, {*add, true}} {
		if err := editTables(edit.items, edit.add); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if *list {
		if err := json.NewEncoder(os.Stdout).Encode(map[string][]string{"coda": sortedKeys(bengalitables.LoanCodas), "suf": sortedKeys(bengalitables.Suffixes)}); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	var disabled []string
	if *disable != "" {
		disabled = strings.Split(*disable, ",")
	}
	var engines []*gomanize.Gomanize
	for _, opts := range []gomanize.Options{{}, {SchwaModel: true}, {Lexicon: true}, {SchwaModel: true, Lexicon: true}} {
		g, err := gomanize.NewWithOptions("bengali", opts, gomanize.WithDisabledRules(disabled...))
		if err != nil {
			panic(err)
		}
		for _, pattern := range disabled {
			// A pattern that names no rule would silently ablate nothing.
			if len(g.ListRules(pattern)) == 0 {
				fmt.Fprintf(os.Stderr, "no rule matches %q\n", pattern)
				os.Exit(1)
			}
		}
		engines = append(engines, g)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	encoder := json.NewEncoder(os.Stdout)
	for scanner.Scan() {
		var native string
		if err := json.Unmarshal(scanner.Bytes(), &native); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		outputs := make([]string, len(engines))
		for i, g := range engines {
			outputs[i] = g.Translit(native)
		}
		if err := encoder.Encode(outputs); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
