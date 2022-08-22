package main

import (
	"fmt"
	"time"

	"github.com/ariary/go-utils/pkg/color"
	"github.com/ariary/quicli/pkg/quicli"
)

func Color(cfg quicli.Config) {
	if cfg.GetBoolFlag("surprise") {
		for {
			if cfg.GetBoolFlag("foreground") {
				fmt.Println(color.RedForeground("HELLO WORLD"))
			} else {
				fmt.Println(color.Red("HELLO WORLD"))
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		if cfg.GetBoolFlag("foreground") {
			fmt.Println(color.RedForeground("HELLO WORLD"))
		} else {
			fmt.Println(color.Red("HELLO WORLD"))
		}
	}

}

func Main(cfg quicli.Config) {
	for i := 0; i < cfg.GetIntFlag("count"); i++ {
		fmt.Println("HELLO WORLD")
	}
}

func Toto(cfg quicli.Config) {
	if cfg.GetBoolFlag("surprise") {
		for {
			fmt.Println("Toto?!")
		}
	} else {
		fmt.Println("Toto?!")
	}

}

func main() {
	cli := quicli.Cli{
		Usage:       "SayToTheWorld [command] [flags]",
		Description: "Say Hello... or not. If you want to make the world aware of it you also could",
		Flags: quicli.Flags{
			{Name: "count", Default: 1, Description: "how many times I want to say it. Sometimes repetition is the key"},
			{Name: "foreground", Description: "change foreground background", ForSubcommand: quicli.SubcommandSet{"color"}},
			{Name: "say", Default: "hello", Description: "say something. If you are polite start with a greeting"},
			{Name: "world", Description: "announce it to the world"},
			{Name: "surprise", Description: "you will see my friend", ForSubcommand: quicli.SubcommandSet{"toto", "color"}, NotForRootCommand: true},
		},
		Function: Main,
		Subcommands: quicli.Subcommands{
			{Name: "color", Description: "print coloured message", Function: Color},
			{Name: "toto", Description: "??", Function: Toto},
		},
	}
	cli.RunWithSubcommand()
}
