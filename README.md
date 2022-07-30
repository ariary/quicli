## 🏃⌨️ quicli
### Build CLI in one line
<sup>..or two</sup>
```golang
cli := quicli.Cli{
  Name:        "SayToTheWorld",
  Usage:       "saytotheworld [flags]",
  Description: "Say Hello... or not",
  Flags: quicli.Flags{
    {Name: "count", Default: 1, Description: "How many times I will say it"},
    {Name: "say", Default: "hello", Description: "Say something"},
    {Name: "world", Default: true, Description: "Say it to the world"},
  },
}
cfg := cli.Parse()
```

With this code you obtain the following help message:
```
Usage: SayToTheWorld [flags]
Description: Say Hello... or not

--count -c      How many times I will say it
--say   -s      Say something
--world -w      to the world
```

### Use flags value in code
```golang
cfg.GetIntFlag("count") // get the --count flag value
// or alternatively
cfg.GetIntFlag("c")
```
