package quicli

import (
	"os"
	"testing"
)

func TestEnvVarName(t *testing.T) {
	cases := []struct {
		progName string
		flagName string
		want     string
	}{
		{"say-hello", "count", "SAY_HELLO_COUNT"},
		{"./mycli", "output-format", "MYCLI_OUTPUT_FORMAT"},
		{"prog", "file", "PROG_FILE"},
	}
	for _, tc := range cases {
		if got := envVarName(tc.progName, tc.flagName); got != tc.want {
			t.Errorf("envVarName(%q, %q) = %q, want %q", tc.progName, tc.flagName, got, tc.want)
		}
	}
}

func TestApplyEnvVarInt(t *testing.T) {
	defer setArgs([]string{"prog"})()
	os.Setenv("PROG_COUNT", "7")
	defer os.Unsetenv("PROG_COUNT")

	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "count", Default: 0, Description: "count"}},
	}
	cfg := cli.Parse()
	if got := cfg.GetIntFlag("count"); got != 7 {
		t.Errorf("env var int: got %d, want 7", got)
	}
}

func TestApplyEnvVarExplicit(t *testing.T) {
	defer setArgs([]string{"prog"})()
	os.Setenv("MY_CUSTOM_COUNT", "99")
	defer os.Unsetenv("MY_CUSTOM_COUNT")

	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "count", Default: 0, Description: "count", EnvVar: "MY_CUSTOM_COUNT"}},
	}
	cfg := cli.Parse()
	if got := cfg.GetIntFlag("count"); got != 99 {
		t.Errorf("explicit EnvVar: got %d, want 99", got)
	}
}

func TestApplyEnvVarOptOut(t *testing.T) {
	defer setArgs([]string{"prog"})()
	os.Setenv("PROG_SECRET", "ignored")
	defer os.Unsetenv("PROG_SECRET")

	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "secret", Default: "default", Description: "secret", EnvVar: "-"}},
	}
	cfg := cli.Parse()
	if got := cfg.GetStringFlag("secret"); got != "default" {
		t.Errorf("opt-out: got %q, want default", got)
	}
}

func TestApplyEnvVarCLIOverridesEnv(t *testing.T) {
	defer setArgs([]string{"prog", "--count", "3"})()
	os.Setenv("PROG_COUNT", "99")
	defer os.Unsetenv("PROG_COUNT")

	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "count", Default: 0, Description: "count"}},
	}
	cfg := cli.Parse()
	if got := cfg.GetIntFlag("count"); got != 3 {
		t.Errorf("CLI should override env var: got %d, want 3", got)
	}
}
