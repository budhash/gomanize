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
	args := os.Args[1:]
	opts := gomanize.NewOptions()
	var textArgs []string

	// Parse flags and collect text arguments
	for _, arg := range args {
		switch arg {
		case "-v", "--version", "version":
			fmt.Printf("gomanize %s (commit: %s, built: %s)\n", version, commit, date)
			os.Exit(0)
		case "-h", "--help", "help":
			printUsage()
			os.Exit(0)
		case "--long-vowels":
			opts.LongVowels = true
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", arg)
				printUsage()
				os.Exit(1)
			}
			textArgs = append(textArgs, arg)
		}
	}

	input := getInput(textArgs)
	if input == "" {
		printUsage()
		os.Exit(1)
	}

	g, err := gomanize.NewWithOptions("hindi", opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	output := g.Translit(input)
	fmt.Println(output)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: gomanize [options] <text>")
	fmt.Fprintln(os.Stderr, "       echo <text> | gomanize [options]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  --long-vowels  Use 'aa' for all ा positions (e.g., गाना→gaana)")
	fmt.Fprintln(os.Stderr, "  --version      Show version information")
	fmt.Fprintln(os.Stderr, "  --help         Show this help message")
}

func getInput(textArgs []string) string {
	if len(textArgs) == 0 {
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
	return strings.Join(textArgs, " ")
}
