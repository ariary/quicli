package quicli

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// buildSourceMap determines where each flag's value came from.
// cliSet contains flag names explicitly set on the command line (snapshot from
// fs.Visit after fs.Parse, before applyEnvVars). Flags set in fs after that
// (via applyEnvVars calling fs.Set) came from env vars.
func buildSourceMap(flags []Flag, fs *flag.FlagSet, cliSet map[string]bool) map[string]string {
	allSet := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		allSet[f.Name] = true
	})

	sm := make(map[string]string, len(flags))
	for _, f := range flags {
		if f.EnvOnly {
			sm[f.Name] = "default"
			continue
		}
		if cliSet[f.Name] {
			sm[f.Name] = "cli"
		} else if allSet[f.Name] {
			sm[f.Name] = "env"
		} else {
			sm[f.Name] = "default"
		}
	}
	return sm
}

// flagValue returns the string representation of a flag's current value from config.
func flagValue(f Flag, cfg Config) string {
	v, ok := cfg.Flags[f.Name]
	if !ok {
		return formatDefault(f.Default)
	}
	switch ptr := v.(type) {
	case *int:
		return fmt.Sprintf("%d", *ptr)
	case *string:
		return *ptr
	case *bool:
		return fmt.Sprintf("%v", *ptr)
	case *float64:
		return fmt.Sprintf("%g", *ptr)
	case *stringSliceValue:
		return ptr.String()
	default:
		if fv, ok := v.(flag.Value); ok {
			return fv.String()
		}
		return fmt.Sprint(v)
	}
}

// formatDebugTable returns a formatted table of flag names, values, and sources.
// Env-only flag values are masked as "***".
func formatDebugTable(flags []Flag, values map[string]string, sources map[string]string) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 2, 8, 2, ' ', 0)
	fmt.Fprintf(w, "FLAG\tVALUE\tSOURCE\n")
	for _, f := range flags {
		val := values[f.Name]
		if f.EnvOnly {
			val = "***"
		}
		src := sources[f.Name]
		fmt.Fprintf(w, "--%s\t%s\t%s\n", f.Name, val, src)
	}
	w.Flush()
	return b.String()
}

// printDebugOptions collects values and sources, then prints the debug table and exits.
func printDebugOptions(flags []Flag, cfg Config, sources map[string]string) {
	values := make(map[string]string, len(flags))
	for _, f := range flags {
		values[f.Name] = flagValue(f, cfg)
	}

	// Enrich env-only sources with actual env var name.
	for _, f := range flags {
		if !f.EnvOnly {
			continue
		}
		envKey := f.EnvVar
		if envKey == "" {
			envKey = envVarName(os.Args[0], f.Name)
		}
		if _, ok := os.LookupEnv(envKey); ok {
			sources[f.Name] = "env (" + envKey + ")"
		}
	}

	// Enrich regular env sources with the env var name.
	for _, f := range flags {
		if sources[f.Name] == "env" {
			envKey := flagEnvVarDisplay(f)
			if envKey != "" {
				sources[f.Name] = "env (" + envKey + ")"
			}
		}
	}

	fmt.Print(formatDebugTable(flags, values, sources))
	os.Exit(0)
}
