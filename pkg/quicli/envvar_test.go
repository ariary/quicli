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
		{"appZ", "flag", "APPZ_FLAG"},   // Z boundary
		{"app0", "flag", "APP0_FLAG"},   // 0 boundary
		{"app9", "flag", "APP9_FLAG"},   // 9 boundary
		{"app09Z", "x", "APP09Z_X"},     // combined boundaries
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

func TestEnvOnlyFlagResolved(t *testing.T) {
	defer setArgs([]string{"prog"})()
	t.Setenv("PROG_SECRET", "s3cret")

	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "secret", Default: "", Description: "a secret", EnvOnly: true}},
	}
	cfg := cli.Parse()
	if got := cfg.GetStringFlag("secret"); got != "s3cret" {
		t.Errorf("env-only string flag: got %q, want %q", got, "s3cret")
	}
}

func TestEnvOnlyFlagWithExplicitEnvVar(t *testing.T) {
	defer setArgs([]string{"prog"})()
	t.Setenv("MY_SECRET", "token123")

	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "secret", Default: "", Description: "a secret", EnvOnly: true, EnvVar: "MY_SECRET"}},
	}
	cfg := cli.Parse()
	if got := cfg.GetStringFlag("secret"); got != "token123" {
		t.Errorf("env-only explicit env var: got %q, want %q", got, "token123")
	}
}

func TestEnvOnlyFlagDefault(t *testing.T) {
	defer setArgs([]string{"prog"})()
	// No env var set — should use default value.

	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "secret", Default: "fallback", Description: "a secret", EnvOnly: true}},
	}
	cfg := cli.Parse()
	if got := cfg.GetStringFlag("secret"); got != "fallback" {
		t.Errorf("env-only default: got %q, want %q", got, "fallback")
	}
}

func TestEnvOnlyFlagIntType(t *testing.T) {
	defer setArgs([]string{"prog"})()
	t.Setenv("PROG_PORT", "9090")

	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "port", Default: 8080, Description: "port", EnvOnly: true}},
	}
	cfg := cli.Parse()
	if got := cfg.GetIntFlag("port"); got != 9090 {
		t.Errorf("env-only int flag: got %d, want 9090", got)
	}
}

func TestEnvOnlyFlagBoolType(t *testing.T) {
	defer setArgs([]string{"prog"})()
	t.Setenv("PROG_DEBUG", "true")

	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "debug", Default: false, Description: "debug mode", EnvOnly: true}},
	}
	cfg := cli.Parse()
	if got := cfg.GetBoolFlag("debug"); got != true {
		t.Errorf("env-only bool flag: got %v, want true", got)
	}
}
