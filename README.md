## 🏃⌨️ quicli
### Build CLI in Go without the ceremony

No init functions. No command registration. No flag pointers.
Define your CLI in one expression — or just tag a struct.

Inspired by [nim's cligen](https://github.com/c-blake/cligen).

---

### The zero-boilerplate way

Tag a struct, pass a function. Your flags are already in `o` — no `GetXxxFlag` calls:

```golang
type Opts struct {
    Count int      `cli:"how many times"    default:"1"`
    Say   string   `cli:"what to say"       default:"hello"`
    World bool     `cli:"announce it"`
    Tags  []string `cli:"filter by tags"`
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

--count  -c   how many times. (default: 1) [env: SAY_HELLO_COUNT]
--say    -s   what to say. (default: "hello") [env: SAY_HELLO_SAY]
--world  -w   announce it. (default: false) [env: SAY_HELLO_WORLD]
--tags   -t   filter by tags. (default: []) [env: SAY_HELLO_TAGS]

Use "say-hello --help" for more information about the command.
```

```bash
$ say-hello --count 3 --say "bonjour"
$ say-hello -w
$ SAY_HELLO_COUNT=5 say-hello          # env var, same as --count 5
$ say-hello --completion zsh           # print zsh completion script
```

Supported field types: `int`, `string`, `bool`, `float64`, `[]string`.
Tags: `cli:"desc"` · `default:"val"` · `short:"x"` · `env:"VAR"` (or `"-"` to opt out).

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

### Batteries included

Every quicli CLI gets these for free — no configuration needed:

**Env var fallback** — `PROGNAME_FLAGNAME` is checked before the default. Shown in help.
```bash
SAY_HELLO_COUNT=10 ./say-hello    # same as --count 10
```
Override per flag: `EnvVar: "MY_CUSTOM_VAR"` · Opt out: `EnvVar: "-"`

**Short flags** — first letter auto-derived (`--count` → `-c`). Override with `ShortName: "n"`.

**Shell completion** — one flag, three shells:
```bash
./say-hello --completion bash >> ~/.bash_completion
./say-hello --completion zsh  >  ~/.zsh/completions/_say-hello
./say-hello --completion fish >  ~/.config/fish/completions/say-hello.fish
```

**Typo detection** — suggests the closest subcommand on misspelling:
```
$ mytool delet
quicli error: unknown subcommand 'delet', did you mean 'delete'?
```

---

Get more [examples](examples/)

> quicli is a thin wrapper around Go's `flag` package. Use it to write CLIs fast, not to build complex command hierarchies.
