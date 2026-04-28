package quicli

import (
	"flag"
	"strings"
	"testing"
	"time"
)

// --- checkFlags unit tests (no os.Exit, no subprocess) ---

func newFlagSet(flags []Flag, args []string) *flag.FlagSet {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	for _, f := range flags {
		switch d := f.Default.(type) {
		case string:
			fs.String(f.Name, d, f.Description)
		case int:
			fs.Int(f.Name, d, f.Description)
		case bool:
			fs.Bool(f.Name, d, f.Description)
		case float64:
			fs.Float64(f.Name, d, f.Description)
		}
	}
	fs.Parse(args)
	return fs
}

func TestCheckFlagsRequiredMissing(t *testing.T) {
	flags := []Flag{{Name: "output", Default: "", Description: "out", Required: true}}
	fs := newFlagSet(flags, []string{})
	errs := checkFlags(flags, fs)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0], "--output") {
		t.Errorf("error should mention --output: %s", errs[0])
	}
}

func TestCheckFlagsRequiredProvided(t *testing.T) {
	flags := []Flag{{Name: "output", Default: "", Description: "out", Required: true}}
	fs := newFlagSet(flags, []string{"--output", "file.txt"})
	errs := checkFlags(flags, fs)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestCheckFlagsChoicesValid(t *testing.T) {
	flags := []Flag{{Name: "format", Default: "json", Description: "fmt", Choices: []string{"json", "yaml"}}}
	fs := newFlagSet(flags, []string{"--format", "yaml"})
	errs := checkFlags(flags, fs)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestCheckFlagsChoicesInvalid(t *testing.T) {
	flags := []Flag{{Name: "format", Default: "json", Description: "fmt", Choices: []string{"json", "yaml"}}}
	fs := newFlagSet(flags, []string{"--format", "xml"})
	errs := checkFlags(flags, fs)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0], "xml") || !strings.Contains(errs[0], "json, yaml") {
		t.Errorf("error should mention value and choices: %s", errs[0])
	}
}

func TestCheckFlagsChoicesDefaultNotSet(t *testing.T) {
	// When flag is not set and default is empty, no validation error.
	flags := []Flag{{Name: "format", Default: "", Description: "fmt", Choices: []string{"json", "yaml"}}}
	fs := newFlagSet(flags, []string{})
	errs := checkFlags(flags, fs)
	if len(errs) != 0 {
		t.Errorf("expected no errors for unset optional choice, got %v", errs)
	}
}

func TestCheckFlagsRequiredAndChoices(t *testing.T) {
	flags := []Flag{{Name: "env", Default: "", Description: "env", Required: true, Choices: []string{"dev", "prod"}}}
	// Missing required → 1 error. Also "" is not in choices but flag not set → only required error.
	fs := newFlagSet(flags, []string{})
	errs := checkFlags(flags, fs)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error (required), got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "required") {
		t.Errorf("error should be about required: %s", errs[0])
	}
}

func TestCheckFlagsRequiredAndChoicesInvalidValue(t *testing.T) {
	flags := []Flag{{Name: "env", Default: "", Description: "env", Required: true, Choices: []string{"dev", "prod"}}}
	fs := newFlagSet(flags, []string{"--env", "staging"})
	errs := checkFlags(flags, fs)
	// Required is satisfied, but "staging" is not a valid choice.
	if len(errs) != 1 {
		t.Fatalf("expected 1 error (invalid choice), got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "staging") {
		t.Errorf("error should mention staging: %s", errs[0])
	}
}

func TestCheckFlagsMultipleErrors(t *testing.T) {
	flags := []Flag{
		{Name: "output", Default: "", Description: "out", Required: true},
		{Name: "target", Default: "", Description: "tgt", Required: true},
	}
	fs := newFlagSet(flags, []string{})
	errs := checkFlags(flags, fs)
	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d: %v", len(errs), errs)
	}
}

// --- RunFunc / Cli.Parse integration tests ---

