package quicli

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	stringSlice "github.com/ariary/go-utils/pkg/stringSlice"
)

// flagEnvVarDisplay returns the env var name to show in help, or "" to suppress.
func flagEnvVarDisplay(f Flag) string {
	if f.EnvVar == "-" {
		return ""
	}
	if f.EnvVar != "" {
		return f.EnvVar
	}
	return envVarName(os.Args[0], f.Name)
}

func createIntFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet) {
	name := f.Name
	shortName := f.ShortName
	if shortName == "" {
		shortName = name[0:1]
	}
	ev := flagEnvVarDisplay(f)
	var intPtr int
	fs.IntVar(&intPtr, name, int(reflect.ValueOf(f.Default).Int()), f.Description)
	if !stringSlice.Contains(*shorts, shortName) && !f.NoShortName {
		fs.IntVar(&intPtr, shortName, int(reflect.ValueOf(f.Default).Int()), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f, shortName, ev))
		cfg.Flags[shortName] = &intPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f, "", ev))
	}
	cfg.Flags[name] = &intPtr
}

func createStringFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet) {
	name := f.Name
	shortName := f.ShortName
	if shortName == "" {
		shortName = name[0:1]
	}
	ev := flagEnvVarDisplay(f)
	var strPtr string
	fs.StringVar(&strPtr, name, reflect.ValueOf(f.Default).String(), f.Description)
	if !stringSlice.Contains(*shorts, shortName) && !f.NoShortName {
		fs.StringVar(&strPtr, shortName, reflect.ValueOf(f.Default).String(), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f, shortName, ev))
		cfg.Flags[shortName] = &strPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f, "", ev))
	}
	cfg.Flags[name] = &strPtr
}

func createBoolFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet) {
	name := f.Name
	shortName := f.ShortName
	if shortName == "" {
		shortName = name[0:1]
	}
	ev := flagEnvVarDisplay(f)
	var bPtr bool
	fs.BoolVar(&bPtr, name, reflect.ValueOf(f.Default).Bool(), f.Description)
	cfg.Flags[name] = &bPtr
	if !stringSlice.Contains(*shorts, shortName) && !f.NoShortName {
		fs.BoolVar(&bPtr, shortName, reflect.ValueOf(f.Default).Bool(), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f, shortName, ev))
		cfg.Flags[shortName] = &bPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f, "", ev))
	}
	cfg.Flags[name] = &bPtr
}

func createFloatFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet) {
	name := f.Name
	shortName := f.ShortName
	if shortName == "" {
		shortName = name[0:1]
	}
	ev := flagEnvVarDisplay(f)
	var floatPtr float64
	fs.Float64Var(&floatPtr, name, reflect.ValueOf(f.Default).Float(), f.Description)
	cfg.Flags[name] = &floatPtr
	if !stringSlice.Contains(*shorts, shortName) && !f.NoShortName {
		fs.Float64Var(&floatPtr, shortName, reflect.ValueOf(f.Default).Float(), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f, shortName, ev))
		cfg.Flags[shortName] = &floatPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f, "", ev))
	}
	cfg.Flags[name] = &floatPtr
}

func createStringSliceFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet) {
	name := f.Name
	shortName := f.ShortName
	if shortName == "" {
		shortName = name[0:1]
	}
	ev := flagEnvVarDisplay(f)
	val := []string{}
	sv := &stringSliceValue{val: &val}
	fs.Var(sv, name, f.Description)
	if !stringSlice.Contains(*shorts, shortName) && !f.NoShortName {
		fs.Var(sv, shortName, f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f, shortName, ev))
		cfg.Flags[shortName] = sv
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f, "", ev))
	}
	cfg.Flags[name] = sv
}

// getFlagLine returns the help line for a flag.
func getFlagLine(f Flag, short string, envVar string) string {
	suffix := ". "

	if f.Required {
		suffix += "(required) "
	}

	if len(f.Choices) > 0 {
		suffix += "(choices: " + strings.Join(f.Choices, ", ") + ") "
	}

	// Show default value unless the flag is required.
	if !f.Required {
		suffix += "(default: " + formatDefault(f.Default) + ") "
	}

	if envVar != "" {
		suffix += "[env: " + envVar + "]"
	}
	suffix += "\n"

	if short == "" {
		return "--" + f.Name + "\t\t\t" + f.Description + suffix
	}
	return "--" + f.Name + "\t-" + short + "\t\t" + f.Description + suffix
}

// formatDefault returns the display string for a flag's default value.
func formatDefault(defaultValue any) string {
	switch v := defaultValue.(type) {
	case int:
		return strconv.Itoa(v)
	case string:
		return `"` + v + `"`
	case bool:
		return strconv.FormatBool(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case []string:
		return "[]"
	case time.Duration:
		return v.String()
	default:
		if fv, ok := defaultValue.(flag.Value); ok {
			return `"` + fv.String() + `"`
		}
		return fmt.Sprint(defaultValue)
	}
}

func createDurationFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet) {
	name := f.Name
	shortName := f.ShortName
	if shortName == "" {
		shortName = name[0:1]
	}
	ev := flagEnvVarDisplay(f)
	var durPtr time.Duration
	fs.DurationVar(&durPtr, name, f.Default.(time.Duration), f.Description)
	if !stringSlice.Contains(*shorts, shortName) && !f.NoShortName {
		fs.DurationVar(&durPtr, shortName, f.Default.(time.Duration), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f, shortName, ev))
		cfg.Flags[shortName] = &durPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f, "", ev))
	}
	cfg.Flags[name] = &durPtr
}

func createValueFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet) {
	name := f.Name
	shortName := f.ShortName
	if shortName == "" {
		shortName = name[0:1]
	}
	ev := flagEnvVarDisplay(f)
	val := f.Default.(flag.Value)
	fs.Var(val, name, f.Description)
	if !stringSlice.Contains(*shorts, shortName) && !f.NoShortName {
		fs.Var(val, shortName, f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f, shortName, ev))
		cfg.Flags[shortName] = val
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f, "", ev))
	}
	cfg.Flags[name] = val
}
