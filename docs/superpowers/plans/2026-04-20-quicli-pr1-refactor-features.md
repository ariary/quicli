# PR1: Internal Refactor + Missing Features Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Unify the duplicated flag-creation helpers, eliminate global state, and add GetFloatFlag, ShortName, SharedSubcommand rename, per-subcommand Flags, []string slice flag, and misspelled subcommand detection.

**Architecture:** Extract all createXxxFlag helpers and getFlagLine into `flag_helpers.go` with a `*flag.FlagSet` parameter. Both `Parse()` and `RunWithSubcommand()` create a local `flag.FlagSet` and delegate to these helpers. New features (slice, levenshtein) each get their own focused file.

**Tech Stack:** Go 1.22 (upgraded in Task 0), stdlib `flag`/`reflect`/`os`, `github.com/ariary/go-utils/pkg/stringSlice`, `github.com/deckarep/golang-set/v2`

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `pkg/quicli/quicli.go` | Modify | Parse() uses local FlagSet, EnvVar field, ShortName field |
| `pkg/quicli/subcommand.go` | Modify | RunWithSubcommand() uses helpers, AllAliases local, Subcommand.Flags wired |
| `pkg/quicli/flag_helpers.go` | Create | createIntFlag/String/Bool/Float/Slice + getFlagLine, all take *flag.FlagSet |
| `pkg/quicli/slice_flag.go` | Create | stringSliceValue implementing flag.Value |
| `pkg/quicli/levenshtein.go` | Create | levenshtein(), findClosestSubcommand() |
| `pkg/quicli/quicli_test.go` | Create | tests for Parse path |
| `pkg/quicli/subcommand_test.go` | Create | tests for subcommand path |
| `pkg/quicli/levenshtein_test.go` | Create | tests for levenshtein |
| `examples/sayhello_subcommand.go` | Modify | ForSubcommand → SharedSubcommand rename |
| `README.md` | Modify | New fields, renamed field, migration note |

---

### Task 0: Upgrade Go version to 1.22

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Update go.mod**

Change the `go` directive in `go.mod` from `go 1.18` to `go 1.22`.

```
go 1.22
```

- [ ] **Step 2: Tidy and verify**

```bash
go mod tidy && go build ./...
```

Expected: builds cleanly. No dependencies change — this is a minimum version bump only.

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: upgrade Go minimum version to 1.22"
```

> **Why 1.22:** built-in `min()`/`max()`, loop-variable-per-iteration semantics, and general ecosystem alignment. Used in levenshtein.go (Task 9) to replace the custom `minOf3` helper.

---

### Task 1: Test infrastructure + GetFloatFlag

**Files:**
- Create: `pkg/quicli/quicli_test.go`
- Modify: `pkg/quicli/quicli.go`

- [ ] **Step 1: Write the failing test**

Create `pkg/quicli/quicli_test.go`:

```go
package quicli

import (
	"os"
	"testing"
)

// setArgs temporarily overrides os.Args and returns a restore func.
func setArgs(args []string) func() {
	old := os.Args
	os.Args = args
	return func() { os.Args = old }
}

