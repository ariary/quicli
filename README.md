## 🏃⌨️ quicli
### Build CLI in one line
<sup>..or two</sup>

### Zero-struct — the cligen way

Define a struct, tag the fields, pass a function:

```golang
type Opts struct {
    Count int    `cli:"how many times I want to say it" default:"1"`
    Say   string `cli:"say something"                   default:"hello"`
    World bool   `cli:"announce it to the world"`
}

func main() {
    quicli.RunFunc("say-hello [flags]", "Say Hello...", func(o Opts) {
        for i := 0; i < o.Count; i++ {
            if o.World { fmt.Print("Message for the world: ") }
            fmt.Println(o.Say)
        }
    })
}
```

This gives you:
```
Say Hello...

Usage: say-hello [flags]

--count  -c   how many times I want to say it. (default: 1) [env: SAY_HELLO_COUNT]
--say    -s   say something. (default: "hello") [env: SAY_HELLO_SAY]
--world  -w   announce it to the world. (default: false) [env: SAY_HELLO_WORLD]

Use "say-hello --help" for more information about the command.
```

```bash
./say-hello --count 3 --say "hi"
./say-hello -w
SAY_HELLO_COUNT=5 ./say-hello          # env var fallback
./say-hello --completion bash          # print bash completion script
```

Supported field types: `int`, `string`, `bool`, `float64`, `[]string`.
Tags: `cli:"desc"` (required), `default:"val"`, `short:"x"`, `env:"VAR"`.
Untagged fields are ignored — safe to mix CLI and non-CLI fields in the same struct.

### Cli struct — for subcommands and full control

```golang
cli := quicli.Cli{Usage:"say-hello [flags]",Description: "Say Hello... or not. If you want to make the world aware of it you also could",Flags: quicli.Flags{{Name: "count", Default: 1, Description: "how many times I want to say it. Sometimes repetition is the key"},{Name: "say", Default: "hello", Description: "say something. If you are polite start with a greeting"},{Name: "world", Description: "announce it to the world"},},}
cfg := cli.Parse()
```

With this code you obtain the following help message:
```
Say Hello... or not. If you want to make the world aware of it you also could

Usage: say-hello [flags]

--count  -c   how many times I want to say it. Sometimes repetition is the key. (default: 1) [env: SAY_HELLO_COUNT]
--say    -s   say something. If you are polite start with a greeting. (default: "hello") [env: SAY_HELLO_SAY]
--world  -w   announce it to the world. (default: false) [env: SAY_HELLO_WORLD]

Use "say-hello --help" for more information about the command.
```

<details>
    <summary>Pretty indented version</summary>

```golang
cli := quicli.Cli{
  Usage:       "say-hello [flags]",
  Description: "Say Hello... or not. If you want to make the world aware of it you also could",
  Flags: quicli.Flags{
    {Name: "count", Default: 1, Description: "how many times I want to say it. Sometimes repetition is the key"},
    {Name: "say", Default: "hello", Description: "say something. If you are polite start with a greeting"},
    {Name: "world", Description: "announce it to the world"},
  },
}
cfg := cli.Parse()
```
</details>

<details>
    <summary>Real one-liner (Parse and run)</summary>

```golang
quicli.Run(quicli.Cli{Usage:"say-hello [flags]",Description: "Say Hello... or not. If you want to make the world aware of it you also could",Flags: quicli.Flags{{Name: "count", Default: 1, Description: "how many times I want to say it. Sometimes repetition is the key"},{Name: "say", Default: "hello", Description: "say something. If you are polite start with a greeting"},{Name: "world", Description: "announce it to the world"},},Function: SayHello,})
```
</details>

<details>
    <summary>You want a subcommand pattern?! okay</summary>

