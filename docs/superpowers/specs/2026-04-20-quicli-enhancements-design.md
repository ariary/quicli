# quicli enhancements — design spec
_2026-04-20_

## Philosophy

quicli is a wrapper around Go's stdlib `flag` package for building CLIs in one expressive block — explicit, minimal, no heavy dependencies. Inspired by nim's cligen. All changes must preserve this contract: the user-facing API stays terse and declarative.

---

## PR 1 — Internal refactor + missing features

### Internal unification

`quicli.go` and `subcommand.go` currently duplicate four `createXxxFlag` / `createXxxFlagFs` helper pairs. Both files will share a single set of helpers:

```go
func createIntFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet)
func createStringFlag(...)
func createBoolFlag(...)
func createFloatFlag(...)
```

`Parse()` will create a local `flag.FlagSet` instead of using the global `flag` package. The package-level `AllAliases` variable moves inside `RunWithSubcommand` as a local variable, eliminating the stale-state bug when multiple `Cli` instances exist.

No public API change.

### Updated `Flag` struct

```go
type Flag struct {
    Name              string
    Description       string
    Default           interface{}      // type determines flag kind; nil → bool false
    NoShortName       bool
    ShortName         string           // NEW: overrides auto first-letter derivation
    NotForRootCommand bool
    SharedSubcommand  SubcommandSet    // RENAMED from ForSubcommand
    EnvVar            string           // field added here, populated in PR 2
}
```

### Updated `Subcommand` struct

```go
type Subcommand struct {
    Name        string
    Aliases     mapset.Set[string]
    Description string
    Function    Runner
    Flags       []Flag    // NEW: flags exclusive to this subcommand
}
```

Flags defined on a `Subcommand.Flags` are only registered and shown when that subcommand is active. `SharedSubcommand` on top-level flags covers flags shared across multiple subcommands.

### New `Config` method

```go
func (c Config) GetFloatFlag(name string) float64
```

Follows the same pattern as `GetIntFlag`, `GetStringFlag`, `GetBoolFlag`.

### `[]string` slice flag

`Default: []string{}` triggers slice handling. Internally implemented via a custom `flag.Value` — no new dependencies. Accepts repeated flags: `-f a -f b` or comma-separated `-f a,b` (both work).

New accessor:
```go
func (c Config) GetStringSliceFlag(name string) []string
```

### Misspelled subcommand detection

When `os.Args[1]` matches no subcommand name or alias, compute edit distance against all known names and aliases. If the closest match has distance ≤ 2, emit:

```
unknown command "colo" — did you mean "color"?
```

Then exit 2. Uses a minimal inline Levenshtein implementation — no new dependency.

### README

Add a brief note on `ShortName`, `SharedSubcommand` rename (with migration note from `ForSubcommand`), per-subcommand `Flags`, `GetFloatFlag`, and `[]string` flags — each shown with a one-line example in the existing style.

---

## PR 2 — Env var mapping + shell completion

### Env var auto-mapping

After flag parsing, for each flag whose parsed value equals its default, quicli checks an env var. Priority: CLI flag → env var → default.

Auto-derived name: basename of `os.Args[0]`, uppercased, non-alphanumeric replaced with `_`, followed by `_FLAGNAME`. Example: `./say-hello` + flag `count` → `SAY_HELLO_COUNT`.

Explicit override: `EnvVar: "MY_CUSTOM_VAR"` on a `Flag` skips auto-derivation for that flag. Set `EnvVar: "-"` to opt a flag out of env var lookup entirely.

Help line shows the env var:
```
--count  -c    how many times. (default: 1, env: SAY_HELLO_COUNT)
```

### Shell completion

A `--completion <shell>` flag is auto-injected into every `Cli`. When triggered it prints a completion script to stdout and exits 0. Supported: `bash`, `zsh`, `fish`.

The script is generated at runtime from the live `Cli` struct (flag names, short names, subcommand names, aliases) — always in sync, no static files.

```bash
./mycli --completion bash >> ~/.bash_completion
./mycli --completion zsh  > ~/.zsh/completions/_mycli
./mycli --completion fish > ~/.config/fish/completions/mycli.fish
```

No new dependencies — pure string generation.

### README

Add env var section: one example showing auto-derived name + `EnvVar` override. Add completion section: three one-liners for bash/zsh/fish install.

---

## PR 3 — RunFunc (struct-tag inference)

### API

```go
func RunFunc[T any](usage, description string, fn func(T))
```

Uses Go 1.18 generics (already the module minimum). `T` must be a struct — quicli reflects on its fields to derive flags.

### Struct tag format

```go
type Opts struct {
    Count int    `cli:"how many times I want to say it"`
    Say   string `cli:"say something" default:"hello"`
    World bool   `cli:"announce it to the world"`
    Name  string `cli:"your name" short:"n" env:"MY_NAME"`
}
```

- `cli:` → description (required for the field to be included)
- `default:` → default value as string, parsed to field type; omit → zero value
- `short:` → short flag name override (same as `ShortName` on `Flag`)
- `env:` → env var override (same as `EnvVar` on `Flag`)

Fields without a `cli:` tag are ignored — safe to mix CLI and non-CLI fields.

### Derivation rules

| Field type  | Flag kind     |
|-------------|---------------|
| `int`       | int           |
| `string`    | string        |
| `bool`      | bool          |
| `float64`   | float64       |
| `[]string`  | string slice  |

Flag name: field name lowercased (`Count` → `--count`).

### Usage

```go
func SayHello(o Opts) {
    for i := 0; i < o.Count; i++ {
        fmt.Println(o.Say)
    }
}

func main() {
    quicli.RunFunc("SayToTheWorld [flags]", "Say Hello...", SayHello)
}
```

`RunFuncWithSubcommand` is out of scope — subcommand interaction is a future concern.

### README

New "Zero-struct" section at the top, showing the full example above. Positioned as the simplest entry point; existing `Cli` struct approach remains documented below it as the path for subcommands and advanced config.

---

## Out of scope

- Nested subcommands (`cmd sub1 sub2`)
- `RunFuncWithSubcommand`
- Man page generation
- Config file (YAML/TOML) integration
