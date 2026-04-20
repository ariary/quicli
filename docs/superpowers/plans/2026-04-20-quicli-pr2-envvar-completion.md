# PR2: Env Var Mapping + Shell Completion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Prerequisite:** PR1 must be merged first (unified flag helpers, local FlagSet in Parse()).

**Goal:** Auto-map every flag to an env var (`PROGNAME_FLAGNAME`), support explicit `EnvVar` override on `Flag`, and auto-inject a `--completion <shell>` flag that generates bash/zsh/fish completion scripts.

**Architecture:** `envvar.go` handles name derivation and post-parse env var injection. `completion.go` contains three string generators (bash/zsh/fish) and the injection logic. Both Parse() and RunWithSubcommand() call the env var injection step after `fs.Parse()`.

**Tech Stack:** Go 1.18, stdlib `flag`/`os`/`strings`/`path/filepath`, no new dependencies.

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `pkg/quicli/envvar.go` | Create | envVarName(), applyEnvVars() |
| `pkg/quicli/completion.go` | Create | generateCompletion(), bash/zsh/fish generators |
| `pkg/quicli/flag_helpers.go` | Modify | getFlagLine gains envVar param, shows env in help |
| `pkg/quicli/quicli.go` | Modify | call applyEnvVars + inject --completion after parse |
| `pkg/quicli/subcommand.go` | Modify | same wiring as quicli.go |
| `pkg/quicli/envvar_test.go` | Create | tests for name derivation and lookup |
| `pkg/quicli/completion_test.go` | Create | tests for script generators |
| `README.md` | Modify | env var section + completion section |

---

### Task 1: Env var name derivation + applyEnvVars

**Files:**
- Create: `pkg/quicli/envvar.go`
- Create: `pkg/quicli/envvar_test.go`

- [ ] **Step 1: Write failing tests**

Create `pkg/quicli/envvar_test.go`:

```go
package quicli

import (
	"os"
	"testing"
)

func TestEnvVarName(t *testing.T) {
	cases := []struct {
		progName string
		flagName string
		want     string
	}{
		{"say-hello", "count", "SAY_HELLO_COUNT"},
		{"./mycli", "output-format", "MYCLI_OUTPUT_FORMAT"},
		{"prog", "file", "PROG_FILE"},
	}
	for _, tc := range cases {
		if got := envVarName(tc.progName, tc.flagName); got != tc.want {
			t.Errorf("envVarName(%q, %q) = %q, want %q", tc.progName, tc.flagName, got, tc.want)
		}
	}
}

func TestApplyEnvVarInt(t *testing.T) {
	defer setArgs([]string{"prog"})()
	os.Setenv("PROG_COUNT", "7")
	defer os.Unsetenv("PROG_COUNT")

	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "count", Default: 0, Description: "count"}},
	}
	cfg := cli.Parse()
	if got := cfg.GetIntFlag("count"); got != 7 {
		t.Errorf("env var int: got %d, want 7", got)
	}
}

func TestApplyEnvVarExplicit(t *testing.T) {
	defer setArgs([]string{"prog"})()
	os.Setenv("MY_CUSTOM_COUNT", "99")
	defer os.Unsetenv("MY_CUSTOM_COUNT")

	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "count", Default: 0, Description: "count", EnvVar: "MY_CUSTOM_COUNT"}},
	}
	cfg := cli.Parse()
	if got := cfg.GetIntFlag("count"); got != 99 {
		t.Errorf("explicit EnvVar: got %d, want 99", got)
	}
}

func TestApplyEnvVarOptOut(t *testing.T) {
	defer setArgs([]string{"prog"})()
	os.Setenv("PROG_SECRET", "ignored")
	defer os.Unsetenv("PROG_SECRET")

	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test",
		Flags:       Flags{{Name: "secret", Default: "default", Description: "secret", EnvVar: "-"}},
	}
	cfg := cli.Parse()
	if got := cfg.GetStringFlag("secret"); got != "default" {
		t.Errorf("opt-out: got %q, want default", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test -run "TestEnvVar|TestApplyEnv" -v
```

Expected: FAIL — envVarName undefined

- [ ] **Step 3: Create pkg/quicli/envvar.go**

