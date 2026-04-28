package quicli

import (
	"os"
	"testing"
	"time"
)

// setArgs temporarily overrides os.Args and returns a restore func.
func setArgs(args []string) func() {
	old := os.Args
	os.Args = args
	return func() { os.Args = old }
}

func TestGetFloatFlag(t *testing.T) {
	defer setArgs([]string{"prog", "--ratio", "3.14"})()
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "ratio", Default: float64(0), Description: "a ratio"}},
	}
	cfg := cli.Parse()
	if got := cfg.GetFloatFlag("ratio"); got != 3.14 {
		t.Errorf("GetFloatFlag: got %f, want 3.14", got)
	}
}

func TestFlagCustomShortName(t *testing.T) {
	defer setArgs([]string{"prog", "-x", "42"})()
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "count", Default: 0, Description: "test", ShortName: "x"}},
	}
	cfg := cli.Parse()
	if got := cfg.GetIntFlag("count"); got != 42 {
		t.Errorf("custom ShortName: got %d, want 42", got)
	}
}

func TestStringSliceFlag(t *testing.T) {
	defer setArgs([]string{"prog", "--file", "a", "--file", "b,c"})()
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "file", Default: []string{}, Description: "files"}},
	}
	cfg := cli.Parse()
	got := cfg.GetStringSliceFlag("file")
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("GetStringSliceFlag: got %v, want [a b c]", got)
	}
}

func TestDurationFlag(t *testing.T) {
	defer setArgs([]string{"prog", "--timeout", "5s"})()
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "timeout", Default: 30 * time.Second, Description: "request timeout"}},
	}
	cfg := cli.Parse()
	if got := cfg.GetDurationFlag("timeout"); got != 5*time.Second {
		t.Errorf("GetDurationFlag: got %v, want 5s", got)
	}
}

func TestDurationFlagDefault(t *testing.T) {
	defer setArgs([]string{"prog"})()
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "timeout", Default: 30 * time.Second, Description: "request timeout"}},
	}
	cfg := cli.Parse()
	if got := cfg.GetDurationFlag("timeout"); got != 30*time.Second {
		t.Errorf("GetDurationFlag default: got %v, want 30s", got)
	}
}

func TestParseDoesNotLeakGlobalState(t *testing.T) {
	defer setArgs([]string{"prog", "--count", "1"})()
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "count", Default: 0, Description: "count"}},
	}
	_ = cli.Parse()

	os.Args = []string{"prog", "--count", "2"}
	cfg2 := cli.Parse()
	if got := cfg2.GetIntFlag("count"); got != 2 {
		t.Errorf("second Parse: got %d, want 2", got)
	}
}
