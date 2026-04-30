package quicli

import "testing"

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
