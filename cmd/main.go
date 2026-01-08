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
	var disableRules []string
	var enableRules []string
	var listRulesPattern string
	debug := false
	listRules := false

	// Parse flags and collect text arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-v" || arg == "--version" || arg == "version":
			fmt.Printf("gomanize %s (commit: %s, built: %s)\n", version, commit, date)
			os.Exit(0)
		case arg == "-h" || arg == "--help" || arg == "help":
			printUsage()
			os.Exit(0)
		case arg == "--long-vowels":
			opts.LongVowels = true
		case arg == "--debug":
			debug = true
		case arg == "--list-rules":
			listRules = true
			// Check if next arg is a pattern (not starting with -)
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				listRulesPattern = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--disable-rule="):
			disableRules = append(disableRules, strings.TrimPrefix(arg, "--disable-rule="))
		case strings.HasPrefix(arg, "--enable-rule="):
			enableRules = append(enableRules, strings.TrimPrefix(arg, "--enable-rule="))
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", arg)
			printUsage()
			os.Exit(1)
		default:
			textArgs = append(textArgs, arg)
		}
	}

	// Create engine with rule overrides
	var engineOpts []gomanize.EngineOption
	if len(disableRules) > 0 {
		engineOpts = append(engineOpts, gomanize.WithDisabledRules(disableRules...))
	}
	if len(enableRules) > 0 {
		engineOpts = append(engineOpts, gomanize.WithEnabledRules(enableRules...))
	}

	g, err := gomanize.NewWithOptions("hindi", opts, engineOpts...)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	// Handle --list-rules
	if listRules {
		rules := g.ListRules(listRulesPattern)
		if rules == nil {
			fmt.Fprintln(os.Stderr, "Rule listing not supported for this language")
			os.Exit(1)
		}
		printRules(rules)
		os.Exit(0)
	}

	input := getInput(textArgs)
	if input == "" {
		printUsage()
		os.Exit(1)
	}

	if debug {
		// Debug mode: process each word individually and show debug info
		words := strings.Split(input, " ")
		for i, word := range words {
			if i > 0 {
				fmt.Println()
			}
			result, info := g.TranslitDebug(word)
			printDebugInfo(word, result, info)
		}
	} else {
		output := g.Translit(input)
		fmt.Println(output)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: gomanize [options] <text>")
	fmt.Fprintln(os.Stderr, "       echo <text> | gomanize [options]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  --long-vowels            Use 'aa' for all ा positions (e.g., गाना→gaana)")
	fmt.Fprintln(os.Stderr, "  --debug                  Show debug info (parsed units, rule applications)")
	fmt.Fprintln(os.Stderr, "  --list-rules [PATTERN]   List all rules (optionally filtered by pattern)")
	fmt.Fprintln(os.Stderr, "  --disable-rule=PATTERN   Disable rules matching pattern (e.g., schwa.*)")
	fmt.Fprintln(os.Stderr, "  --enable-rule=PATTERN    Enable rules matching pattern")
	fmt.Fprintln(os.Stderr, "  --version                Show version information")
	fmt.Fprintln(os.Stderr, "  --help                   Show this help message")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Rule Patterns:")
	fmt.Fprintln(os.Stderr, "  Exact match: schwa.delete.ccv")
	fmt.Fprintln(os.Stderr, "  Glob:        schwa.* (all schwa rules)")
	fmt.Fprintln(os.Stderr, "  Glob:        schwa.delete.* (all schwa deletion rules)")
}

func printRules(rules []gomanize.RuleStatus) {
	fmt.Println("Rules:")
	for _, r := range rules {
		status := "enabled"
		if !r.Enabled {
			status = "disabled"
		}
		if r.Conditional != "" {
			status = fmt.Sprintf("needs --%s", toFlagName(r.Conditional))
		}
		fmt.Printf("  %-40s [%s] %s:%d\n", r.Name, status, r.Phase, r.Priority)
	}
}

// toFlagName converts a Go field name to a CLI flag name (e.g., "LongVowels" -> "long-vowels")
func toFlagName(fieldName string) string {
	var result []rune
	for i, r := range fieldName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '-')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

func printDebugInfo(input, result string, info *gomanize.DebugInfo) {
	fmt.Printf("Input:  %s\n", input)
	fmt.Printf("Output: %s\n", result)

	if info == nil {
		fmt.Println("(Debug info not available for this engine)")
		return
	}

	fmt.Println("\nParsed Units:")
	for _, u := range info.Units {
		meta := ""
		if u.Metadata != "" {
			meta = fmt.Sprintf(" [%s]", u.Metadata)
		}
		fmt.Printf("  [%d] %s → %s (%s @ rune %d)%s\n",
			u.Index, u.Chars, u.BaseRom, u.Type, u.RunePos, meta)
	}

	if len(info.Traces) > 0 {
		fmt.Println("\nRule Applications:")
		for _, t := range info.Traces {
			change := ""
			if t.Before != t.After {
				change = fmt.Sprintf(" %s→%s", t.Before, t.After)
			}
			meta := ""
			if t.Metadata != "" {
				meta = fmt.Sprintf(" [%s]", t.Metadata)
			}
			fmt.Printf("  %s: %s on %s (idx %d)%s%s\n",
				t.Phase, t.Rule, t.Unit, t.UnitIdx, change, meta)
		}
	}
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
