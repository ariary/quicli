## 🏃⌨️ quicli
### Build CLI in one line
<sup>..or two</sup>


```golang
cli := quicli.Cli{Usage:"SayToTheWorld [flags]",Description: "Say Hello... or not. If you want to make the world aware of it you also could",Flags: quicli.Flags{{Name: "count", Default: 1, Description: "how many times I want to say it. Sometimes repetition is the key"},{Name: "say", Default: "hello", Description: "say something. If you are polite start with a greeting"},{Name: "world", Description: "announce it to the world"},},}
cfg := cli.Parse()
```

With this code you obtain the following help message:
```
Say Hello... or not. If you want to make the world aware of it you also could

Usage: SayToTheWorld [flags]

--count -c              how many times I want to say it. Sometimes repetition is the key. (default: 1)
--say   -s              say something. If you are polite start with a greeting. (default: "hello")
--world -w              announce it to the world. (default: false)

Use "./sayhello --help" for more information about the command.
```

<details>
    <summary>Pretty indented version</summary>

```golang
cli := quicli.Cli{
  Usage:       "SayToTheWorld [flags]",
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
quicli.Run(quicli.Cli{Usage:"SayToTheWorld [flags]",Description: "Say Hello... or not. If you want to make the world aware of it you also could",Flags: quicli.Flags{{Name: "count", Default: 1, Description: "how many times I want to say it. Sometimes repetition is the key"},{Name: "say", Default: "hello", Description: "say something. If you are polite start with a greeting"},{Name: "world", Description: "announce it to the world"},},Function: SayHello,})
```
</details>

<details>
    <summary>You want a subcommand pattern?! okay</summary>

```golang
	cli := quicli.Cli{
		Usage:       "SayToTheWorld [command] [flags]",
		Description: "Say Hello... or not. If you want to make the world aware of it you also could",
		Flags: quicli.Flags{
			{Name: "count", Default: 1, Description: "how many times I want to say it. Sometimes repetition is the key"},
			{Name: "foreground", Description: "change foreground background", SharedSubcommand: quicli.SubcommandSet{"color"}},
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
</details>

### Use flag values in code
```golang
cfg.GetIntFlag("count")    // --count  (int)
cfg.GetStringFlag("say")   // --say    (string)
cfg.GetBoolFlag("world")   // --world  (bool)
cfg.GetFloatFlag("ratio")  // --ratio  (float64)
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
./mycli --completion bash >> ~/.bash_completion
./mycli --completion zsh  > ~/.zsh/completions/_mycli
./mycli --completion fish > ~/.config/fish/completions/mycli.fish
```

Get more  [examples](examples/)

### Disclaimer
The library is a wrapper of the built-in go `flag` package. It should only be used to quickly built CLI and it is not intented for complex CLI usage.