```go
package quicli

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
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

// applyEnvVars applies env var values to flags that were not set on the command line.
// Priority: CLI flag > env var > default.
// A flag is considered "set on CLI" if its value differs from the default,
// or if it appears in the explicitly-set flags tracked by the FlagSet.
func applyEnvVars(c *Cli, config Config, fs interface{ Lookup(string) interface{ Value() interface{ String() string } } }) {
	// We use a simpler approach: after fs.Parse(), call os.LookupEnv for each flag
	// and use fs.Set to override only if the flag was not explicitly passed.
	// We detect "not explicitly passed" by checking if the flag still holds its default value.
	// This works because the FlagSet tracks which flags were visited.
}
```

Wait — the cleanest approach uses `fs.Visit` (iterates only explicitly-set flags) vs `fs.VisitAll`. Use `fs.Visit` to build a set of CLI-provided flags, then apply env vars for the rest.

Replace the `applyEnvVars` stub above with:

```go
package quicli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// envVarName derives the auto env var name from program name and flag name.
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
// and updates config accordingly. Called after fs.Parse().
func applyEnvVars(flags []Flag, config Config, fs *flag.FlagSet) {
	// Collect flags that were explicitly set on the command line.
	cliProvided := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		cliProvided[f.Name] = true
	})

	for _, f := range flags {
		if cliProvided[f.Name] {
			continue
		}
		// Determine env var name to check.
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
		// Apply the value via fs.Set so the flag package parses it correctly.
		if err := fs.Set(f.Name, val); err != nil {
			fmt.Fprintf(os.Stderr, QUICLI_ERROR_PREFIX+"env var %s=%q invalid for flag --%s: %v\n", envKey, val, f.Name, err)
			os.Exit(1)
		}
	}
}
```

- [ ] **Step 4: Run tests**

```bash
go test -run "TestEnvVar|TestApplyEnv" -v
```

Expected: TestEnvVarName PASS, TestApplyEnv* still FAIL (not wired yet)

- [ ] **Step 5: Commit**

```bash
git add pkg/quicli/envvar.go pkg/quicli/envvar_test.go
git commit -m "feat: add envvar.go with envVarName() and applyEnvVars()"
```

---

### Task 2: Wire applyEnvVars into Parse() and RunWithSubcommand()

**Files:**
- Modify: `pkg/quicli/quicli.go`
- Modify: `pkg/quicli/subcommand.go`

- [ ] **Step 1: Wire into Parse() in quicli.go**

In `Parse()`, after the line `config.Args = fs.Args()`, add:

```go
applyEnvVars(c.Flags, config, fs)
```

- [ ] **Step 2: Wire into RunWithSubcommand() in subcommand.go**

In `RunWithSubcommand()`, after `config.Args = fs.Args()`, add:

```go
// Apply env vars for top-level flags and (if subcommand active) its exclusive flags.
allFlags := c.Flags
if !isRootCommand(c.Subcommands) {
	sub := getSubcommandByName(c.Subcommands, os.Args[1])
	allFlags = append(allFlags, sub.Flags...)
}
applyEnvVars(allFlags, config, fs)
```

- [ ] **Step 3: Run all tests**

```bash
go test ./pkg/quicli/... -v
```

Expected: TestApplyEnv* PASS, all others still PASS

- [ ] **Step 4: Commit**

```bash
git add pkg/quicli/quicli.go pkg/quicli/subcommand.go
git commit -m "feat: wire applyEnvVars into Parse() and RunWithSubcommand()"
```

---

### Task 3: Show env var in help output

**Files:**
- Modify: `pkg/quicli/flag_helpers.go`

- [ ] **Step 1: Update getFlagLine signature to accept envVar**

In `pkg/quicli/flag_helpers.go`, update `getFlagLine`:

```go
func getFlagLine(description string, defaultValue interface{}, long string, short string, envVar string) string {
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
	if envVar != "" && envVar != "-" {
		defaultStr += " [env: " + envVar + "]"
	}
	defaultStr += "\n"

	if short == "" {
		return "--" + long + "\t\t\t" + description + defaultStr
	}
	return "--" + long + "\t-" + short + "\t\t" + description + defaultStr
}
```

- [ ] **Step 2: Update all callers of getFlagLine in flag_helpers.go**

