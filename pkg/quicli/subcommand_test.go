package quicli

import "testing"

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
