//go:build ignore

// sayhello_runfunc_subcommand — demonstrates NewSubcommand: define per-subcommand
// structs with cli tags, pass typed functions, done. No flag registration, no
// GetXxxFlag calls, no manual Subcommand.Flags slice.
//
// Try:
//
//	go run examples/sayhello_runfunc_subcommand.go --help
//	go run examples/sayhello_runfunc_subcommand.go --count 3
//	go run examples/sayhello_runfunc_subcommand.go color --foreground
//	go run examples/sayhello_runfunc_subcommand.go whisper --say "shhh" --times 2

package main

import (
	"fmt"
	"strings"

	q "github.com/ariary/quicli/pkg/quicli"
)

type MainOpts struct {
	Count int    `cli:"how many times to say it" default:"1"`
	Say   string `cli:"what to say"              default:"hello"`
}

type ColorOpts struct {
	Foreground bool `cli:"use foreground color instead of background"`
}

type WhisperOpts struct {
	Say   string `cli:"what to whisper" default:"psst"`
	Times int    `cli:"how many times"  default:"1"`
}

func main() {
	colorSub := q.NewSubcommand("color", "print in red", func(o ColorOpts) {
		if o.Foreground {
			fmt.Println("\033[31mHELLO WORLD\033[0m") // red foreground
		} else {
			fmt.Println("\033[41mHELLO WORLD\033[0m") // red background
		}
	})
	colorSub.Aliases = q.Aliases("co")

	cli := q.Cli{
		Usage:       "say-hello [command] [flags]",
		Description: "Say Hello to the world — zero-boilerplate subcommands",
		Flags: q.Flags{
			{Name: "count", Default: 1, Description: "how many times (root command)"},
		},
		Function: func(cfg q.Config) {
			for i := 0; i < cfg.GetIntFlag("count"); i++ {
				fmt.Println("HELLO WORLD")
			}
		},
		Subcommands: q.Subcommands{
			colorSub,
			q.NewSubcommand("whisper", "say something quietly", func(o WhisperOpts) {
				for i := 0; i < o.Times; i++ {
					fmt.Println(strings.ToLower(o.Say) + "…")
				}
			}),
		},
	}
	cli.RunWithSubcommand()
}