Each `createXxxFlag` function calls `getFlagLine(...)`. Add the env var as the last argument. For the env var to display correctly, we need to derive it at call time:

In each `createXxxFlag` function, derive the display env var before calling `getFlagLine`:

```go
displayEnv := f.EnvVar
if displayEnv == "" && f.EnvVar != "-" {
    displayEnv = envVarName(os.Args[0], f.Name)
}
if f.EnvVar == "-" {
    displayEnv = ""
}
```

Then pass `displayEnv` as the last argument to `getFlagLine`.

Update all four `createXxxFlag` calls (two per function: the short-name branch and the no-short-name branch).

Also update `createStringSliceFlag` the same way.

- [ ] **Step 3: Run all tests**

```bash
go test ./pkg/quicli/... -v
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add pkg/quicli/flag_helpers.go
git commit -m "feat: show env var name in flag help output"
```

---

### Task 4: Shell completion — bash + zsh + fish generators

**Files:**
- Create: `pkg/quicli/completion.go`
- Create: `pkg/quicli/completion_test.go`

- [ ] **Step 1: Write failing tests**

Create `pkg/quicli/completion_test.go`:

```go
package quicli

import (
	"strings"
	"testing"
)

func testCli() Cli {
	return Cli{
		Usage:       "prog [command] [flags]",
		Description: "test cli",
		Flags: Flags{
			{Name: "verbose", Default: false, Description: "verbose output"},
			{Name: "output", Default: "text", Description: "output format"},
		},
		Subcommands: Subcommands{
			{Name: "build", Aliases: Aliases("b"), Description: "build the project", Function: func(Config) {}},
			{Name: "test", Description: "run tests", Function: func(Config) {}},
		},
		Function: func(Config) {},
	}
}

func TestGenerateBashCompletion(t *testing.T) {
	cli := testCli()
	script, err := generateCompletion(&cli, "bash")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, "build") {
		t.Error("bash completion missing subcommand 'build'")
	}
	if !strings.Contains(script, "--verbose") {
		t.Error("bash completion missing flag '--verbose'")
	}
	if !strings.Contains(script, "complete -F") {
		t.Error("bash completion missing complete builtin")
	}
}

func TestGenerateZshCompletion(t *testing.T) {
	cli := testCli()
	script, err := generateCompletion(&cli, "zsh")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, "#compdef") {
		t.Error("zsh completion missing #compdef header")
	}
	if !strings.Contains(script, "build") {
		t.Error("zsh completion missing subcommand 'build'")
	}
}

func TestGenerateFishCompletion(t *testing.T) {
	cli := testCli()
	script, err := generateCompletion(&cli, "fish")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, "complete -c") {
		t.Error("fish completion missing 'complete -c'")
	}
	if !strings.Contains(script, "build") {
		t.Error("fish completion missing subcommand 'build'")
	}
}

func TestGenerateCompletionUnknownShell(t *testing.T) {
	cli := testCli()
	_, err := generateCompletion(&cli, "powershell")
	if err == nil {
		t.Error("expected error for unknown shell")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test -run "TestGenerate" -v
```

Expected: FAIL — generateCompletion undefined

- [ ] **Step 3: Create pkg/quicli/completion.go**

