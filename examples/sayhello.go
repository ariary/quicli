package main

import (
	"fmt"

	"github.com/ariary/quicli/pkg/quicli"
)

func main() {
	cli := quicli.Cli{
		Usage:       "SayToTheWorld [flags]",
		Description: "Say Hello... or not",
		Flags: quicli.Flags{
			{Name: "count", Default: 1, Description: "how many times I will say it"},
			{Name: "say", Default: "hello", Description: "say something"},
			{Name: "world", Default: true, Description: "announce it to the world"},
		},
	}
	cfg := cli.Parse()

	for i := 0; i < cfg.GetIntFlag("count"); i++ {
		if cfg.GetBoolFlag("world") {
			fmt.Print("Message for the world: ")
		}
		fmt.Println(cfg.GetStringFlag("say"))
	}

}
