## 🏃⌨️ quicli
### Zero-boilerplate CLI in Go

**This struct:**

```golang
type Opts struct {
    Count  int      `cli:"how many times"    default:"1"`
    Say    string   `cli:"what to say"       default:"hello"`
    World  bool     `cli:"announce it"`
    Format string   `cli:"output format"     choices:"text,json,yaml" default:"text"`
    Tags   []string `cli:"filter by tags"`
}
```

**Is a complete CLI.** Help text, short flags, env vars, shell completion, input validation, JSON Schema (for AI tool integration). All generated.

```
$ say-hello --help

Say Hello to the world

Usage: say-hello [flags]

--count   -c   how many times. (default: 1) [env: SAY_HELLO_COUNT]
--say     -s   what to say. (default: "hello") [env: SAY_HELLO_SAY]
--world   -w   announce it. (default: false) [env: SAY_HELLO_WORLD]
--format  -f   output format. (choices: text, json, yaml) (default: "text") [env: SAY_HELLO_FORMAT]
--tags    -t   filter by tags. (default: []) [env: SAY_HELLO_TAGS]
```

```bash
$ say-hello --count 3 --say "bonjour"
$ say-hello -w
$ SAY_HELLO_COUNT=5 say-hello          # env var, same as --count 5
$ say-hello --completion zsh           # shell completion
$ say-hello --json-schema              # JSON Schema for AI tools
```

No init functions. No command registration. No flag pointers.
Inspired by [nim's cligen](https://github.com/c-blake/cligen).

---

### Getting started

Tag a struct, pass a function:

```golang
func main() {
    quicli.RunFunc("say-hello [flags]", "Say Hello to the world", func(o Opts) {
        for i := 0; i < o.Count; i++ {
            msg := o.Say
            if o.World { msg = "🌍 " + msg }
            fmt.Println(msg)
        }
    })
}
```

That's it. Flags are parsed, validated, and passed to your function as a typed struct.

Supported types: `int`, `string`, `bool`, `float64`, `[]string`, `time.Duration`, or any type implementing `flag.Value`.
Tags: `cli:"desc"` · `default:"val"` · `short:"x"` · `env:"VAR"` · `required:"true"` · `choices:"a,b,c"`

<details>
<summary>With subcommands</summary>

Give each subcommand its own struct. `NewSubcommand` infers the flags from it:

```golang
type ColorOpts struct {
    Foreground bool `cli:"use foreground color"`
}

type WhisperOpts struct {
    Say   string `cli:"what to whisper" default:"psst"`
    Times int    `cli:"how many times"  default:"1"`
}

func main() {
    colorSub := quicli.NewSubcommand("color", "print in red", func(o ColorOpts) {
        fmt.Println("foreground:", o.Foreground)
    })
    colorSub.Aliases = quicli.Aliases("co")   // optional

    quicli.Cli{
        Usage:       "say-hello [command] [flags]",
        Description: "Say Hello to the world",
        Function:    func(cfg quicli.Config) { fmt.Println("hello") },
        Subcommands: quicli.Subcommands{
            colorSub,
            quicli.NewSubcommand("whisper", "say quietly", func(o WhisperOpts) {
                for i := 0; i < o.Times; i++ { fmt.Println(o.Say) }
            }),
        },
    }.RunWithSubcommand()
}
```

```bash
$ say-hello color --foreground
$ say-hello co --foreground      # alias works
$ say-hello w --say "shhh"       # unambiguous prefix works too
$ say-hello whisper --say "shhh" --times 2
```

</details>

---

### The one-liner way

Everything in one expression:

```golang
quicli.Run(quicli.Cli{
    Usage:       "say-hello [flags]",
    Description: "Say Hello to the world",
    Flags: quicli.Flags{
        {Name: "count", Default: 1,       Description: "how many times"},
        {Name: "say",   Default: "hello", Description: "what to say"},
        {Name: "world",                   Description: "announce it"},
    },
    Function: func(cfg quicli.Config) {
        count := cfg.GetIntFlag("count")
        say   := cfg.GetStringFlag("say")
        world := cfg.GetBoolFlag("world")
        for i := 0; i < count; i++ {
            if world { fmt.Print("🌍 ") }
            fmt.Println(say)
        }
    },
})
```

<details>
<summary>With subcommands</summary>

```golang
quicli.Cli{
    Usage:       "mytool [command] [flags]",
    Description: "A tool that does things",
    Flags: quicli.Flags{
        {Name: "verbose", Description: "verbose output"},
        {Name: "output",  Default: "text", Description: "output format",
            SharedSubcommand: quicli.SubcommandSet{"get", "list"}},
    },
    Function: Root,
    Subcommands: quicli.Subcommands{
        {Name: "get",    Aliases: quicli.Aliases("g"),  Description: "get a resource",  Function: Get,
            Flags: quicli.Flags{{Name: "id", Default: "", Description: "resource id"}}},
        {Name: "list",   Aliases: quicli.Aliases("ls"), Description: "list resources",  Function: List},
        {Name: "delete",                                 Description: "delete a resource", Function: Delete},
    },
}.RunWithSubcommand()
```

```
$ mytool --help

A tool that does things

Usage: mytool [command] [flags]
Available commands: get, g, list, ls, delete

--verbose  -v   verbose output. (default: false) [env: MYTOOL_VERBOSE]

Use "mytool --help" for more information about the command.

$ mytool get --help

Command get: get a resource

--id      -i   resource id. (default: "") [env: MYTOOL_ID]
--output  -o   output format. (default: "text") [env: MYTOOL_OUTPUT]
--verbose -v   verbose output. (default: false) [env: MYTOOL_VERBOSE]
```

```bash
$ mytool get --id abc123
$ mytool g --id abc123         # alias works
$ mytool ge --id abc123        # unambiguous prefix works too
$ mytool lst                   # quicli error: unknown subcommand 'lst', did you mean 'list'?
$ MYTOOL_VERBOSE=true mytool list
```

</details>

---

### Validation & constraints

```golang
type DeployOpts struct {
    Target  string        `cli:"deploy target"    required:"true"`
    Env     string        `cli:"environment"      choices:"dev,staging,prod" default:"dev"`
    Timeout time.Duration `cli:"deploy timeout"   default:"5m"`
}
```

**Required flags** `required:"true"` must be explicitly provided (via CLI or env var). Help shows `(required)` instead of a default.

**Choices** `choices:"a,b,c"` restricts the flag to a set of valid values. Invalid input exits with a clear error.

**Custom types** `time.Duration` works out of the box (`"30s"`, `"5m"`, `"1h"`). Any type implementing `flag.Value` is supported too:

```golang
type LogLevel struct{ v string }
func (l *LogLevel) String() string    { return l.v }
func (l *LogLevel) Set(s string) error { l.v = s; return nil }

type Opts struct {
    Level LogLevel `cli:"log level" default:"info"`
}
```

---

### Batteries included

Every quicli CLI gets these for free, no configuration needed:

**Env var fallback** `PROGNAME_FLAGNAME` is checked before the default. Shown in help.
```bash
SAY_HELLO_COUNT=10 ./say-hello    # same as --count 10
```
Override per flag: `EnvVar: "MY_CUSTOM_VAR"` · Opt out: `EnvVar: "-"`

**Short flags** first letter auto-derived (`--count` -> `-c`). Override with `ShortName: "n"`.

**Shell completion** one flag, three shells:
```bash
./say-hello --completion bash >> ~/.bash_completion
./say-hello --completion zsh  >  ~/.zsh/completions/_say-hello
./say-hello --completion fish >  ~/.config/fish/completions/say-hello.fish
```

**Prefix matching** type just enough to be unambiguous:
```bash
$ mytool g --id abc123           # matches "get" (unique prefix)
$ mytool d                       # matches "delete" (unique prefix)
```

**Typo detection** suggests the closest subcommand on misspelling:
```
$ mytool delet
quicli error: unknown subcommand 'delet', did you mean 'delete'?
```

**JSON Schema** expose your CLI's contract for AI tools and code generation:
```bash
$ mytool --json-schema
```
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "description": "A tool that does things",
  "properties": {
    "verbose": {
      "type": "boolean",
      "description": "verbose output",
      "x-quicli-env-var": "MYTOOL_VERBOSE"
    }
  },
  "x-quicli-subcommands": {
    "get": { "..." : "..." },
    "list": { "..." : "..." }
  }
}
```
Programmatic access: `cli.JSONSchema()`, `cli.SubcommandSchemas()`, `cli.JSONSchemaString()`.

---

### AI agent integration

quicli CLIs are AI-ready out of the box. Two approaches depending on your setup:

#### With MCP (recommended for Claude, Cursor, etc.)

Install the bridge:
```bash
go install github.com/ariary/quicli/cmd/quicli-mcp@latest
```

Add to your MCP client config (e.g. `~/.claude/claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "my-tool": {
      "command": "quicli-mcp",
      "args": ["/path/to/my-tool"]
    }
  }
}
```

That's it. Your CLI's flags become typed tool parameters. Subcommands become separate tools. Env-only flags (`EnvOnly: true`) are passed as environment variables, not CLI args.

#### Without MCP (direct API integration)

Any quicli CLI exposes its contract via `--json-schema`. Feed it to any LLM with tool/function calling:

```bash
# 1. Get the schema
SCHEMA=$(./my-tool --json-schema)

