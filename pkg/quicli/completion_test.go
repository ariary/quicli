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

func hasStr(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func TestAllFlagNamesWithAndWithoutShort(t *testing.T) {
	flags := []Flag{
		{Name: "verbose", Description: "verbose"},
		{Name: "output", Description: "output", NoShortName: true},
	}
	names := allFlagNames(flags)
	if !hasStr(names, "--verbose") {
		t.Error("--verbose should be present")
	}
	if !hasStr(names, "-v") {
		t.Error("-v should be present for flag without NoShortName")
	}
	if !hasStr(names, "--output") {
		t.Error("--output should be present")
	}
	if hasStr(names, "-o") {
		t.Error("-o should NOT be present (NoShortName=true)")
	}
}

func TestBashCompletionNoSubcommands(t *testing.T) {
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "verbose", Default: false, Description: "verbose"}},
		Function:    func(Config) {},
	}
	script, err := generateCompletion(&cli, "bash")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(script, "local commands") {
		t.Error("bash: no subcommands should not have 'local commands'")
	}
	if !strings.Contains(script, "--verbose") {
		t.Error("bash: flags should appear")
	}
	if strings.Contains(script, "COMP_CWORD -eq 1") {
		t.Error("bash: no subcommand branch expected without subcommands")
	}
}

func TestZshCompletionNoSubcommands(t *testing.T) {
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "verbose", Default: false, Description: "verbose"}},
		Function:    func(Config) {},
	}
	script, err := generateCompletion(&cli, "zsh")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(script, "commands=(") {
		t.Error("zsh: no subcommands should not have commands array")
	}
	if strings.Contains(script, "_describe") {
		t.Error("zsh: no subcommands should not use _describe")
	}
	if !strings.Contains(script, "_arguments") {
		t.Error("zsh: should use _arguments for flags")
	}
}

func TestFishCompletionAliases(t *testing.T) {
	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Subcommands: Subcommands{
			{Name: "build", Aliases: Aliases("b"), Description: "build stuff", Function: func(Config) {}},
		},
		Flags:    Flags{{Name: "verbose", Default: false, Description: "verbose"}},
		Function: func(Config) {},
	}
	script, err := generateCompletion(&cli, "fish")
	if err != nil {
		t.Fatal(err)
	}
	// Check for the alias-specific line (contains "(alias)" marker)
	if !strings.Contains(script, "(alias)") {
		t.Error("fish: alias should produce '(alias)' marker in completion")
	}
}

func TestFishCompletionNoShortName(t *testing.T) {
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "verbose", Default: false, Description: "verbose", NoShortName: true}},
		Function:    func(Config) {},
	}
	script, err := generateCompletion(&cli, "fish")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, "-l verbose") {
		t.Error("fish: long name should appear")
	}
	if strings.Contains(script, "-s v") {
		t.Error("fish: NoShortName should not produce -s flag")
	}
}

func TestFishCompletionWithShortName(t *testing.T) {
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "verbose", Default: false, Description: "verbose"}},
		Function:    func(Config) {},
	}
	script, err := generateCompletion(&cli, "fish")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, "-s v") {
		t.Error("fish: normal flag should produce -s short name")
	}
}

func TestAllFlagNamesSkipsEnvOnly(t *testing.T) {
	flags := []Flag{
		{Name: "output", Default: "", Description: "output file"},
		{Name: "secret", Default: "", Description: "API secret", EnvOnly: true},
	}
	names := allFlagNames(flags)
	for _, n := range names {
		if strings.Contains(n, "secret") {
			t.Errorf("env-only flag should not appear in completion: %v", names)
		}
	}
	if len(names) != 2 {
		t.Errorf("expected 2 entries (--output, -o), got %d: %v", len(names), names)
	}
}