func TestRequiredFlagProvided(t *testing.T) {
	defer setArgs([]string{"prog", "--output", "out.txt"})()
	type Opts struct {
		Output string `cli:"output file" required:"true"`
	}
	var got string
	RunFunc("prog [flags]", "test", func(o Opts) {
		got = o.Output
	})
	if got != "out.txt" {
		t.Errorf("required flag: got %q, want out.txt", got)
	}
}

func TestRequiredFlagViaShort(t *testing.T) {
	defer setArgs([]string{"prog", "-o", "out.txt"})()
	type Opts struct {
		Output string `cli:"output file" required:"true"`
	}
	var got string
	RunFunc("prog [flags]", "test", func(o Opts) {
		got = o.Output
	})
	if got != "out.txt" {
		t.Errorf("required flag via short: got %q, want out.txt", got)
	}
}

func TestRequiredFlagViaEnvVar(t *testing.T) {
	defer setArgs([]string{"prog"})()
	t.Setenv("PROG_OUTPUT", "from-env.txt")
	type Opts struct {
		Output string `cli:"output file" required:"true"`
	}
	var got string
	RunFunc("prog [flags]", "test", func(o Opts) {
		got = o.Output
	})
	if got != "from-env.txt" {
		t.Errorf("required via env var: got %q, want from-env.txt", got)
	}
}

func TestChoicesValidRunFunc(t *testing.T) {
	defer setArgs([]string{"prog", "--format", "json"})()
	type Opts struct {
		Format string `cli:"output format" choices:"json,yaml,csv" default:"json"`
	}
	var got string
	RunFunc("prog [flags]", "test", func(o Opts) {
		got = o.Format
	})
	if got != "json" {
		t.Errorf("choices valid: got %q, want json", got)
	}
}

func TestChoicesDefaultNotTriggered(t *testing.T) {
	defer setArgs([]string{"prog"})()
	type Opts struct {
		Format string `cli:"output format" choices:"json,yaml" default:"json"`
	}
	var got string
	RunFunc("prog [flags]", "test", func(o Opts) {
		got = o.Format
	})
	if got != "json" {
		t.Errorf("choices default: got %q, want json", got)
	}
}

func TestRequiredWithChoicesValid(t *testing.T) {
	defer setArgs([]string{"prog", "--format", "yaml"})()
	type Opts struct {
		Format string `cli:"output format" required:"true" choices:"json,yaml,csv"`
	}
	var got string
	RunFunc("prog [flags]", "test", func(o Opts) {
		got = o.Format
	})
	if got != "yaml" {
		t.Errorf("required+choices: got %q, want yaml", got)
	}
}

func TestChoicesExplicitCliPath(t *testing.T) {
	defer setArgs([]string{"prog", "--format", "json"})()
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "format", Default: "json", Description: "format", Choices: []string{"json", "yaml"}}},
	}
	cfg := cli.Parse()
	if got := cfg.GetStringFlag("format"); got != "json" {
		t.Errorf("choices explicit: got %q, want json", got)
	}
}

// --- Subcommand integration tests ---

func TestDurationInSubcommand(t *testing.T) {
	defer setArgs([]string{"prog", "serve", "--timeout", "10s"})()
	type ServeOpts struct {
		Timeout time.Duration `cli:"request timeout" default:"30s"`
	}
	var got time.Duration
	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) {},
		Subcommands: Subcommands{
			NewSubcommand("serve", "start server", func(o ServeOpts) {
				got = o.Timeout
			}),
		},
	}
	cli.RunWithSubcommand()
	if got != 10*time.Second {
		t.Errorf("duration in subcommand: got %v, want 10s", got)
	}
}

func TestFlagValueInSubcommand(t *testing.T) {
	defer setArgs([]string{"prog", "run", "--level", "debug"})()
	type RunOpts struct {
		Level testLevel `cli:"log level" default:"info"`
	}
	var got string
	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) {},
		Subcommands: Subcommands{
			NewSubcommand("run", "run with logging", func(o RunOpts) {
				got = o.Level.val
			}),
		},
	}
	cli.RunWithSubcommand()
	if got != "debug" {
		t.Errorf("flag.Value in subcommand: got %q, want debug", got)
	}
}

