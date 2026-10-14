// Command render_profiles streams the Bengali B1, model, lexicon and
// model+lexicon profiles for JSON-encoded words read one per line.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	gomanize "github.com/budhash/gomanize"
)

func main() {
	var engines []*gomanize.Gomanize
	for _, opts := range []gomanize.Options{{}, {SchwaModel: true}, {Lexicon: true}, {SchwaModel: true, Lexicon: true}} {
		g, err := gomanize.NewWithOptions("bengali", opts)
		if err != nil {
			panic(err)
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
