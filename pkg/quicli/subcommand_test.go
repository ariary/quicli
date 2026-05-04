package quicli

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestSubcommandAlias(t *testing.T) {
	defer setArgs([]string{"prog", "f"})()
	var called bool
	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) {},
		Subcommands: Subcommands{
			{
				Name:        "feed",
				Aliases:     Aliases("f"),
				Description: "run feed",
				Function: func(cfg Config) {
					called = true
				},
			},
		},
	}
	cli.RunWithSubcommand()
	if !called {
		t.Error("subcommand alias 'f' did not invoke 'feed' function")
	}
}

func TestSubcommandPrefixMatch(t *testing.T) {
	defer setArgs([]string{"prog", "fe"})()
	var called bool
	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) {},
		Subcommands: Subcommands{
			{
				Name:        "feed",
				Description: "run feed",
				Function: func(cfg Config) {
					called = true
				},
			},
			{
				Name:        "get",
				Description: "get something",
				Function:    func(cfg Config) {},
			},
		},
	}
	cli.RunWithSubcommand()
	if !called {
		t.Error("prefix 'fe' did not resolve to 'feed'")
	}
}


func TestSubcommandExactMatchOverPrefix(t *testing.T) {
	// "get" is an exact match for the "get" subcommand, NOT a prefix of "getter"
	defer setArgs([]string{"prog", "get"})()
	var which string
	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) {},
		Subcommands: Subcommands{
			{
				Name:        "getter",
				Description: "getter sub",
				Function:    func(cfg Config) { which = "getter" },
			},
			{
				Name:        "get",
				Description: "get sub",
				Function:    func(cfg Config) { which = "get" },
			},
		},
	}
	cli.RunWithSubcommand()
	if which != "get" {
		t.Errorf("exact match: got %q, want 'get'", which)
	}
}

func TestSubcommandPrefixMatchOnAlias(t *testing.T) {
	// "li" is a prefix of alias "ls" — wait, "li" doesn't prefix "ls".
	// Use: "l" is a prefix of alias "ls" for "list"
	defer setArgs([]string{"prog", "l"})()
	var called bool
	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) {},
		Subcommands: Subcommands{
			{
				Name:        "get",
				Description: "get something",
				Function:    func(cfg Config) {},
			},
			{
				Name:        "list",
				Aliases:     Aliases("ls"),
				Description: "list things",
				Function:    func(cfg Config) { called = true },
			},
		},
	}
	cli.RunWithSubcommand()
	if !called {
		t.Error("prefix 'l' should resolve to 'list'")
	}
}

func TestSubcommandPrefixMatchWithSharedFlag(t *testing.T) {
	defer setArgs([]string{"prog", "fe", "--since", "3d"})()
	var receivedSince string
	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) {},
		Flags: Flags{
			{Name: "since", Default: "", Description: "time range", SharedSubcommand: SubcommandSet{"feed"}},
		},
		Subcommands: Subcommands{
			{
				Name:        "feed",
				Description: "run feed",
				Function: func(cfg Config) {
					receivedSince = cfg.GetStringFlag("since")
				},
			},
			{
				Name:        "get",
				Description: "get something",
				Function:    func(cfg Config) {},
			},
		},
	}
	cli.RunWithSubcommand()
	if receivedSince != "3d" {
		t.Errorf("prefix match with shared flag: got %q, want '3d'", receivedSince)
	}
}

func TestSubcommandPrefixMatchWithExclusiveFlag(t *testing.T) {
	defer setArgs([]string{"prog", "gre", "--name", "World"})()
	var receivedName string
	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) {},
		Subcommands: Subcommands{
			{
				Name:        "greet",
				Description: "greet someone",
				Flags:       Flags{{Name: "name", Default: "", Description: "who to greet"}},
				Function: func(cfg Config) {
					receivedName = cfg.GetStringFlag("name")
				},
			},
		},
	}
	cli.RunWithSubcommand()
	if receivedName != "World" {
		t.Errorf("prefix match with exclusive flag: got %q, want 'World'", receivedName)
	}
}

func TestSubcommandExclusiveFlag(t *testing.T) {
	defer setArgs([]string{"prog", "greet", "--name", "World"})()
	var receivedName string
	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) {},
		Subcommands: Subcommands{
			{
				Name:        "greet",
				Description: "greet someone",
				Flags:       Flags{{Name: "name", Default: "", Description: "who to greet"}},
				Function: func(cfg Config) {
					receivedName = cfg.GetStringFlag("name")
				},
			},
		},
	}
	cli.RunWithSubcommand()
	if receivedName != "World" {
		t.Errorf("subcommand Flags: got %q, want World", receivedName)
	}
}

func TestSubcommandAmbiguousPrefix(t *testing.T) {
	// "g" prefixes both "get" and "greet" — ambiguous, no match.
	sub := getSubcommandByName(
		Subcommands{
			{Name: "get", Description: "get", Function: func(Config) {}},
			{Name: "greet", Description: "greet", Function: func(Config) {}},
		},
		"g",
	)
	if sub.Name != "" {
		t.Errorf("ambiguous prefix: expected empty, got %q", sub.Name)
	}
}

func TestSubcommandRootWithNoArgs(t *testing.T) {
	defer setArgs([]string{"prog"})()
	var rootCalled bool
	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) { rootCalled = true },
		Subcommands: Subcommands{
			{Name: "sub", Description: "sub", Function: func(cfg Config) {}},
		},
	}
	cli.RunWithSubcommand()
	if !rootCalled {
		t.Error("root function should be called with no args")
	}
}

