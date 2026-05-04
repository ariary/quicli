package quicli

import (
	"os"
	"os/exec"
	"strings"
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

func TestBoolFlagShortName(t *testing.T) {
	defer setArgs([]string{"prog", "-v"})()
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "verbose", Default: false, Description: "verbose"}},
	}
	cfg := cli.Parse()
	if !cfg.GetBoolFlag("verbose") {
		t.Error("bool flag via short -v should be true")
	}
}

func TestFloatFlagShortName(t *testing.T) {
	defer setArgs([]string{"prog", "-r", "2.5"})()
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "ratio", Default: float64(0), Description: "ratio"}},
	}
	cfg := cli.Parse()
	if got := cfg.GetFloatFlag("ratio"); got != 2.5 {
		t.Errorf("float via short -r: got %f, want 2.5", got)
	}
}

func TestStringSliceFlagShortName(t *testing.T) {
	defer setArgs([]string{"prog", "-f", "a", "-f", "b"})()
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "file", Default: []string{}, Description: "files"}},
	}
	cfg := cli.Parse()
	got := cfg.GetStringSliceFlag("file")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("slice via short -f: got %v, want [a b]", got)
	}
}

func TestDurationFlagShortName(t *testing.T) {
	defer setArgs([]string{"prog", "-t", "5s"})()
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "timeout", Default: 30 * time.Second, Description: "timeout"}},
	}
	cfg := cli.Parse()
	if got := cfg.GetDurationFlag("timeout"); got != 5*time.Second {
		t.Errorf("duration via short -t: got %v, want 5s", got)
	}
}

func TestValueFlagShortName(t *testing.T) {
	defer setArgs([]string{"prog", "-l", "warn"})()
	lv := &testLevel{val: "info"}
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "level", Default: lv, Description: "log level"}},
	}
	cfg := cli.Parse()
	got := cfg.Flags["level"].(*testLevel)
	if got.val != "warn" {
		t.Errorf("value via short -l: got %q, want warn", got.val)
	}
}

// --- Subprocess tests for CheatSheet / os.Exit paths ---

func TestCheatSheetCSFlagRecognized(t *testing.T) {
	if os.Getenv("TEST_CS_RECOGNIZED") == "1" {
		defer setArgs([]string{"prog", "--cs"})()
		cli := Cli{
			Usage:       "prog [flags]",
			Description: "test",
			CheatSheet:  Examples{{Title: "ex", CommandLine: "prog --foo"}},
		}
		cli.Parse()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestCheatSheetCSFlagRecognized$")
	cmd.Env = append(os.Environ(), "TEST_CS_RECOGNIZED=1")
	err := cmd.Run()
	if err != nil {
		t.Errorf("--cs should be recognized with CheatSheet: %v", err)
	}
}

func TestCheatSheetCSFlagNotRegisteredWhenEmpty(t *testing.T) {
	if os.Getenv("TEST_CS_EMPTY") == "1" {
		defer setArgs([]string{"prog", "--cs"})()
		cli := Cli{
			Usage:       "prog [flags]",
			Description: "test",
		}
		cli.Parse()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestCheatSheetCSFlagNotRegisteredWhenEmpty$")
	cmd.Env = append(os.Environ(), "TEST_CS_EMPTY=1")
	err := cmd.Run()
	if err == nil {
		t.Error("--cs should not be recognized without CheatSheet entries")
	}
}

func TestCheatSheetPrinted(t *testing.T) {
	if os.Getenv("TEST_CS_PRINT") == "1" {
		defer setArgs([]string{"prog", "--cs"})()
		cli := Cli{
			Usage:       "prog [flags]",
			Description: "test",
			CheatSheet:  Examples{{Title: "my-example", CommandLine: "prog --foo"}},
		}
		cli.Parse()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestCheatSheetPrinted$")
	cmd.Env = append(os.Environ(), "TEST_CS_PRINT=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error: %v, output: %s", err, out)
	}
	if !strings.Contains(string(out), "my-example") {
		t.Errorf("cheat sheet should be printed, got: %s", out)
	}
}
