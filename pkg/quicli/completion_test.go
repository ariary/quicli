package quicli

import (
	"strings"
	"testing"
)

func testCli() Cli {
	return Cli{
		Usage:       "prog [command] [flags]",
		Description: "test cli",
		Flags: Flags{
			{Name: "verbose", Default: false, Description: "verbose output"},
			{Name: "output", Default: "text", Description: "output format"},
		},
		Subcommands: Subcommands{
			{Name: "build", Aliases: Aliases("b"), Description: "build the project", Function: func(Config) {}},
			{Name: "test", Description: "run tests", Function: func(Config) {}},
		},
		Function: func(Config) {},
	}
}

func TestGenerateBashCompletion(t *testing.T) {
	cli := testCli()
	script, err := generateCompletion(&cli, "bash")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, "build") {
		t.Error("bash completion missing subcommand 'build'")
	}
	if !strings.Contains(script, "--verbose") {
		t.Error("bash completion missing flag '--verbose'")
	}
	if !strings.Contains(script, "complete -F") {
		t.Error("bash completion missing complete builtin")
	}
}

func TestGenerateZshCompletion(t *testing.T) {
	cli := testCli()
	script, err := generateCompletion(&cli, "zsh")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, "#compdef") {
		t.Error("zsh completion missing #compdef header")
	}
	if !strings.Contains(script, "build") {
		t.Error("zsh completion missing subcommand 'build'")
	}
}

func TestGenerateFishCompletion(t *testing.T) {
	cli := testCli()
	script, err := generateCompletion(&cli, "fish")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, "complete -c") {
		t.Error("fish completion missing 'complete -c'")
	}
	if !strings.Contains(script, "build") {
		t.Error("fish completion missing subcommand 'build'")
	}
}

func TestGenerateCompletionUnknownShell(t *testing.T) {
	cli := testCli()
	_, err := generateCompletion(&cli, "powershell")
	if err == nil {
		t.Error("expected error for unknown shell")
	}
}