```go
package quicli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateCompletion returns a shell completion script for the given Cli and shell.
func generateCompletion(c *Cli, shell string) (string, error) {
	switch shell {
	case "bash":
		return generateBashCompletion(c), nil
	case "zsh":
		return generateZshCompletion(c), nil
	case "fish":
		return generateFishCompletion(c), nil
	default:
		return "", fmt.Errorf("unsupported shell %q — use bash, zsh, or fish", shell)
	}
}

func progName() string {
	return filepath.Base(os.Args[0])
}

// allSubcommandNames returns all subcommand names and aliases as a flat slice.
func allSubcommandNames(subs Subcommands) []string {
	var names []string
	for _, s := range subs {
		names = append(names, s.Name)
		if s.Aliases != nil {
			names = append(names, s.Aliases.ToSlice()...)
		}
	}
	return names
}

// allFlagNames returns --long and -short flag names.
func allFlagNames(flags []Flag) []string {
	var names []string
	seen := map[string]bool{}
	for _, f := range flags {
		names = append(names, "--"+f.Name)
		short := f.ShortName
		if short == "" {
			short = f.Name[0:1]
		}
		if !f.NoShortName && !seen[short] {
			names = append(names, "-"+short)
			seen[short] = true
		}
	}
	return names
}

func generateBashCompletion(c *Cli) string {
	prog := progName()
	fnName := "_" + strings.ReplaceAll(prog, "-", "_") + "_completion"
	cmds := strings.Join(allSubcommandNames(c.Subcommands), " ")
	flags := strings.Join(allFlagNames(c.Flags), " ")

	var b strings.Builder
	fmt.Fprintf(&b, "%s() {\n", fnName)
	fmt.Fprintf(&b, "    local cur=\"${COMP_WORDS[COMP_CWORD]}\"\n")
	if len(c.Subcommands) > 0 {
		fmt.Fprintf(&b, "    local commands=\"%s\"\n", cmds)
	}
	fmt.Fprintf(&b, "    local flags=\"%s\"\n", flags)
	fmt.Fprintf(&b, "\n")
	if len(c.Subcommands) > 0 {
		fmt.Fprintf(&b, "    if [[ $COMP_CWORD -eq 1 ]]; then\n")
		fmt.Fprintf(&b, "        COMPREPLY=( $(compgen -W \"$commands $flags\" -- \"$cur\") )\n")
		fmt.Fprintf(&b, "    else\n")
		fmt.Fprintf(&b, "        COMPREPLY=( $(compgen -W \"$flags\" -- \"$cur\") )\n")
		fmt.Fprintf(&b, "    fi\n")
	} else {
		fmt.Fprintf(&b, "    COMPREPLY=( $(compgen -W \"$flags\" -- \"$cur\") )\n")
	}
	fmt.Fprintf(&b, "}\n")
	fmt.Fprintf(&b, "complete -F %s %s\n", fnName, prog)
	return b.String()
}

func generateZshCompletion(c *Cli) string {
	prog := progName()
	fnName := "_" + strings.ReplaceAll(prog, "-", "_")

	var b strings.Builder
	fmt.Fprintf(&b, "#compdef %s\n", prog)
	fmt.Fprintf(&b, "%s() {\n", fnName)

	if len(c.Subcommands) > 0 {
		fmt.Fprintf(&b, "    local -a commands\n")
		fmt.Fprintf(&b, "    commands=(\n")
		for _, s := range c.Subcommands {
			fmt.Fprintf(&b, "        '%s:%s'\n", s.Name, s.Description)
		}
		fmt.Fprintf(&b, "    )\n")
		fmt.Fprintf(&b, "    local -a flags\n")
		fmt.Fprintf(&b, "    flags=(\n")
		for _, f := range c.Flags {
			fmt.Fprintf(&b, "        '--%s[%s]'\n", f.Name, f.Description)
		}
		fmt.Fprintf(&b, "    )\n")
		fmt.Fprintf(&b, "    if (( CURRENT == 2 )); then\n")
		fmt.Fprintf(&b, "        _describe 'commands' commands\n")
		fmt.Fprintf(&b, "    else\n")
		fmt.Fprintf(&b, "        _arguments $flags\n")
		fmt.Fprintf(&b, "    fi\n")
	} else {
		fmt.Fprintf(&b, "    local -a flags\n")
		fmt.Fprintf(&b, "    flags=(\n")
		for _, f := range c.Flags {
			fmt.Fprintf(&b, "        '--%s[%s]'\n", f.Name, f.Description)
		}
		fmt.Fprintf(&b, "    )\n")
		fmt.Fprintf(&b, "    _arguments $flags\n")
	}

	fmt.Fprintf(&b, "}\n")
	fmt.Fprintf(&b, "%s \"$@\"\n", fnName)
	return b.String()
}

func generateFishCompletion(c *Cli) string {
	prog := progName()
	var b strings.Builder

	for _, s := range c.Subcommands {
		fmt.Fprintf(&b, "complete -c %s -n \"__fish_use_subcommand\" -f -a %s -d '%s'\n",
			prog, s.Name, s.Description)
		if s.Aliases != nil {
			for _, alias := range s.Aliases.ToSlice() {
				fmt.Fprintf(&b, "complete -c %s -n \"__fish_use_subcommand\" -f -a %s -d '%s (alias)'\n",
					prog, alias, s.Description)
			}
		}
	}
	for _, f := range c.Flags {
		short := f.ShortName
		if short == "" {
			short = f.Name[0:1]
		}
		if f.NoShortName {
			fmt.Fprintf(&b, "complete -c %s -l %s -d '%s'\n", prog, f.Name, f.Description)
		} else {
			fmt.Fprintf(&b, "complete -c %s -s %s -l %s -d '%s'\n", prog, short, f.Name, f.Description)
		}
	}
	return b.String()
}
```

