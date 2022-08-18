package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ariary/quicli/pkg/quicli"
)

//Say: contains the logic of the CLi application
func Say(cfg quicli.Config) {
	if len(cfg.Args) < 1 {
		fmt.Println("Please provide message as argument see --help")
		os.Exit(1)
	}
	message := strings.Join(cfg.Args[0:], " ")
	for i := 0; i < cfg.GetIntFlag("count"); i++ {
		if cfg.GetBoolFlag("world") {
			fmt.Print("Message for the world: ")
		}
		fmt.Println(message)
	}
}

func main() {
	cli := quicli.Cli{
		Usage:       "SayToTheWorld [flags] [message]",
		Description: "Say Hello... or not. If you want to make the world aware of it you also could",
		Flags: quicli.Flags{
			{Name: "count", Default: 1, Description: "how many times I want to say it. Sometimes repetition is the key"},
			{Name: "world", Description: "announce it to the world"},
		},
		Function: Say,
	}
	quicli.Run(cli)
}
