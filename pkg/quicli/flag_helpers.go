package quicli

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"text/tabwriter"

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
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName, ev))
		cfg.Flags[shortName] = &intPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, "", ev))
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
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName, ev))
		cfg.Flags[shortName] = &strPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, "", ev))
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
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName, ev))
		cfg.Flags[shortName] = &bPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, "", ev))
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
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName, ev))
		cfg.Flags[shortName] = &floatPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, "", ev))
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
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName, ev))
		cfg.Flags[shortName] = sv
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, "", ev))
	}
	cfg.Flags[name] = sv
}

// getFlagLine returns the help line for a flag.
func getFlagLine(description string, defaultValue any, long string, short string, envVar string) string {
	defaultStr := ". (default: "
	switch v := defaultValue.(type) {
	case int:
		defaultStr += strconv.Itoa(v)
	case string:
		defaultStr += `"` + v + `"`
	case bool:
		defaultStr += strconv.FormatBool(v)
	case float64:
		defaultStr += strconv.FormatFloat(v, 'f', -1, 64)
	case []string:
		defaultStr += "[]"
	default:
		fmt.Println(QUICLI_ERROR_PREFIX+"Unknown type for default value:", defaultValue)
		os.Exit(2)
	}
	defaultStr += ")"
	if envVar != "" {
		defaultStr += " [env: " + envVar + "]"
	}
	defaultStr += "\n"

	if short == "" {
		return "--" + long + "\t\t\t" + description + defaultStr
	}
	return "--" + long + "\t-" + short + "\t\t" + description + defaultStr
}