func TestSubcommandRootWithNoArgsAndSubcommandFlag(t *testing.T) {
	defer setArgs([]string{"prog"})()
	var rootCalled bool
	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) { rootCalled = true },
		Flags: Flags{
			{Name: "since", Default: "", Description: "time range", NotForRootCommand: true, SharedSubcommand: SubcommandSet{"sub"}},
		},
		Subcommands: Subcommands{
			{Name: "sub", Description: "sub", Function: func(cfg Config) {}},
		},
	}
	cli.RunWithSubcommand()
	if !rootCalled {
		t.Error("root function should be called with no args and subcommand-only flags")
	}
}

func TestRunWithSubcommandRootCommandWithFlags(t *testing.T) {
	defer setArgs([]string{"prog", "--verbose"})()
	var gotVerbose bool
	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) { gotVerbose = cfg.GetBoolFlag("verbose") },
		Flags:       Flags{{Name: "verbose", Default: false, Description: "verbose"}},
		Subcommands: Subcommands{
			{Name: "sub", Description: "sub", Function: func(cfg Config) {}},
		},
	}
	cli.RunWithSubcommand()
	if !gotVerbose {
		t.Error("root command with --verbose flag should parse and set verbose to true")
	}
}

// --- Subprocess tests for RunWithSubcommand os.Exit paths ---

func TestUsageShowsAvailableCommands(t *testing.T) {
	if os.Getenv("TEST_AVAIL_CMDS") == "1" {
		defer setArgs([]string{"prog", "--help"})()
		cli := Cli{
			Usage:       "prog [command]",
			Description: "test",
			Function:    func(cfg Config) {},
			Subcommands: Subcommands{
				{Name: "build", Description: "build it", Function: func(cfg Config) {}},
			},
		}
		cli.RunWithSubcommand()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestUsageShowsAvailableCommands$")
	cmd.Env = append(os.Environ(), "TEST_AVAIL_CMDS=1")
	out, _ := cmd.CombinedOutput()
	if !strings.Contains(string(out), "Available commands:") {
		t.Errorf("usage should show 'Available commands:', got: %s", out)
	}
}

func TestUsageNoAvailableCommandsWithoutSubs(t *testing.T) {
	if os.Getenv("TEST_NO_AVAIL_CMDS") == "1" {
		defer setArgs([]string{"prog", "--help"})()
		cli := Cli{
			Usage:       "prog [flags]",
			Description: "test",
			Function:    func(cfg Config) {},
		}
		cli.RunWithSubcommand()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestUsageNoAvailableCommandsWithoutSubs$")
	cmd.Env = append(os.Environ(), "TEST_NO_AVAIL_CMDS=1")
	out, _ := cmd.CombinedOutput()
	if strings.Contains(string(out), "Available commands:") {
		t.Errorf("usage should NOT show 'Available commands:' without subcommands, got: %s", out)
	}
}

func TestRunWithSubcommandCheatSheetCSRecognized(t *testing.T) {
	if os.Getenv("TEST_SUB_CS") == "1" {
		defer setArgs([]string{"prog", "--cs"})()
		cli := Cli{
			Usage:       "prog [command]",
			Description: "test",
			Function:    func(cfg Config) {},
			CheatSheet:  Examples{{Title: "ex", CommandLine: "prog build"}},
			Subcommands: Subcommands{
				{Name: "build", Description: "build it", Function: func(cfg Config) {}},
			},
		}
		cli.RunWithSubcommand()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestRunWithSubcommandCheatSheetCSRecognized$")
	cmd.Env = append(os.Environ(), "TEST_SUB_CS=1")
	err := cmd.Run()
	if err != nil {
		t.Errorf("--cs should be recognized in RunWithSubcommand with CheatSheet: %v", err)
	}
}

func TestRunWithSubcommandCheatSheetCSNotRegisteredWhenEmpty(t *testing.T) {
	if os.Getenv("TEST_SUB_CS_EMPTY") == "1" {
		defer setArgs([]string{"prog", "--cs"})()
		cli := Cli{
			Usage:       "prog [command]",
			Description: "test",
			Function:    func(cfg Config) {},
			Subcommands: Subcommands{
				{Name: "build", Description: "build it", Function: func(cfg Config) {}},
			},
		}
		cli.RunWithSubcommand()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestRunWithSubcommandCheatSheetCSNotRegisteredWhenEmpty$")
	cmd.Env = append(os.Environ(), "TEST_SUB_CS_EMPTY=1")
	err := cmd.Run()
	if err == nil {
		t.Error("--cs should not be recognized without CheatSheet entries in RunWithSubcommand")
	}
}

func TestRunWithSubcommandCheatSheetPrinted(t *testing.T) {
	if os.Getenv("TEST_SUB_CS_PRINT") == "1" {
		defer setArgs([]string{"prog", "--cs"})()
		cli := Cli{
			Usage:       "prog [command]",
			Description: "test",
			Function:    func(cfg Config) {},
			CheatSheet:  Examples{{Title: "sub-example", CommandLine: "prog build"}},
			Subcommands: Subcommands{
				{Name: "build", Description: "build it", Function: func(cfg Config) {}},
			},
		}
		cli.RunWithSubcommand()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestRunWithSubcommandCheatSheetPrinted$")
	cmd.Env = append(os.Environ(), "TEST_SUB_CS_PRINT=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error: %v, output: %s", err, out)
	}
	if !strings.Contains(string(out), "sub-example") {
		t.Errorf("cheat sheet should be printed, got: %s", out)
	}
}
