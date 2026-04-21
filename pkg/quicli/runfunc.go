package quicli

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// RunFunc builds and runs a CLI inferred from the exported fields of T.
// T must be a struct. Fields without a `cli:` tag are ignored.
//
// Supported field types: int, string, bool, float64, []string.
//
// Tags:
//
//	cli:"description"  — field description (required to include the field)
//	default:"value"    — default value as string; zero value if omitted
//	short:"x"          — short flag name override
//	env:"VAR"          — env var name override (use "-" to opt out)
//
// Example:
//
//	type Opts struct {
//	    Count int    `cli:"how many times" default:"1"`
//	    Say   string `cli:"what to say"    default:"hello"`
//	}
//	quicli.RunFunc("prog [flags]", "Say something", func(o Opts) {
//	    for i := 0; i < o.Count; i++ { fmt.Println(o.Say) }
//	})
func RunFunc[T any](usage, description string, fn func(T)) {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() != reflect.Struct {
		fmt.Println(QUICLI_ERROR_PREFIX + "RunFunc: T must be a struct")
		return
	}

	flags, err := flagsFromStruct(t)
	if err != nil {
		fmt.Println(QUICLI_ERROR_PREFIX + err.Error())
		return
	}

	cli := &Cli{
		Usage:       usage,
		Description: description,
		Flags:       flags,
	}
	cfg := cli.Parse()
	fn(populateStruct[T](t, cfg))
}

// flagsFromStruct reflects on struct type t and returns Flag definitions
// for each exported field that has a `cli:` tag.
func flagsFromStruct(t reflect.Type) ([]Flag, error) {
	var flags []Flag
	for i := range t.NumField() {
		field := t.Field(i)
		cliTag := field.Tag.Get("cli")
		if cliTag == "" {
			continue
		}

		f := Flag{
			Name:        strings.ToLower(field.Name),
			Description: cliTag,
			ShortName:   field.Tag.Get("short"),
			EnvVar:      field.Tag.Get("env"),
		}

		defaultTag := field.Tag.Get("default")
		switch field.Type.Kind() {
		case reflect.Int:
			if defaultTag != "" {
				v, err := strconv.Atoi(defaultTag)
				if err != nil {
					return nil, fmt.Errorf("field %s: invalid default int %q: %w", field.Name, defaultTag, err)
				}
				f.Default = v
			} else {
				f.Default = 0
			}
		case reflect.String:
			f.Default = defaultTag
		case reflect.Bool:
			if defaultTag != "" {
				v, err := strconv.ParseBool(defaultTag)
				if err != nil {
					return nil, fmt.Errorf("field %s: invalid default bool %q: %w", field.Name, defaultTag, err)
				}
				f.Default = v
			} else {
				f.Default = false
			}
		case reflect.Float64:
			if defaultTag != "" {
				v, err := strconv.ParseFloat(defaultTag, 64)
				if err != nil {
					return nil, fmt.Errorf("field %s: invalid default float64 %q: %w", field.Name, defaultTag, err)
				}
				f.Default = v
			} else {
				f.Default = float64(0)
			}
		case reflect.Slice:
			if field.Type.Elem().Kind() != reflect.String {
				return nil, fmt.Errorf("field %s: only []string slices are supported", field.Name)
			}
			f.Default = []string{}
		default:
			return nil, fmt.Errorf("field %s: unsupported type %s (supported: int, string, bool, float64, []string)", field.Name, field.Type.Kind())
		}

		flags = append(flags, f)
	}
	return flags, nil
}

// NewSubcommand builds a Subcommand whose flags are inferred from the exported
// fields of T that carry a `cli:` tag — the same rules as RunFunc.
// The inferred flags become exclusive to this subcommand (Subcommand.Flags).
//
// Example:
//
//	type ColorOpts struct {
//	    Foreground bool `cli:"use foreground color"`
//	    Surprise   bool `cli:"loop forever"`
//	}
//	sub := quicli.NewSubcommand("color", "print coloured message", func(o ColorOpts) {
//	    fmt.Println("foreground:", o.Foreground)
//	})
//	// optionally add aliases: sub.Aliases = quicli.Aliases("co", "x")
func NewSubcommand[T any](name, description string, fn func(T)) Subcommand {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() != reflect.Struct {
		fmt.Println(QUICLI_ERROR_PREFIX + "NewSubcommand: T must be a struct")
		os.Exit(2)
	}

	flags, err := flagsFromStruct(t)
	if err != nil {
		fmt.Println(QUICLI_ERROR_PREFIX + err.Error())
		os.Exit(2)
	}

	return Subcommand{
		Name:        name,
		Description: description,
		Flags:       flags,
		Function: func(cfg Config) {
			fn(populateStruct[T](t, cfg))
		},
	}
}

// populateStruct fills a new T from Config, reading each field by its lowercased name.
func populateStruct[T any](t reflect.Type, cfg Config) T {
	var opts T
	v := reflect.ValueOf(&opts).Elem()
	for i := range t.NumField() {
		field := t.Field(i)
		if field.Tag.Get("cli") == "" {
			continue
		}
		name := strings.ToLower(field.Name)
		switch field.Type.Kind() {
		case reflect.Int:
			v.Field(i).SetInt(int64(cfg.GetIntFlag(name)))
		case reflect.String:
			v.Field(i).SetString(cfg.GetStringFlag(name))
		case reflect.Bool:
			v.Field(i).SetBool(cfg.GetBoolFlag(name))
		case reflect.Float64:
			v.Field(i).SetFloat(cfg.GetFloatFlag(name))
		case reflect.Slice:
			v.Field(i).Set(reflect.ValueOf(cfg.GetStringSliceFlag(name)))
		}
	}
	return opts
}
