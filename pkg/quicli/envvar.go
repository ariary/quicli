package quicli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// envVarName derives the auto env var name from program name and flag name.
// Example: progName="say-hello", flagName="count" → "SAY_HELLO_COUNT"
func envVarName(progName, flagName string) string {
	base := filepath.Base(progName)
	sanitize := func(s string) string {
		var b strings.Builder
		for _, r := range strings.ToUpper(s) {
			if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				b.WriteRune(r)
			} else {
				b.WriteRune('_')
			}
		}
		return b.String()
	}
	return sanitize(base) + "_" + sanitize(flagName)
}

// applyEnvVars reads env vars for flags not explicitly provided on the CLI
// and updates config accordingly. Priority: CLI flag > env var > default.
func applyEnvVars(flags []Flag, fs *flag.FlagSet) {
	// Collect flags explicitly set on the command line.
	cliProvided := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		cliProvided[f.Name] = true
	})

	for _, f := range flags {
		if cliProvided[f.Name] {
			continue
		}
		// Determine the env var key to check.
		envKey := f.EnvVar
		if envKey == "-" {
			continue // opted out
		}
		if envKey == "" {
			envKey = envVarName(os.Args[0], f.Name)
		}
		val, ok := os.LookupEnv(envKey)
		if !ok {
			continue
		}
		// Apply via fs.Set so the flag package parses and stores it correctly.
		if err := fs.Set(f.Name, val); err != nil {
			fmt.Fprintf(os.Stderr, QUICLI_ERROR_PREFIX+"env var %s=%q invalid for flag --%s: %v\n", envKey, val, f.Name, err)
			os.Exit(1)
		}
	}
}