```golang
cli := quicli.Cli{
    Usage:       "say-hello [command] [flags]",
    Description: "Say Hello... or not. If you want to make the world aware of it you also could",
    Flags: quicli.Flags{
        {Name: "count", Default: 1, Description: "how many times I want to say it. Sometimes repetition is the key"},
        {Name: "foreground", Description: "change foreground color", SharedSubcommand: quicli.SubcommandSet{"color"}},
        {Name: "say", Default: "hello", Description: "say something. If you are polite start with a greeting"},
        {Name: "world", Description: "announce it to the world"},
        {Name: "surprise", Description: "you will see my friend", SharedSubcommand: quicli.SubcommandSet{"toto", "color"}, NotForRootCommand: true},
    },
    Function: Main,
    Subcommands: quicli.Subcommands{
        {Name: "color", Description: "print coloured message", Function: Color},
        {Name: "toto", Description: "??", Function: Toto},
    },
}
cli.RunWithSubcommand()
```

Root help (`./say-hello --help`):
```
Say Hello... or not. If you want to make the world aware of it you also could

Usage: say-hello [command] [flags]
Available commands: color, toto

--count  -c   how many times I want to say it. (default: 1) [env: SAY_HELLO_COUNT]
--say    -s   say something. (default: "hello") [env: SAY_HELLO_SAY]
--world  -w   announce it to the world. (default: false) [env: SAY_HELLO_WORLD]

Use "say-hello --help" for more information about the command.
```

Subcommand help (`./say-hello color --help`):
```
Say Hello... or not. ...

Usage: say-hello [command] [flags]
Command color: print coloured message

--count       -c   how many times I want to say it. (default: 1) [env: SAY_HELLO_COUNT]
--foreground  -f   change foreground color. (default: false) [env: SAY_HELLO_FOREGROUND]
--say         -s   say something. (default: "hello") [env: SAY_HELLO_SAY]
--surprise    -S   you will see my friend. (default: false) [env: SAY_HELLO_SURPRISE]
--world       -w   announce it to the world. (default: false) [env: SAY_HELLO_WORLD]
```

```bash
./say-hello color --foreground --say "hello"
./say-hello toto --surprise
./say-hello clour          # quicli error: unknown subcommand 'clour', did you mean 'color'?
```
</details>

### Use flag values in code
```golang
cfg.GetIntFlag("count")        // --count  (int)
cfg.GetStringFlag("say")       // --say    (string)
cfg.GetBoolFlag("world")       // --world  (bool)
cfg.GetFloatFlag("ratio")      // --ratio  (float64)
cfg.GetStringSliceFlag("tags") // --tags a,b --tags c  ([]string)
// or use the short name: cfg.GetIntFlag("c")
```

**Custom short name** — override the auto-derived first letter:
```golang
{Name: "config", ShortName: "C", Default: "", Description: "config file"}
// registers --config and -C
```

**Per-subcommand exclusive flags** — flags that only exist for one subcommand:
```golang
Subcommands: quicli.Subcommands{
    {
        Name: "push", Description: "push changes", Function: Push,
        Flags: quicli.Flags{
            {Name: "force", Description: "force push"},
        },
    },
},
```

**Typo detection** — if a user types an unknown subcommand, quicli automatically suggests the closest match:
```
$ mytool pish
quicli error: unknown subcommand 'pish', did you mean 'push'?
```

### Env vars

Every flag automatically reads from an env var as fallback before using the default.
Auto-derived name: `PROGNAME_FLAGNAME` (uppercase, non-alphanumeric → `_`).

```bash
SAY_HELLO_COUNT=5 ./say-hello   # same as --count 5
```

Override the env var name per flag:
```golang
{Name: "token", Default: "", Description: "API token", EnvVar: "MY_API_TOKEN"}
```

Opt a flag out of env var lookup:
```golang
{Name: "secret", Default: "", Description: "...", EnvVar: "-"}
```

The env var name is shown in help output: `(default: 0) [env: SAY_HELLO_COUNT]`

### Shell completion

Every CLI built with quicli gets `--completion <shell>` for free:

```bash
./say-hello --completion bash >> ~/.bash_completion
./say-hello --completion zsh  > ~/.zsh/completions/_say-hello
./say-hello --completion fish > ~/.config/fish/completions/say-hello.fish
```

Get more  [examples](examples/)

### Disclaimer
The library is a wrapper of the built-in go `flag` package. It should only be used to quickly built CLI and it is not intented for complex CLI usage.