# 2. Use it as a tool definition in your API call
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -d "{
    \"model\": \"claude-sonnet-4-20250514\",
    \"tools\": [{
      \"name\": \"my-tool\",
      \"description\": \"$(echo $SCHEMA | jq -r .description)\",
      \"input_schema\": $SCHEMA
    }],
    \"messages\": [{\"role\": \"user\", \"content\": \"...\"}]
  }"

# 3. When the model calls the tool, run:
./my-tool --count 5 --name hello
```

The schema works with Claude API tool_use, OpenAI function calling, or any agent framework that accepts JSON Schema.

#### Env-only flags for secrets

Keep API keys and tokens out of shell history:

```golang
type Opts struct {
    Target string `cli:"target to scan" required:"true"`
    APIKey string `cli:"API key"        env:"only"`
}
```

`--help` won't show `APIKey`. `--json-schema` marks it as `"x-quicli-input": "env-only"`. The MCP bridge passes it as an environment variable automatically.

#### Debug your flags

See where each flag's value comes from:

```bash
$ MYTOOL_NAME=world ./mytool --count 5 --debug-options
FLAG      VALUE    SOURCE
--count   5        cli
--name    world    env (MYTOOL_NAME)
--format  json     default
--secret  ***      env (MYTOOL_SECRET)
```

---

Get more [examples](examples/)

> quicli is a thin wrapper around Go's `flag` package. Use it to write CLIs fast, not to build complex command hierarchies.
