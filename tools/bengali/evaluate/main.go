// Command evaluate streams B1/model predictions for reproducible corpus scoring.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	gomanize "github.com/budhash/gomanize"
)

func main() {
	baseline, err := gomanize.New("bengali")
	if err != nil {
		panic(err)
	}
	model, err := gomanize.NewWithOptions("bengali", gomanize.Options{SchwaModel: true})
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	encoder := json.NewEncoder(os.Stdout)
	for scanner.Scan() {
		var native string
		if err = json.Unmarshal(scanner.Bytes(), &native); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err = encoder.Encode([]string{baseline.Translit(native), model.Translit(native)}); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if err = scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
