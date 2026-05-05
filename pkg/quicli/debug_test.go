package quicli

import (
	"flag"
	"strings"
	"testing"
)

func TestBuildSourceMapCLI(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("output", "", "output file")
	fs.Parse([]string{"--output", "out.txt"})

	flags := []Flag{{Name: "output", Default: "", Description: "output file"}}
	sm := buildSourceMap(flags, fs, map[string]bool{"output": true})
	if sm["output"] != "cli" {
		t.Errorf("output source: got %q, want cli", sm["output"])
	}
}

func TestBuildSourceMapEnv(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("output", "", "output file")
	fs.Parse([]string{})
	fs.Set("output", "from-env")

	flags := []Flag{{Name: "output", Default: "", Description: "output file"}}
	sm := buildSourceMap(flags, fs, map[string]bool{})
	if sm["output"] != "env" {
		t.Errorf("output source: got %q, want env", sm["output"])
	}
}

func TestBuildSourceMapDefault(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("output", "default.txt", "output file")
	fs.Parse([]string{})

	flags := []Flag{{Name: "output", Default: "default.txt", Description: "output file"}}
	sm := buildSourceMap(flags, fs, map[string]bool{})
	if sm["output"] != "default" {
		t.Errorf("output source: got %q, want default", sm["output"])
	}
}

func TestBuildSourceMapEnvOnly(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Parse([]string{})

	flags := []Flag{{Name: "secret", Default: "", Description: "secret", EnvOnly: true}}
	sm := buildSourceMap(flags, fs, map[string]bool{})
	if sm["secret"] != "default" {
		t.Errorf("secret source: got %q, want default", sm["secret"])
	}
}

func TestFormatDebugTable(t *testing.T) {
	flags := []Flag{
		{Name: "output", Default: "", Description: "output file"},
		{Name: "secret", Default: "", Description: "secret", EnvOnly: true},
	}
	values := map[string]string{"output": "out.txt", "secret": "s3cret"}
	sources := map[string]string{"output": "cli", "secret": "env (MY_SECRET)"}

	table := formatDebugTable(flags, values, sources)
	if !strings.Contains(table, "output") {
		t.Error("table should contain output flag")
	}
	if !strings.Contains(table, "out.txt") {
		t.Error("table should contain output value")
	}
	if !strings.Contains(table, "***") {
		t.Error("env-only values should be masked")
	}
	if strings.Contains(table, "s3cret") {
		t.Error("env-only value should NOT appear in plain text")
	}
	if !strings.Contains(table, "cli") {
		t.Error("table should show source")
	}
}
