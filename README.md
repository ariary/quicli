### Build CLi in one line
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
Usage of SayToTheWorld: saytotheworld [flags]
Make http request from raw request. [url] is required and on the form: [protocol]://[addr]:[port]
  -c, --count     How many times I will say it
  -s, --say		    Say something
  -w, --world	    Say it to the world
```

### Use flags value in code
```golang
cfg.GetIntFlag("count") // get the --count flag value
// or alternatively
cfg.GetIntFlag("c")
```
