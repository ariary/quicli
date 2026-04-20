package quicli

import (
	"os"
	"testing"
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
