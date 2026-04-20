# PR3: RunFunc — Struct-Tag Inference Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Prerequisite:** PR1 must be merged (GetFloatFlag, GetStringSliceFlag, EnvVar field, local FlagSet, ShortName).

**Goal:** Add `RunFunc[T any]()` — a generic entry point that builds a CLI from a plain Go struct with `cli:` tags, following the spirit of nim's cligen.

**Architecture:** Single new file `runfunc.go`. `flagsFromStruct` reflects on type T to produce `[]Flag`. `populateStruct` reflects back to fill T from Config after parsing. `RunFunc` wires them together via the existing `Cli.Parse()`. No new dependencies.

**Tech Stack:** Go 1.22 (upgraded in PR1), stdlib `reflect`/`strconv`/`strings`.

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `pkg/quicli/runfunc.go` | Create | RunFunc[T], flagsFromStruct, populateStruct |
| `pkg/quicli/runfunc_test.go` | Create | full coverage of tag parsing + population |
| `README.md` | Modify | "Zero-struct" section at top |

---

### Task 1: Write tests for flagsFromStruct

**Files:**
- Create: `pkg/quicli/runfunc_test.go`

- [ ] **Step 1: Write failing tests**

Create `pkg/quicli/runfunc_test.go`:

```go
package quicli

import (
	"reflect"
	"testing"
)

// --- flagsFromStruct tests ---

type basicOpts struct {
	Count  int     `cli:"how many times" default:"3"`
	Say    string  `cli:"what to say" default:"hello"`
	World  bool    `cli:"announce to the world"`
	Ratio  float64 `cli:"scaling ratio" default:"1.5"`
	Files  []string `cli:"input files"`
	Ignore string  // no cli tag — must be skipped
}

func TestFlagsFromStructNames(t *testing.T) {
	t.Skip("implement flagsFromStruct first")
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
	t.Skip("implement flagsFromStruct first")
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
	t.Skip("implement flagsFromStruct first")
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
	t.Skip("implement flagsFromStruct first")
	type bad struct {
		Count int `cli:"count" default:"notanint"`
	}
	_, err := flagsFromStruct(reflect.TypeOf(bad{}))
	if err == nil {
		t.Error("expected error for invalid default int")
	}
}
```

- [ ] **Step 2: Remove t.Skip calls and run tests to verify they fail**

Remove the `t.Skip(...)` line from each test function, then:

```bash
go test -run "TestFlagsFromStruct" -v
```

Expected: FAIL — flagsFromStruct undefined

---

### Task 2: Implement flagsFromStruct

**Files:**
- Create: `pkg/quicli/runfunc.go`

- [ ] **Step 1: Create pkg/quicli/runfunc.go with flagsFromStruct**

```go
package quicli

import (
	"fmt"
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
//	cli:"description"   — field description (required to include the field)
//	default:"value"     — default value as string; zero value if omitted
//	short:"x"           — short flag name override
//	env:"VAR"           — env var name override (use "-" to opt out)
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
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		cliTag := field.Tag.Get("cli")
		if cliTag == "" {
			continue // skip fields without cli tag
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
			f.Default = defaultTag // empty string if tag absent
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

// populateStruct fills a new T from Config, reading each field by its lowercased name.
func populateStruct[T any](t reflect.Type, cfg Config) T {
	var opts T
	v := reflect.ValueOf(&opts).Elem()
	for i := 0; i < t.NumField(); i++ {
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
```

- [ ] **Step 2: Run flagsFromStruct tests**

```bash
go test -run "TestFlagsFromStruct" -v
```

Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add pkg/quicli/runfunc.go pkg/quicli/runfunc_test.go
git commit -m "feat: add flagsFromStruct + populateStruct for struct-tag inference"
```

---

### Task 3: Write and run RunFunc end-to-end tests

**Files:**
- Modify: `pkg/quicli/runfunc_test.go`

- [ ] **Step 1: Add end-to-end RunFunc tests**

Append to `pkg/quicli/runfunc_test.go`:

```go
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
```

- [ ] **Step 2: Run all RunFunc tests**

```bash
go test -run "TestRunFunc" -v
```

Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add pkg/quicli/runfunc_test.go
git commit -m "test: add RunFunc end-to-end tests"
```

---

### Task 4: Update README + create PR

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Add Zero-struct section at the top of README**

Insert before the existing one-liner example (keep it as the second approach):

```markdown
### Zero-struct — the cligen way

Define a struct, tag the fields, pass a function:

```go
type Opts struct {
    Count int    `cli:"how many times I want to say it" default:"1"`
    Say   string `cli:"say something"                   default:"hello"`
    World bool   `cli:"announce it to the world"`
}

func main() {
    quicli.RunFunc("SayToTheWorld [flags]", "Say Hello...", func(o Opts) {
        for i := 0; i < o.Count; i++ {
            if o.World { fmt.Print("Message for the world: ") }
            fmt.Println(o.Say)
        }
    })
}
```

Supported field types: `int`, `string`, `bool`, `float64`, `[]string`.

Tags: `cli:"desc"` (required), `default:"val"`, `short:"x"`, `env:"VAR"`.

Untagged fields are ignored — safe to mix CLI and non-CLI fields in the same struct.
```

Keep the existing `Cli` struct section below with a note:
```markdown
### Cli struct — for subcommands and full control
```

- [ ] **Step 2: Final check**

```bash
go test ./pkg/quicli/... && go build ./examples/...
```

Expected: PASS

- [ ] **Step 3: Create PR**

```bash
git add README.md
git commit -m "docs: add RunFunc zero-struct section to README"
gh pr create --title "PR3: RunFunc — struct-tag CLI inference" --body "$(cat <<'EOF'
## Summary
- New `RunFunc[T any](usage, description string, fn func(T))` entry point
- Derives flags from struct fields with `cli:` tags (description, default, short, env)
- Supports int, string, bool, float64, []string field types
- Untagged fields are ignored — safe to mix CLI and non-CLI fields
- No new dependencies — pure stdlib reflect

## Test plan
- [ ] `go test ./pkg/quicli/... -v` passes
- [ ] `RunFunc` example in README compiles and runs correctly
- [ ] Untagged fields remain zero-value in populated struct

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```
