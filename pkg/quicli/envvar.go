package quicli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
		if f.EnvOnly {
			continue // handled by applyEnvOnlyFlags
		}
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

// applyEnvOnlyFlags populates config.Flags for flags marked EnvOnly.
// These flags are never registered in the flag.FlagSet; their values come
// exclusively from environment variables (or the default).
func applyEnvOnlyFlags(flags []Flag, cfg Config) {
	for _, f := range flags {
		if !f.EnvOnly {
			continue
		}

		// Determine the env var key to check.
		envKey := f.EnvVar
		if envKey == "" {
			envKey = envVarName(os.Args[0], f.Name)
		}
		envVal, envSet := os.LookupEnv(envKey)

		// Ensure Default is not nil (same as Parse does for regular flags).
		def := f.Default
		if def == nil {
			def = false
		}

		switch def.(type) {
		case string:
			v := def.(string)
			if envSet {
				v = envVal
			}
			cfg.Flags[f.Name] = &v
		case int:
			v := def.(int)
			if envSet {
				parsed, err := strconv.Atoi(envVal)
				if err != nil {
					fmt.Fprintf(os.Stderr, QUICLI_ERROR_PREFIX+"env var %s=%q invalid for env-only flag %q (int): %v\n", envKey, envVal, f.Name, err)
					os.Exit(1)
				}
				v = parsed
			}
			cfg.Flags[f.Name] = &v
		case bool:
			v := def.(bool)
			if envSet {
				parsed, err := strconv.ParseBool(envVal)
				if err != nil {
					fmt.Fprintf(os.Stderr, QUICLI_ERROR_PREFIX+"env var %s=%q invalid for env-only flag %q (bool): %v\n", envKey, envVal, f.Name, err)
					os.Exit(1)
				}
				v = parsed
			}
			cfg.Flags[f.Name] = &v
		case float64:
			v := def.(float64)
			if envSet {
				parsed, err := strconv.ParseFloat(envVal, 64)
				if err != nil {
					fmt.Fprintf(os.Stderr, QUICLI_ERROR_PREFIX+"env var %s=%q invalid for env-only flag %q (float64): %v\n", envKey, envVal, f.Name, err)
					os.Exit(1)
				}
				v = parsed
			}
			cfg.Flags[f.Name] = &v
		case time.Duration:
			v := def.(time.Duration)
			if envSet {
				parsed, err := time.ParseDuration(envVal)
				if err != nil {
					fmt.Fprintf(os.Stderr, QUICLI_ERROR_PREFIX+"env var %s=%q invalid for env-only flag %q (duration): %v\n", envKey, envVal, f.Name, err)
					os.Exit(1)
				}
				v = parsed
			}
			cfg.Flags[f.Name] = &v
		case []string:
			v := def.([]string)
			if envSet {
				v = strings.Split(envVal, ",")
			}
			sv := &stringSliceValue{val: &v}
			cfg.Flags[f.Name] = sv
		default:
			fmt.Fprintf(os.Stderr, QUICLI_ERROR_PREFIX+"unsupported type for env-only flag %q\n", f.Name)
			os.Exit(2)
		}
	}
}
