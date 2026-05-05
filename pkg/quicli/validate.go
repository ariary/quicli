package quicli

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
)

// checkFlags returns validation errors for required flags and choice constraints.
// It is called after all sources (CLI, env vars) have been applied.
func checkFlags(flags []Flag, fs *flag.FlagSet) []string {
	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		set[f.Name] = true
	})

	var errs []string
	for _, f := range flags {
		// Required: flag must have been explicitly provided (CLI or env var).
		if f.Required && !flagWasSet(f, set) {
			errs = append(errs, fmt.Sprintf("required flag --%s is missing", f.Name))
		}

		// Choices: if the flag was set, its value must be one of the allowed choices.
		if len(f.Choices) > 0 {
			fl := fs.Lookup(f.Name)
			if fl != nil {
				val := fl.Value.String()
				wasSet := flagWasSet(f, set)
				// Validate if the flag was explicitly set or has a non-empty value.
				if wasSet || val != "" {
					if !slices.Contains(f.Choices, val) {
						errs = append(errs, fmt.Sprintf(
							"flag --%s: %q is not a valid choice (%s)",
							f.Name, val, strings.Join(f.Choices, ", ")))
					}
				}
			}
		}
	}
	return errs
}

// validateFlags prints validation errors and exits if any are found.
func validateFlags(flags []Flag, fs *flag.FlagSet) {
	if errs := checkFlags(flags, fs); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, QUICLI_ERROR_PREFIX+e)
		}
		os.Exit(2)
	}
}

// checkEnvOnlyFlags validates env-only flags.
// Currently checks: required env-only flags must have their env var set.
func checkEnvOnlyFlags(flags []Flag) []string {
	var errs []string
	for _, f := range flags {
		if !f.EnvOnly || !f.Required {
			continue
		}
		envKey := f.EnvVar
		if envKey == "" {
			envKey = envVarName(os.Args[0], f.Name)
		}
		if _, ok := os.LookupEnv(envKey); !ok {
			errs = append(errs, fmt.Sprintf("required env-only flag --%s is missing (set %s)", f.Name, envKey))
		}
	}
	return errs
}

// validateEnvOnlyFlags prints env-only validation errors and exits if any found.
func validateEnvOnlyFlags(flags []Flag) {
	if errs := checkEnvOnlyFlags(flags); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, QUICLI_ERROR_PREFIX+e)
		}
		os.Exit(2)
	}
}

// flagWasSet returns true if the flag was explicitly set via CLI or env var.
func flagWasSet(f Flag, set map[string]bool) bool {
	if set[f.Name] {
		return true
	}
	if f.NoShortName {
		return false
	}
	short := f.ShortName
	if short == "" && len(f.Name) > 0 {
		short = f.Name[0:1]
	}
	return short != "" && set[short]
}
