package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	gomanize "github.com/budhash/gomanize"
)

// Set by GoReleaser via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Handle version flag
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Printf("gomanize %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	input := getInput()
	if input == "" {
		fmt.Fprintln(os.Stderr, "Usage: gomanize <text>")
		fmt.Fprintln(os.Stderr, "       echo <text> | gomanize")
		fmt.Fprintln(os.Stderr, "       gomanize --version")
		os.Exit(1)
	}

	g, err := gomanize.New("hindi")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	output := g.Translit(input)
	fmt.Println(output)
}

func getInput() string {
	args := os.Args[1:]
	if len(args) == 0 {
		// Check if there's data on stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Data is being piped
			stdin, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error reading stdin:", err)
				os.Exit(1)
			}
			return strings.TrimSpace(string(stdin))
		}
		return ""
	}
	return strings.Join(args, " ")
}
