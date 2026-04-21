package quicli

import (
	"reflect"
	"testing"
)

// --- flagsFromStruct tests ---

type basicOpts struct {
	Count  int      `cli:"how many times" default:"3"`
	Say    string   `cli:"what to say" default:"hello"`
	World  bool     `cli:"announce to the world"`
	Ratio  float64  `cli:"scaling ratio" default:"1.5"`
	Files  []string `cli:"input files"`
	Ignore string   // no cli tag — must be skipped
}

func TestFlagsFromStructNames(t *testing.T) {
	flags, err := flagsFromStruct(reflect.TypeOf(basicOpts{}))
	if err != nil {
		t.Fatal(err)
	}
	if len(flags) != 5 {
		t.Fatalf("got %d flags, want 5 (Ignore must be skipped)", len(flags))
	}
	names := map[string]bool{}
	for _, f := range flags {
		names[f.Name] = true
	}
	for _, want := range []string{"count", "say", "world", "ratio", "files"} {
		if !names[want] {
			t.Errorf("missing flag %q", want)
		}
	}
}

func TestFlagsFromStructDefaults(t *testing.T) {
	flags, err := flagsFromStruct(reflect.TypeOf(basicOpts{}))
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Flag{}
	for _, f := range flags {
		byName[f.Name] = f
	}
	if byName["count"].Default.(int) != 3 {
		t.Errorf("count default: got %v, want 3", byName["count"].Default)
	}
	if byName["say"].Default.(string) != "hello" {
		t.Errorf("say default: got %v, want hello", byName["say"].Default)
	}
	if byName["world"].Default.(bool) != false {
		t.Errorf("world default: got %v, want false", byName["world"].Default)
	}
	if byName["ratio"].Default.(float64) != 1.5 {
		t.Errorf("ratio default: got %v, want 1.5", byName["ratio"].Default)
	}
}

func TestFlagsFromStructTags(t *testing.T) {
	type tagged struct {
		Name string `cli:"your name" short:"n" env:"MY_NAME"`
	}
	flags, err := flagsFromStruct(reflect.TypeOf(tagged{}))
	if err != nil {
		t.Fatal(err)
	}
	if len(flags) != 1 {
		t.Fatalf("got %d flags, want 1", len(flags))
	}
	f := flags[0]
	if f.ShortName != "n" {
		t.Errorf("ShortName: got %q, want n", f.ShortName)
	}
	if f.EnvVar != "MY_NAME" {
		t.Errorf("EnvVar: got %q, want MY_NAME", f.EnvVar)
	}
}

func TestFlagsFromStructBadDefault(t *testing.T) {
	type bad struct {
		Count int `cli:"count" default:"notanint"`
	}
	_, err := flagsFromStruct(reflect.TypeOf(bad{}))
	if err == nil {
		t.Error("expected error for invalid default int")
	}
}

// --- RunFunc end-to-end tests ---

func TestRunFuncInt(t *testing.T) {
	defer setArgs([]string{"prog", "--count", "5"})()
	type Opts struct {
		Count int `cli:"how many times" default:"1"`
	}
	var got int
	RunFunc("prog [flags]", "test", func(o Opts) {
		got = o.Count
	})
	if got != 5 {
		t.Errorf("RunFunc int: got %d, want 5", got)
	}
}

func TestRunFuncString(t *testing.T) {
	defer setArgs([]string{"prog", "--say", "world"})()
	type Opts struct {
		Say string `cli:"what to say" default:"hello"`
	}
	var got string
	RunFunc("prog [flags]", "test", func(o Opts) {
		got = o.Say
	})
	if got != "world" {
		t.Errorf("RunFunc string: got %q, want world", got)
	}
}

func TestRunFuncDefaultUsed(t *testing.T) {
	defer setArgs([]string{"prog"})()
	type Opts struct {
		Say string `cli:"what to say" default:"hello"`
	}
	var got string
	RunFunc("prog [flags]", "test", func(o Opts) {
		got = o.Say
	})
	if got != "hello" {
		t.Errorf("RunFunc default: got %q, want hello", got)
	}
}

