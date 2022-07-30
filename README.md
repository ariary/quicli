## 🏃⌨️ quicli
### Build CLI in one line
<sup>..or two</sup>
```golang
cli := quicli.Cli{
  Name:        "SayToTheWorld",
  Usage:       "SayToTheWorld [flags]",
  Description: "Say Hello... or not",
  Flags: quicli.Flags{
    {Name: "count", Default: 1, Description: "how many times I will say it"},
    {Name: "say", Default: "hello", Description: "say somhething"},
    {Name: "world", Default: true, Description: "say it to the world"},
  },
}
cfg := cli.Parse()
```

With this code you obtain the following help message:
```
Usage: SayToTheWorld [flags]
Description: Say Hello... or not

--count -c      how many times I will say it
--say   -s      say something
--world -w      announce it to the world
```

### Use flags value in code
```golang
cfg.GetIntFlag("count") // get the --count flag value
// or alternatively
cfg.GetIntFlag("c")
```