func TestFlagValueExplicitCliPath(t *testing.T) {
	defer setArgs([]string{"prog", "--level", "error"})()
	lv := &testLevel{val: "info"}
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "level", Default: lv, Description: "log level"}},
	}
	cfg := cli.Parse()
	got := cfg.Flags["level"].(*testLevel)
	if got.val != "error" {
		t.Errorf("flag.Value explicit: got %q, want error", got.val)
	}
}

// TestJSONSchemaFlagIntegration verifies --json-schema outputs valid JSON and exits.
// We test the method directly since --json-schema calls os.Exit(0).
func TestJSONSchemaFlagIntegration(t *testing.T) {
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test tool",
		Flags: Flags{
			{Name: "output", Default: "", Description: "output file", Required: true},
			{Name: "format", Default: "json", Description: "format", Choices: []string{"json", "yaml"}},
		},
	}
	s := cli.JSONSchemaString()
	if !strings.Contains(s, `"required"`) {
		t.Error("JSON schema should contain required array")
	}
	if !strings.Contains(s, `"enum"`) {
		t.Error("JSON schema should contain enum for choices")
	}
	if !strings.Contains(s, `"output"`) {
		t.Error("JSON schema should contain output property")
	}
}

func TestRequiredSubcommandProvided(t *testing.T) {
	defer setArgs([]string{"prog", "deploy", "--target", "prod"})()
	type DeployOpts struct {
		Target string `cli:"deploy target" required:"true"`
	}
	var got string
	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) {},
		Subcommands: Subcommands{
			NewSubcommand("deploy", "deploy to target", func(o DeployOpts) {
				got = o.Target
			}),
		},
	}
	cli.RunWithSubcommand()
	if got != "prod" {
		t.Errorf("required in subcommand: got %q, want prod", got)
	}
}

func TestChoicesInSubcommandValid(t *testing.T) {
	defer setArgs([]string{"prog", "deploy", "--env", "staging"})()
	type DeployOpts struct {
		Env string `cli:"environment" choices:"dev,staging,prod" default:"dev"`
	}
	var got string
	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) {},
		Subcommands: Subcommands{
			NewSubcommand("deploy", "deploy somewhere", func(o DeployOpts) {
				got = o.Env
			}),
		},
	}
	cli.RunWithSubcommand()
	if got != "staging" {
		t.Errorf("choices in subcommand: got %q, want staging", got)
	}
}

func TestRequiredViaEnvVarInParse(t *testing.T) {
	// Ensure env var satisfies required even through the Cli.Parse() path.
	defer setArgs([]string{"prog"})()
	t.Setenv("PROG_NAME", "from-env")
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "name", Default: "", Description: "name", Required: true}},
	}
	cfg := cli.Parse()
	if got := cfg.GetStringFlag("name"); got != "from-env" {
		t.Errorf("required via env (Parse): got %q, want from-env", got)
	}
}

func TestShortNameFlagWasSet(t *testing.T) {
	// Verify flagWasSet detects short-name usage.
	set := map[string]bool{"o": true}
	f := Flag{Name: "output", ShortName: "o"}
	if !flagWasSet(f, set) {
		t.Error("flagWasSet should return true when short name is in set")
	}
}

func TestFlagWasSetAutoShort(t *testing.T) {
	// Auto-derived short (first letter of name).
	set := map[string]bool{"o": true}
	f := Flag{Name: "output"}
	if !flagWasSet(f, set) {
		t.Error("flagWasSet should return true for auto-derived short")
	}
}

func TestFlagWasSetNoShortName(t *testing.T) {
	set := map[string]bool{"o": true}
	f := Flag{Name: "output", NoShortName: true}
	if flagWasSet(f, set) {
		t.Error("flagWasSet should return false when NoShortName is true")
	}
}

func TestFlagWasSetNotSet(t *testing.T) {
	set := map[string]bool{}
	f := Flag{Name: "output"}
	if flagWasSet(f, set) {
		t.Error("flagWasSet should return false when nothing is set")
	}
}
