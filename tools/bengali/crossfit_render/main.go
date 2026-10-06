// Command crossfit_render renders offline vowel decisions with the production
// rule order. It provides no runtime model-loading or public API extension.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	gomanize "github.com/budhash/gomanize"
	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/lang/bengali"
	"github.com/budhash/gomanize/scheme/colloquial"
	"github.com/budhash/gomanize/script/brahmic"
)

type injected struct {
	bengali.Bengali
	labels map[int]int
}

func (l *injected) Rules() core.RuleCatalog {
	catalog := l.Bengali.Rules()
	for i := range catalog.Schwa {
		r := &catalog.Schwa[i]
		if r.Name != "schwa.bengali.model" {
			continue
		}
		r.Condition = func(u *core.Unit, w *core.Word) bool {
			_, ok := l.labels[u.Start.Rune]
			return w.Options.SchwaModel && ok && u.Type == core.UnitConsonant && brahmic.GetBrahmicData(u) != nil && brahmic.GetSchwa(u) == brahmic.SchwaPending
		}
		r.Action = func(u *core.Unit, _ *core.Word) {
			label := l.labels[u.Start.Rune]
			if label == 0 {
				brahmic.SetSchwa(u, brahmic.SchwaDelete)
			} else {
				brahmic.SetSchwa(u, brahmic.SchwaKeep)
				if label == 2 {
					brahmic.GetBrahmicData(u).SchwaQuality = brahmic.SchwaRaised
				}
			}
		}
	}
	return catalog
}

type request struct {
	Native   string        `json:"native"`
	Variants []map[int]int `json:"variants"`
	Check    map[int]int   `json:"check"`
}

type response struct {
	B1       string   `json:"b1"`
	Model    string   `json:"model"`
	Variants []string `json:"variants"`
	Check    string   `json:"check"`
}

func validate(r request) error {
	if r.Native == "" || len(r.Variants) == 0 || len(r.Variants) > 9 {
		return fmt.Errorf("invalid word or candidate count")
	}
	for _, labels := range append(r.Variants, r.Check) {
		for index, label := range labels {
			if index < 0 || index >= len([]rune(r.Native)) || label < 0 || label > 2 {
				return fmt.Errorf("invalid vowel decision")
			}
		}
	}
	return nil
}

func run() error {
	l := &injected{}
	engine := core.NewEngine(l, colloquial.Colloquial{})
	b1, err := gomanize.New("bengali")
	if err != nil {
		return err
	}
	model, err := gomanize.NewWithOptions("bengali", gomanize.Options{SchwaModel: true})
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	encoder := json.NewEncoder(os.Stdout)
	for scanner.Scan() {
		var r request
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			return err
		}
		if err := validate(r); err != nil {
			return err
		}
		out := response{B1: b1.Translit(r.Native), Model: model.Translit(r.Native)}
		for _, labels := range r.Variants {
			l.labels = labels
			out.Variants = append(out.Variants, engine.TransliterateWithOptions(r.Native, core.Options{SchwaModel: true}))
		}
		l.labels = r.Check
		out.Check = engine.TransliterateWithOptions(r.Native, core.Options{SchwaModel: true})
		if err := encoder.Encode(out); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