func TestRunFuncBool(t *testing.T) {
	defer setArgs([]string{"prog", "--verbose"})()
	type Opts struct {
		Verbose bool `cli:"enable verbose output"`
	}
	var got bool
	RunFunc("prog [flags]", "test", func(o Opts) {
		got = o.Verbose
	})
	if !got {
		t.Error("RunFunc bool: expected true")
	}
}

func TestRunFuncSkipsUntaggedFields(t *testing.T) {
	defer setArgs([]string{"prog"})()
	type Opts struct {
		Count    int    `cli:"count" default:"1"`
		Internal string // no tag — must be skipped, zero value
	}
	var gotInternal string
	RunFunc("prog [flags]", "test", func(o Opts) {
		gotInternal = o.Internal
	})
	if gotInternal != "" {
		t.Errorf("untagged field should be zero, got %q", gotInternal)
	}
}

func TestRunFuncFloat(t *testing.T) {
	defer setArgs([]string{"prog", "--ratio", "2.5"})()
	type Opts struct {
		Ratio float64 `cli:"scaling ratio" default:"1.0"`
	}
	var got float64
	RunFunc("prog [flags]", "test", func(o Opts) {
		got = o.Ratio
	})
	if got != 2.5 {
		t.Errorf("RunFunc float64: got %f, want 2.5", got)
	}
}

func TestRunFuncSlice(t *testing.T) {
	defer setArgs([]string{"prog", "--file", "a", "--file", "b"})()
	type Opts struct {
		File []string `cli:"input files"`
	}
	var got []string
	RunFunc("prog [flags]", "test", func(o Opts) {
		got = o.File
	})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("RunFunc slice: got %v, want [a b]", got)
	}
}

// --- NewSubcommand tests ---

func TestNewSubcommandMetadata(t *testing.T) {
	type ColorOpts struct {
		Foreground bool `cli:"use foreground color"`
		Repeat     int  `cli:"how many times" default:"1"`
	}
	sub := NewSubcommand("color", "print coloured message", func(o ColorOpts) {})

	if sub.Name != "color" {
		t.Errorf("Name: got %q, want color", sub.Name)
	}
	if sub.Description != "print coloured message" {
		t.Errorf("Description: got %q, want 'print coloured message'", sub.Description)
	}
	if len(sub.Flags) != 2 {
		t.Errorf("got %d flags, want 2", len(sub.Flags))
	}
	if sub.Function == nil {
		t.Error("Function must not be nil")
	}
}

func TestNewSubcommandIntegration(t *testing.T) {
	defer setArgs([]string{"prog", "greet", "--name", "World", "--count", "3"})()

	type GreetOpts struct {
		Name  string `cli:"who to greet" default:"stranger"`
		Count int    `cli:"how many times" default:"1"`
	}

	var gotName string
	var gotCount int

	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) {},
		Subcommands: Subcommands{
			NewSubcommand("greet", "greet someone", func(o GreetOpts) {
				gotName = o.Name
				gotCount = o.Count
			}),
		},
	}
	cli.RunWithSubcommand()

	if gotName != "World" {
		t.Errorf("Name: got %q, want World", gotName)
	}
	if gotCount != 3 {
		t.Errorf("Count: got %d, want 3", gotCount)
	}
}

func TestNewSubcommandDefaultsUsed(t *testing.T) {
	defer setArgs([]string{"prog", "greet"})()

	type GreetOpts struct {
		Name string `cli:"who to greet" default:"stranger"`
	}

	var gotName string

	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) {},
		Subcommands: Subcommands{
			NewSubcommand("greet", "greet someone", func(o GreetOpts) {
				gotName = o.Name
			}),
		},
	}
	cli.RunWithSubcommand()

	if gotName != "stranger" {
		t.Errorf("default: got %q, want stranger", gotName)
	}
}