func TestGetFloatFlag(t *testing.T) {
	defer setArgs([]string{"prog", "--ratio", "3.14"})()
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "ratio", Default: float64(0), Description: "a ratio"}},
	}
	cfg := cli.Parse()
	if got := cfg.GetFloatFlag("ratio"); got != 3.14 {
		t.Errorf("GetFloatFlag: got %f, want 3.14", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd pkg/quicli && go test -run TestGetFloatFlag -v
```

Expected: `FAIL — cfg.GetFloatFlag undefined`

- [ ] **Step 3: Add GetFloatFlag to quicli.go**

In `pkg/quicli/quicli.go`, after `GetBoolFlag`:

```go
// GetFloatFlag returns the float64 value of a float64 flag.
func (c Config) GetFloatFlag(name string) float64 {
	elem := c.Flags[name]
	if elem == nil {
		fmt.Println(QUICLI_ERROR_PREFIX, "failed to retrieve value for flag:", name)
		os.Exit(92)
	}
	f := reflect.ValueOf(elem).Interface().(*float64)
	return *f
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test -run TestGetFloatFlag -v
```

Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add pkg/quicli/quicli.go pkg/quicli/quicli_test.go
git commit -m "feat: add GetFloatFlag + test infrastructure"
```

---

### Task 2: Add ShortName field to Flag

**Files:**
- Modify: `pkg/quicli/quicli.go`

- [ ] **Step 1: Write the failing test**

Add to `pkg/quicli/quicli_test.go`:

```go
func TestFlagCustomShortName(t *testing.T) {
	defer setArgs([]string{"prog", "-r", "42"})()
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "count", Default: 0, Description: "test", ShortName: "r"}},
	}
	cfg := cli.Parse()
	if got := cfg.GetIntFlag("count"); got != 42 {
		t.Errorf("custom ShortName: got %d, want 42", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test -run TestFlagCustomShortName -v
```

Expected: `FAIL — unknown field ShortName`

- [ ] **Step 3: Add ShortName to Flag struct**

In `pkg/quicli/quicli.go`, update the `Flag` struct:

```go
type Flag struct {
	Name              string
	Description       string
	Default           interface{}
	NoShortName       bool
	ShortName         string   // overrides auto first-letter derivation
	NotForRootCommand bool
	ForSubcommand     SubcommandSet
	EnvVar            string   // env var override (activated in PR2)
}
```

- [ ] **Step 4: Update createIntFlag, createStringFlag, createBoolFlag, createFloatFlag in quicli.go to use ShortName**

In each `createXxxFlag` function in `quicli.go`, replace the `shortName := name[0:1]` line with:

```go
shortName := f.ShortName
if shortName == "" {
    shortName = name[0:1]
}
```

Do this for all four functions: `createIntFlag`, `createStringFlag`, `createBoolFlag`, `createFloatFlag`.

Do the same in the four `createXxxFlagFs` functions in `subcommand.go`.

- [ ] **Step 5: Run tests**

```bash
go test ./pkg/quicli/... -v
```

Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add pkg/quicli/quicli.go pkg/quicli/subcommand.go pkg/quicli/quicli_test.go
git commit -m "feat: add ShortName field to Flag for custom short-name override"
```

---

### Task 3: Rename ForSubcommand → SharedSubcommand

**Files:**
- Modify: `pkg/quicli/quicli.go`
- Modify: `pkg/quicli/subcommand.go`
- Modify: `examples/sayhello_subcommand.go`
- Modify: `README.md`

- [ ] **Step 1: Rename field in Flag struct**

In `pkg/quicli/quicli.go`, in the `Flag` struct, rename `ForSubcommand SubcommandSet` to `SharedSubcommand SubcommandSet`.

- [ ] **Step 2: Update all usages in subcommand.go**

In `pkg/quicli/subcommand.go`, replace every occurrence of `f.ForSubcommand` with `f.SharedSubcommand`.

There are four occurrences — the validation loop at lines ~89-95, the alias expansion loop at lines ~99-106, and the two `f.isForSubcommand` calls.

Also update the `isForSubcommand` method receiver body (it iterates `f.ForSubcommand`) — rename to `f.SharedSubcommand` there too.

- [ ] **Step 3: Update examples/sayhello_subcommand.go**

Replace `ForSubcommand:` with `SharedSubcommand:` in the flag definitions (two occurrences).

- [ ] **Step 4: Run tests**

```bash
go test ./pkg/quicli/... -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/quicli/quicli.go pkg/quicli/subcommand.go examples/sayhello_subcommand.go
git commit -m "refactor: rename ForSubcommand to SharedSubcommand for clarity"
```

---

### Task 4: Extract flag_helpers.go (unified FlagSet-based helpers)

**Files:**
- Create: `pkg/quicli/flag_helpers.go`
- Modify: `pkg/quicli/quicli.go`
- Modify: `pkg/quicli/subcommand.go`

- [ ] **Step 1: Create pkg/quicli/flag_helpers.go**

```go
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

func createIntFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet) {
	name := f.Name
	shortName := f.ShortName
	if shortName == "" {
		shortName = name[0:1]
	}
	var intPtr int
	fs.IntVar(&intPtr, name, int(reflect.ValueOf(f.Default).Int()), f.Description)
	if !stringSlice.Contains(*shorts, shortName) && !f.NoShortName {
		fs.IntVar(&intPtr, shortName, int(reflect.ValueOf(f.Default).Int()), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName))
		cfg.Flags[shortName] = &intPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, ""))
	}
	cfg.Flags[name] = &intPtr
}

func createStringFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet) {
	name := f.Name
	shortName := f.ShortName
	if shortName == "" {
		shortName = name[0:1]
	}
	var strPtr string
	fs.StringVar(&strPtr, name, reflect.ValueOf(f.Default).String(), f.Description)
	if !stringSlice.Contains(*shorts, shortName) && !f.NoShortName {
		fs.StringVar(&strPtr, shortName, reflect.ValueOf(f.Default).String(), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName))
		cfg.Flags[shortName] = &strPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, ""))
	}
	cfg.Flags[name] = &strPtr
}

func createBoolFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet) {
	name := f.Name
	shortName := f.ShortName
	if shortName == "" {
		shortName = name[0:1]
	}
	var bPtr bool
	fs.BoolVar(&bPtr, name, reflect.ValueOf(f.Default).Bool(), f.Description)
	cfg.Flags[name] = &bPtr
	if !stringSlice.Contains(*shorts, shortName) && !f.NoShortName {
		fs.BoolVar(&bPtr, shortName, reflect.ValueOf(f.Default).Bool(), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName))
		cfg.Flags[shortName] = &bPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, ""))
	}
	cfg.Flags[name] = &bPtr
}

func createFloatFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet) {
	name := f.Name
	shortName := f.ShortName
	if shortName == "" {
		shortName = name[0:1]
	}
	var floatPtr float64
	fs.Float64Var(&floatPtr, name, reflect.ValueOf(f.Default).Float(), f.Description)
	cfg.Flags[name] = &floatPtr
	if !stringSlice.Contains(*shorts, shortName) && !f.NoShortName {
		fs.Float64Var(&floatPtr, shortName, reflect.ValueOf(f.Default).Float(), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName))
		cfg.Flags[shortName] = &floatPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, ""))
	}
	cfg.Flags[name] = &floatPtr
}

