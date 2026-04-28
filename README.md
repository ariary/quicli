## 🏃⌨️ quicli
### Build CLI in Go without the ceremony

No init functions. No command registration. No flag pointers.
Define your CLI in one expression *(or just tag a struct)*

Inspired by [nim's cligen](https://github.com/c-blake/cligen).

---

### The zero-boilerplate way

Tag a struct, pass a function. Your flags are already in `Opts` Struct:

```golang
type Opts struct {
    Count  int           `cli:"how many times"    default:"1"`
    Say    string        `cli:"what to say"       default:"hello"`
    World  bool          `cli:"announce it"`
    Format string        `cli:"output format"     choices:"text,json,yaml" default:"text"`
    Tags   []string      `cli:"filter by tags"`
}

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

```
$ say-hello --help

Say Hello to the world

Usage: say-hello [flags]

--count   -c   how many times. (default: 1) [env: SAY_HELLO_COUNT]
--say     -s   what to say. (default: "hello") [env: SAY_HELLO_SAY]
--world   -w   announce it. (default: false) [env: SAY_HELLO_WORLD]
--format  -f   output format. (choices: text, json, yaml) (default: "text") [env: SAY_HELLO_FORMAT]
--tags    -t   filter by tags. (default: []) [env: SAY_HELLO_TAGS]

Use "say-hello --help" for more information about the command.
```

```bash
$ say-hello --count 3 --say "bonjour"
$ say-hello -w
$ SAY_HELLO_COUNT=5 say-hello          # env var, same as --count 5
$ say-hello --completion zsh           # print zsh completion script
$ say-hello --json-schema              # print JSON Schema for AI tools
```

Supported field types: `int`, `string`, `bool`, `float64`, `[]string`, `time.Duration`, or any type implementing `flag.Value`.
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

### From CLI to MCP server

quicli generates JSON Schema from your flags automatically. This is the foundation for turning any quicli CLI into an [MCP](https://modelcontextprotocol.io/) tool server so AI agents can call your CLI with typed inputs.

**Step 1** Build your CLI normally (you already did this):

```golang
type ScanOpts struct {
    Target string `cli:"target to scan" required:"true"`
    Port   int    `cli:"port number"    default:"443"`
}

cli := quicli.Cli{
    Usage:       "scanner [command] [flags]",
    Description: "Network scanner",
    Function:    func(cfg quicli.Config) {},
    Subcommands: quicli.Subcommands{
        quicli.NewSubcommand("scan", "scan a target", func(o ScanOpts) {
            fmt.Printf("scanning %s:%d\n", o.Target, o.Port)
        }),
    },
}
```

**Step 2** Get the schemas for your subcommands:

```golang
schemas := cli.SubcommandSchemas()
// schemas["scan"] → {
//   "type": "object",
//   "properties": {
//     "target": {"type":"string", "description":"target to scan"},
//     "port":   {"type":"integer", "description":"port number", "default":443}
//   },
//   "required": ["target"]
// }
```

Each schema is a valid JSON Schema, exactly what MCP's `tools/list` needs for `inputSchema`, and what OpenAI/Claude function calling needs for `parameters`.

**Step 3** Wire it into an MCP server:

The MCP tool-server protocol is JSON-RPC over stdio. The mapping is:

| MCP method | quicli API |
|---|---|
| `tools/list` | `cli.SubcommandSchemas()` -> tool name + inputSchema |
| `tools/call` | Parse input JSON -> populate flags -> call `Function` |

The schema generation is the hard part, you already have it. The protocol framing is ~200 lines of Go on top.

```bash
$ scanner --json-schema   # verify your schemas look right
```

---

Get more [examples](examples/)

> quicli is a thin wrapper around Go's `flag` package. Use it to write CLIs fast, not to build complex command hierarchies.
