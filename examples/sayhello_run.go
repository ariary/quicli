package main

import (
	"fmt"

	"github.com/ariary/quicli/pkg/quicli"
)

//SayHello: contains the logic of the CLi application
func SayHello(cfg quicli.Config) {
	for i := 0; i < cfg.GetIntFlag("count"); i++ {
		if cfg.GetBoolFlag("world") {
			fmt.Print("Message for the world: ")
		}
		fmt.Println(cfg.GetStringFlag("say"))
	}
}

func main() {
	cli := quicli.Cli{
		Usage:       "SayToTheWorld [flags]",
		Description: "Say Hello... or not. If you want to make the world aware of it you also could",
		Flags: quicli.Flags{
			{Name: "count", Default: 1, Description: "how many times I want to say it. Sometimes repetition is the key"},
			{Name: "say", Default: "hello", Description: "say something. If you are polite start with a greeting"},
			{Name: "world", Description: "announce it to the world"},
		},
		Function: SayHello,
	}
	cli.Run()
}

// ALTERNATIVE
// quicli.Run(quicli.Cli{
// 	Usage:       "SayToTheWorld [flags]",
// 	Description: "Say Hello... or not. If you want to make the world aware of it you also could",
// 	Flags: quicli.Flags{
// 		{Name: "count", Default: 1, Description: "how many times I want to say it. Sometimes repetition is the key"},
// 		{Name: "say", Default: "hello", Description: "say something. If you are polite start with a greeting"},
// 		{Name: "world", Description: "announce it to the world"},
// 	},
// 	Function: SayHello,
// })