// getFlagLine returns the help line for a flag.
func getFlagLine(description string, defaultValue interface{}, long string, short string) string {
	defaultStr := ". (default: "
	switch v := defaultValue.(type) {
	case int:
		defaultStr += strconv.Itoa(v) + ")\n"
	case string:
		defaultStr += `"` + v + `")` + "\n"
	case bool:
		defaultStr += strconv.FormatBool(v) + ")\n"
	case float64:
		defaultStr += strconv.FormatFloat(v, 'f', -1, 64) + ")\n"
	case []string:
		defaultStr += "[])\n"
	default:
		fmt.Println(QUICLI_ERROR_PREFIX+"Unknown type for default value:", defaultValue)
		os.Exit(2)
	}
	if short == "" {
		return "--" + long + "\t\t\t" + description + defaultStr
	}
	return "--" + long + "\t-" + short + "\t\t" + description + defaultStr
}
```

- [ ] **Step 2: Remove duplicate functions from quicli.go**

Delete the four `createIntFlag`, `createStringFlag`, `createBoolFlag`, `createFloatFlag` functions and the `getFlagLine` function from `pkg/quicli/quicli.go`. Also remove their now-unused imports (`strconv` may still be needed elsewhere — check first; `reflect` is still needed for GetXxxFlag methods).

- [ ] **Step 3: Remove duplicate Fs-suffix functions from subcommand.go**

Delete `createIntFlagFs`, `createStringFlagFs`, `createBoolFlagFs`, `createFloatFlagFs` from `pkg/quicli/subcommand.go`.

- [ ] **Step 4: Update subcommand.go to call the unified helpers**

In `RunWithSubcommand()`, the four switch cases still call `createXxxFlagFs(...)` — update them to call `createXxxFlag(...)` (same signature, now unified in flag_helpers.go).

- [ ] **Step 5: Run tests**

```bash
go test ./pkg/quicli/... -v
```

Expected: PASS. Also verify the examples still compile:

```bash
go build ./examples/...
```

- [ ] **Step 6: Commit**

```bash
git add pkg/quicli/flag_helpers.go pkg/quicli/quicli.go pkg/quicli/subcommand.go
git commit -m "refactor: extract unified createXxxFlag helpers into flag_helpers.go"
```

---

### Task 5: Refactor Parse() to use local FlagSet

**Files:**
- Modify: `pkg/quicli/quicli.go`

- [ ] **Step 1: Add a test that will catch global-state leakage**

Add to `pkg/quicli/quicli_test.go`:

```go
func TestParseDoesNotLeakGlobalState(t *testing.T) {
	// Calling Parse() twice with different args must not panic or produce
	// "flag redefined" errors — only possible with a local FlagSet.
	defer setArgs([]string{"prog", "--count", "1"})()
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "count", Default: 0, Description: "count"}},
	}
	// If Parse() uses the global flag set, second call panics.
	_ = cli.Parse()

	os.Args = []string{"prog", "--count", "2"}
	cfg2 := cli.Parse()
	if got := cfg2.GetIntFlag("count"); got != 2 {
		t.Errorf("second Parse: got %d, want 2", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test -run TestParseDoesNotLeakGlobalState -v
```

Expected: FAIL or panic with "flag redefined"

- [ ] **Step 3: Rewrite Parse() to use a local FlagSet**

Replace the `Parse()` method in `pkg/quicli/quicli.go` with:

```go
func (c *Cli) Parse() (config Config) {
	usage := new(strings.Builder)
	wUsage := new(tabwriter.Writer)
	wUsage.Init(usage, 2, 8, 1, '\t', 1)
	var shorts []string
	config.Flags = make(map[string]interface{})
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	fmt.Fprintf(wUsage, color.Yellow(c.Description)+"\n\nUsage: "+c.Usage+"\n\n")

	for i := 0; i < len(c.Flags); i++ {
		f := c.Flags[i]
		if len(f.Name) == 0 {
			fmt.Println(QUICLI_ERROR_PREFIX + "empty flag name definition")
			os.Exit(2)
		}
		if f.Default == nil {
			f.Default = false
		}
		switch f.Default.(type) {
		case int:
			createIntFlag(config, f, &shorts, wUsage, fs)
		case string:
			createStringFlag(config, f, &shorts, wUsage, fs)
		case bool:
			createBoolFlag(config, f, &shorts, wUsage, fs)
		case float64:
			createFloatFlag(config, f, &shorts, wUsage, fs)
		case []string:
			createStringSliceFlag(config, f, &shorts, wUsage, fs)
		default:
			fmt.Println(QUICLI_ERROR_PREFIX+"Unknown flag type:", f.Default)
			os.Exit(2)
		}
	}
	fmt.Fprintf(wUsage, "\nUse \""+color.Yellow(os.Args[0])+" --help\" for more information about the command.\n")

	var cheatSheet bool
	if len(c.CheatSheet) > 0 {
		fmt.Fprintf(wUsage, "\nSee command examples with \""+color.Yellow(os.Args[0])+" --cheat-sheet\"\n")
		fs.BoolVar(&cheatSheet, "cheat-sheet", false, "print cheat sheet")
		fs.BoolVar(&cheatSheet, "cs", false, "print cheat sheet")
	}

	wUsage.Flush()
	fs.Usage = func() { fmt.Print(usage.String()) }
	fs.Parse(os.Args[1:])
	config.Args = fs.Args()

	if len(c.CheatSheet) > 0 && cheatSheet {
		c.PrintCheatSheet()
		os.Exit(0)
	}

	return config
}
```

Note: `createStringSliceFlag` is added in Task 8 — add a `// []string case` comment for now if needed, or implement Task 8 before this one.

- [ ] **Step 4: Clean up now-unused imports in quicli.go**

Remove the `"flag"` import if it is no longer referenced directly. Verify with:

```bash
go build ./pkg/quicli/
```

- [ ] **Step 5: Run all tests**

```bash
go test ./pkg/quicli/... -v
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add pkg/quicli/quicli.go pkg/quicli/quicli_test.go
git commit -m "refactor: Parse() uses local FlagSet, eliminates global flag state"
```

---

### Task 6: Refactor RunWithSubcommand() + local AllAliases

**Files:**
- Modify: `pkg/quicli/subcommand.go`

- [ ] **Step 1: Remove the package-level AllAliases var**

Delete this line from `pkg/quicli/subcommand.go`:

```go
var AllAliases mapset.Set[string]
```

- [ ] **Step 2: Update checkSubcommandAliasesUniqueness signature**

Change the function to take the set as a parameter instead of using the global:

```go
func checkSubcommandAliasesUniqueness(c *Cli, allAliases mapset.Set[string]) {
	for i := 0; i < len(c.Subcommands); i++ {
		subcommandAliases := c.Subcommands[i].Aliases
		if subcommandAliases != nil {
			commonAliases := allAliases.Intersect(subcommandAliases)
			if commonAliases.Cardinality() == 0 {
				allAliases.Append(subcommandAliases.ToSlice()...)
			} else {
				fmt.Println(QUICLI_ERROR_PREFIX+"subcommand", c.Subcommands[i].Name, "define some already defined aliases ('", strings.Join(commonAliases.ToSlice(), ","), "')")
				os.Exit(2)
			}
		}
	}
}
```

- [ ] **Step 3: Update RunWithSubcommand() to pass local allAliases**

Inside `RunWithSubcommand()`, replace:

```go
AllAliases = mapset.NewSet[string]()
checkSubcommandAliasesUniqueness(c)
```

with:

```go
allAliases := mapset.NewSet[string]()
checkSubcommandAliasesUniqueness(c, allAliases)
```

- [ ] **Step 4: Run tests**

```bash
go test ./pkg/quicli/... -v && go build ./examples/...
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/quicli/subcommand.go
git commit -m "refactor: make AllAliases local to RunWithSubcommand, remove global var"
```

---

### Task 7: Add Flags []Flag to Subcommand + wire up

**Files:**
- Modify: `pkg/quicli/subcommand.go`
- Create: `pkg/quicli/subcommand_test.go`

- [ ] **Step 1: Write the failing test**

Create `pkg/quicli/subcommand_test.go`:

```go
package quicli

import (
	"testing"
)

func TestSubcommandExclusiveFlag(t *testing.T) {
	defer setArgs([]string{"prog", "greet", "--name", "World"})()
	var receivedName string
	cli := Cli{
		Usage:       "prog [command]",
		Description: "test",
		Function:    func(cfg Config) {},
		Subcommands: Subcommands{
			{
				Name:        "greet",
				Description: "greet someone",
				Flags:       Flags{{Name: "name", Default: "", Description: "who to greet"}},
				Function: func(cfg Config) {
					receivedName = cfg.GetStringFlag("name")
				},
			},
		},
	}
	cli.RunWithSubcommand()
	if receivedName != "World" {
		t.Errorf("subcommand Flag: got %q, want World", receivedName)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test -run TestSubcommandExclusiveFlag -v
```

Expected: FAIL — unknown field Flags on Subcommand

- [ ] **Step 3: Add Flags to Subcommand struct**

In `pkg/quicli/subcommand.go`, update the `Subcommand` struct:

```go
type Subcommand struct {
	Name        string
	Aliases     mapset.Set[string]
	Description string
	Function    Runner
	Flags       []Flag // flags exclusive to this subcommand
}
```

- [ ] **Step 4: Wire Subcommand.Flags into RunWithSubcommand()**

In `RunWithSubcommand()`, after the existing flags loop (which handles `c.Flags`), add a block that registers the active subcommand's exclusive flags:

```go
// Register exclusive flags for the active subcommand
if !isRootCommand(c.Subcommands) {
	sub := getSubcommandByName(c.Subcommands, os.Args[1])
	for i := 0; i < len(sub.Flags); i++ {
		f := sub.Flags[i]
		if len(f.Name) == 0 {
			fmt.Println(QUICLI_ERROR_PREFIX + "empty flag name definition in subcommand " + sub.Name)
			os.Exit(2)
		}
		if f.Default == nil {
			f.Default = false
		}
		switch f.Default.(type) {
		case int:
			createIntFlag(config, f, &shorts, wUsage, fs)
		case string:
			createStringFlag(config, f, &shorts, wUsage, fs)
		case bool:
			createBoolFlag(config, f, &shorts, wUsage, fs)
		case float64:
			createFloatFlag(config, f, &shorts, wUsage, fs)
		case []string:
			createStringSliceFlag(config, f, &shorts, wUsage, fs)
		default:
			fmt.Println(QUICLI_ERROR_PREFIX+"Unknown flag type:", f.Default)
			os.Exit(2)
		}
	}
}
```

Place this block after the existing `c.Flags` loop and before the cheat-sheet block.

- [ ] **Step 5: Run tests**

```bash
go test ./pkg/quicli/... -v
```

Expected: PASS (note: `createStringSliceFlag` will cause a compile error until Task 8 — you may add a stub in `slice_flag.go` now, or swap Tasks 7 and 8)

- [ ] **Step 6: Commit**

```bash
git add pkg/quicli/subcommand.go pkg/quicli/subcommand_test.go
git commit -m "feat: add Flags field to Subcommand for subcommand-exclusive flags"
```

---

### Task 8: Add []string slice flag support

**Files:**
- Create: `pkg/quicli/slice_flag.go`
- Modify: `pkg/quicli/flag_helpers.go`
- Modify: `pkg/quicli/quicli.go` (Config method)
- Modify: `pkg/quicli/quicli_test.go`

- [ ] **Step 1: Write the failing test**

Add to `pkg/quicli/quicli_test.go`:

```go
func TestStringSliceFlag(t *testing.T) {
	defer setArgs([]string{"prog", "--file", "a", "--file", "b,c"})()
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "file", Default: []string{}, Description: "files"}},
	}
	cfg := cli.Parse()
	got := cfg.GetStringSliceFlag("file")
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("GetStringSliceFlag: got %v, want [a b c]", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test -run TestStringSliceFlag -v
```

Expected: FAIL — GetStringSliceFlag undefined

- [ ] **Step 3: Create pkg/quicli/slice_flag.go**

```go
package quicli

import "strings"

// stringSliceValue implements flag.Value for []string flags.
// It supports repeated flags (-f a -f b) and comma-separated values (-f a,b).
type stringSliceValue struct {
	val *[]string
}

func (s *stringSliceValue) String() string {
	if s.val == nil {
		return ""
	}
	return strings.Join(*s.val, ",")
}

func (s *stringSliceValue) Set(v string) error {
	*s.val = append(*s.val, strings.Split(v, ",")...)
	return nil
}
```

- [ ] **Step 4: Add createStringSliceFlag to flag_helpers.go**

Add at the end of `pkg/quicli/flag_helpers.go`:

```go
func createStringSliceFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet) {
	name := f.Name
	shortName := f.ShortName
	if shortName == "" {
		shortName = name[0:1]
	}
	val := []string{}
	sv := &stringSliceValue{val: &val}
	fs.Var(sv, name, f.Description)
	if !stringSlice.Contains(*shorts, shortName) && !f.NoShortName {
		fs.Var(sv, shortName, f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName))
		cfg.Flags[shortName] = sv
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, ""))
	}
	cfg.Flags[name] = sv
}
```

- [ ] **Step 5: Add GetStringSliceFlag to quicli.go**

Add after `GetFloatFlag` in `pkg/quicli/quicli.go`:

```go
// GetStringSliceFlag returns the []string value of a string slice flag.
func (c Config) GetStringSliceFlag(name string) []string {
	elem := c.Flags[name]
	if elem == nil {
		fmt.Println(QUICLI_ERROR_PREFIX, "failed to retrieve value for flag:", name)
		os.Exit(92)
	}
	sv := reflect.ValueOf(elem).Interface().(*stringSliceValue)
	return *sv.val
}
```

- [ ] **Step 6: Run tests**

```bash
go test ./pkg/quicli/... -v
```

Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add pkg/quicli/slice_flag.go pkg/quicli/flag_helpers.go pkg/quicli/quicli.go pkg/quicli/quicli_test.go
git commit -m "feat: add []string slice flag support with GetStringSliceFlag"
```

---

### Task 9: Misspelled subcommand detection

**Files:**
- Create: `pkg/quicli/levenshtein.go`
- Create: `pkg/quicli/levenshtein_test.go`
- Modify: `pkg/quicli/subcommand.go`

- [ ] **Step 1: Write the test for levenshtein**

Create `pkg/quicli/levenshtein_test.go`:

```go
package quicli

import "testing"

func TestLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"color", "color", 0},
		{"colo", "color", 1},
		{"cokor", "color", 2},
		{"foo", "bar", 3},
		{"", "abc", 3},
		{"abc", "", 3},
	}
	for _, tc := range cases {
		if got := levenshtein(tc.a, tc.b); got != tc.want {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestFindClosestSubcommand(t *testing.T) {
	subs := Subcommands{
		{Name: "color", Aliases: Aliases("co")},
		{Name: "toto"},
	}
	if got := findClosestSubcommand(subs, "colo"); got != "color" {
		t.Errorf("findClosestSubcommand: got %q, want color", got)
	}
	if got := findClosestSubcommand(subs, "xyz"); got != "" {
		t.Errorf("findClosestSubcommand: got %q, want empty", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test -run "TestLevenshtein|TestFindClosest" -v
```

Expected: FAIL — levenshtein undefined

- [ ] **Step 3: Create pkg/quicli/levenshtein.go**

```go
package quicli

// levenshtein returns the edit distance between strings a and b.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	dp := make([][]int, la+1)
	for i := range dp {
		dp[i] = make([]int, lb+1)
		dp[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		dp[0][j] = j
	}
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1]
			} else {
				dp[i][j] = 1 + min(dp[i-1][j], dp[i][j-1], dp[i-1][j-1])
			}
		}
	}
	return dp[la][lb]
}

// findClosestSubcommand returns the closest subcommand name/alias within
// edit distance 2, or "" if nothing is close enough.
func findClosestSubcommand(subcommands Subcommands, name string) string {
	best := ""
	bestDist := 3 // exclusive threshold
	for _, sub := range subcommands {
		if d := levenshtein(name, sub.Name); d < bestDist {
			bestDist = d
			best = sub.Name
		}
		if sub.Aliases != nil {
			for _, alias := range sub.Aliases.ToSlice() {
				if d := levenshtein(name, alias); d < bestDist {
					bestDist = d
					best = alias
				}
			}
		}
	}
	return best
}
```

- [ ] **Step 4: Wire detection into RunWithSubcommand()**

In `pkg/quicli/subcommand.go`, in `RunWithSubcommand()`, find the `isRootCommand` check that currently has `//TODO: check if subcommand is misspelled` (around line 64). Replace it with:

```go
} else {
	sub := getSubcommandByName(c.Subcommands, os.Args[1])
	if sub.Name == "" {
		suggestion := findClosestSubcommand(c.Subcommands, os.Args[1])
		if suggestion != "" {
			fmt.Fprintf(os.Stderr, QUICLI_ERROR_PREFIX+"unknown command %q — did you mean %q?\n", os.Args[1], suggestion)
		} else {
			fmt.Fprintf(os.Stderr, QUICLI_ERROR_PREFIX+"unknown command %q\n", os.Args[1])
		}
		os.Exit(2)
	}
	fmt.Fprintf(wUsage, c.Description+"\n\nUsage: "+c.Usage+"\n"+"Command "+color.Cyan(sub.Name)+": "+sub.Description+"\n\n")
}
```

- [ ] **Step 5: Run all tests**

```bash
go test ./pkg/quicli/... -v
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add pkg/quicli/levenshtein.go pkg/quicli/levenshtein_test.go pkg/quicli/subcommand.go
git commit -m "feat: misspelled subcommand detection with did-you-mean suggestion"
```

---

### Task 10: Update README + create PR

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Update README**

Keep the existing spirit (terse, practical, code-first). Add the following sections:

After the existing flags example, add:

```markdown
### Custom short name
```go
{Name: "output", Default: "text", Description: "output format", ShortName: "o"}
// registers --output and -o (instead of auto-derived -o from first letter)
```

### Slice flags
```go
{Name: "file", Default: []string{}, Description: "input files"}
// use: -file a -file b  or  -file a,b
// get: cfg.GetStringSliceFlag("file")
```

### Per-subcommand flags
```go
Subcommands: quicli.Subcommands{
  {Name: "build", Description: "build the project", Function: Build,
   Flags: quicli.Flags{{Name: "target", Default: "linux", Description: "build target"}}},
}
// --target is only registered when "build" subcommand is active
```

For flags shared across multiple subcommands use `SharedSubcommand` (renamed from `ForSubcommand`):
```go
{Name: "verbose", Description: "verbose output", SharedSubcommand: quicli.SubcommandSet{"build", "test"}}
```
```

- [ ] **Step 2: Final check**

```bash
go test ./pkg/quicli/... && go build ./examples/...
```

Expected: PASS and builds cleanly.

- [ ] **Step 3: Create PR**

```bash
git add README.md
git commit -m "docs: update README for PR1 features"
gh pr create --title "PR1: refactor internals + add missing features" --body "$(cat <<'EOF'
## Summary
- Unified createXxxFlag helpers into flag_helpers.go (eliminates 4 duplicated function pairs)
- Parse() now uses a local FlagSet — no global flag state leakage between calls
- AllAliases moved from package-level to local variable in RunWithSubcommand()
- New: GetFloatFlag, ShortName on Flag, []string slice flag + GetStringSliceFlag
- New: Flags []Flag on Subcommand for subcommand-exclusive flags
- Renamed: ForSubcommand → SharedSubcommand
- New: misspelled subcommand detection with "did you mean?" suggestion

## Test plan
- [ ] `go test ./pkg/quicli/... -v` passes
- [ ] `go build ./examples/...` succeeds
- [ ] Existing example binaries behave identically

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```
