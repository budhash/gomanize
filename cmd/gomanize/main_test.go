package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	gomanize "github.com/budhash/gomanize"
)

// Exercise the actual CLI, including os.Exit, flags, stdin and file modes.
func TestCLIProcess(t *testing.T) {
	if os.Getenv("GOMANIZE_CLI_TEST_CHILD") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{os.Args[0]}, os.Args[i+1:]...)
			main()
			os.Exit(0)
		}
	}
	t.Fatal("missing argument separator")
}

func cli(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestCLIProcess$", "--"}, args...)...)
	cmd.Env = append(os.Environ(), "GOMANIZE_CLI_TEST_CHILD=1")
	cmd.Stdin = strings.NewReader(stdin)
	var out, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &stderr
	err := cmd.Run()
	return out.String(), stderr.String(), err
}

func TestCLILanguage(t *testing.T) {
	for _, language := range []string{"hindi", "bengali"} {
		text := "নমস্কার, সমতা।\nनमस्ते दुनिया English 123"
		for _, learned := range []bool{false, true} {
			opts := gomanize.Options{SchwaModel: learned, Rerank: learned, Lexicon: learned}
			g, err := gomanize.NewWithOptions(language, opts)
			if err != nil {
				t.Fatal(err)
			}
			args := []string{"--language=" + language}
			if learned {
				args = append(args, "--schwa-model", "--rerank", "--lexicon")
			}
			for _, stdin := range []bool{false, true} {
				a := append([]string{}, args...)
				input := ""
				if stdin {
					input = text
				} else {
					a = append(a, text)
				}
				got, stderr, err := cli(t, input, a...)
				if err != nil || stderr != "" || got != g.Translit(text)+"\n" {
					t.Fatalf("%s learned=%v stdin=%v: %q %q %v", language, learned, stdin, got, stderr, err)
				}
			}
		}
	}
	for _, args := range [][]string{{"नमस्ते दुनिया"}, {"--language", "HINDI", "नमस्ते दुनिया"}, {"--language", "bengali", "নমস্কার"}} {
		got, stderr, err := cli(t, "", args...)
		want := "namaste duniya\n"
		if args[len(args)-1] == "নমস্কার" {
			want = "nomoskar\n"
		}
		if err != nil || stderr != "" || got != want {
			t.Fatalf("%v: %q %q %v", args, got, stderr, err)
		}
	}
	for _, args := range [][]string{{"--language"}, {"--language", "--debug"}, {"--language="}, {"--language=unknown", "নমস্কার"}} {
		out, stderr, err := cli(t, "", args...)
		if err == nil || out != "" || !strings.Contains(stderr, "Error:") {
			t.Fatalf("invalid language %v: %q %q %v", args, out, stderr, err)
		}
	}
}

func TestCLIBengaliModes(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.txt")
	cases := filepath.Join(dir, "cases.tsv")
	for path, text := range map[string]string{input: "# comment\n\nনমস্কার\n", cases: "নমস্কার\tnomoskar\n"} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"--input=" + input}, "nomoskar\n"},
		{[]string{"--test=" + cases}, "Passed: 1 / 1 (100.0%)\n"},
		{[]string{"--test=" + cases, "--diff"}, ""},
		{[]string{"--debug", "নমস্কার"}, "Output: nomoskar"},
		{[]string{"--list-rules"}, "bengali."},
	} {
		out, stderr, err := cli(t, "", append([]string{"--language=bengali"}, tc.args...)...)
		if err != nil || stderr != "" || !strings.Contains(out, tc.want) {
			t.Fatalf("%v: %q %q %v", tc.args, out, stderr, err)
		}
	}
	g, _ := gomanize.NewWithOptions("bengali", gomanize.Options{}, gomanize.WithDisabledRules("*"))
	out, stderr, err := cli(t, "", "--language=bengali", "--disable-rule=*", "নমস্কার")
	if err != nil || stderr != "" || out != g.Translit("নমস্কার")+"\n" {
		t.Fatalf("rule override: %q %q %v", out, stderr, err)
	}
}