- [ ] **Step 4: Run tests**

```bash
go test -run "TestGenerate" -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/quicli/completion.go pkg/quicli/completion_test.go
git commit -m "feat: add bash/zsh/fish completion generators"
```

---

### Task 5: Auto-inject --completion flag into Parse() and RunWithSubcommand()

**Files:**
- Modify: `pkg/quicli/quicli.go`
- Modify: `pkg/quicli/subcommand.go`

- [ ] **Step 1: Add --completion injection to Parse() in quicli.go**

In `Parse()`, in the block where `--cheat-sheet` is added (near the bottom of the flags setup, before `wUsage.Flush()`), add:

```go
// completion injection
var completionShell string
fs.StringVar(&completionShell, "completion", "", "generate shell completion script (bash, zsh, fish)")
```

After `config.Args = fs.Args()` and before the cheat-sheet check, add:

```go
if completionShell != "" {
    script, err := generateCompletion(c, completionShell)
    if err != nil {
        fmt.Fprintln(os.Stderr, QUICLI_ERROR_PREFIX+err.Error())
        os.Exit(1)
    }
    fmt.Print(script)
    os.Exit(0)
}
```

- [ ] **Step 2: Add the same injection to RunWithSubcommand() in subcommand.go**

Apply the identical pattern: register `completionShell` on `fs`, then check and handle it after `config.Args = fs.Args()`.

- [ ] **Step 3: Run all tests**

```bash
go test ./pkg/quicli/... -v
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add pkg/quicli/quicli.go pkg/quicli/subcommand.go
git commit -m "feat: auto-inject --completion flag for shell completion scripts"
```

---

### Task 6: Update README + create PR

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Add env var section to README**

After the "Use flag values in code" section, add:

```markdown
### Env vars

Every flag automatically reads from an env var as fallback before using the default.
Auto-derived name: `PROGNAME_FLAGNAME` (uppercase, non-alphanumeric → `_`).

```bash
SAY_HELLO_COUNT=5 ./say-hello   # same as --count 5
```

Override the env var name per flag:
```go
{Name: "token", Default: "", Description: "API token", EnvVar: "MY_API_TOKEN"}
```

Opt a flag out of env var lookup:
```go
{Name: "secret", Default: "", Description: "...", EnvVar: "-"}
```
```

- [ ] **Step 2: Add completion section to README**

```markdown
### Shell completion

Every CLI built with quicli gets `--completion <shell>` for free:

```bash
./mycli --completion bash >> ~/.bash_completion
./mycli --completion zsh  > ~/.zsh/completions/_mycli
./mycli --completion fish > ~/.config/fish/completions/mycli.fish
```
```

- [ ] **Step 3: Final check**

```bash
go test ./pkg/quicli/... && go build ./examples/...
```

Expected: PASS

- [ ] **Step 4: Create PR**

```bash
git add README.md
git commit -m "docs: update README for env var and shell completion features"
gh pr create --title "PR2: env var mapping + shell completion" --body "$(cat <<'EOF'
## Summary
- Auto-map every flag to PROGNAME_FLAGNAME env var (opt-out with EnvVar: "-", override with EnvVar: "CUSTOM")
- Env var shown in help: `(default: 1) [env: SAY_HELLO_COUNT]`
- Auto-injected --completion flag generates bash/zsh/fish scripts from live Cli struct

## Test plan
- [ ] `go test ./pkg/quicli/... -v` passes
- [ ] `SAY_HELLO_COUNT=3 go run examples/sayhello.go` prints hello 3 times
- [ ] `go run examples/sayhello.go --completion bash` prints valid bash script

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```
