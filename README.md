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
    <summary>Real one-line (Parse and run)</summary>

```golang
quicli.Run(quicli.Cli{Usage:"SayToTheWorld [flags]",Description: "Say Hello... or not. If you want to make the world aware of it you also could",Flags: quicli.Flags{{Name: "count", Default: 1, Description: "how many times I want to say it. Sometimes repetition is the key"},{Name: "say", Default: "hello", Description: "say something. If you are polite start with a greeting"},{Name: "world", Description: "announce it to the world"},},Function: SayHello,})
```
</details>

### Use flag values in code
```golang
cfg.GetIntFlag("count") // get the --count flag value
// or alternatively
cfg.GetIntFlag("c")
```

Get more  [examples](examples/)

### Disclaimer
The library is a wrapper of the built-in go `flag` package. It should only be used to quickly built CLI and it is not intented for complex CLI usage.
