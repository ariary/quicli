package main

import (
	"fmt"
	"time"

	"github.com/ariary/go-utils/pkg/color"
	q "github.com/ariary/quicli/pkg/quicli"
)

func Color(cfg q.Config) {
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

func Main(cfg q.Config) {
	for i := 0; i < cfg.GetIntFlag("count"); i++ {
		fmt.Println("HELLO WORLD")
	}
}

func Toto(cfg q.Config) {
	if cfg.GetBoolFlag("surprise") {
		for {
			fmt.Println("Toto?!")
		}
	} else {
		fmt.Println("Toto?!")
	}

}

func Titi(cfg q.Config) {
	fmt.Println("🔥🔥🔥🔥🔥🔥🔥")
}

func main() {
	cli := q.Cli{
		Usage:       "SayToTheWorld [command] [flags]",
		Description: "Say Hello... or not. If you want to make the world aware of it you also could",
		Flags: q.Flags{
			{Name: "count", Default: 1, Description: "how many times I want to say it. Sometimes repetition is the key"},
			{Name: "foreground", Description: "change foreground background", ForSubcommand: q.SubcommandSet{"color"}},
			{Name: "say", Default: "hello", Description: "say something. If you are polite start with a greeting"},
			{Name: "world", Description: "announce it to the world"},
			{Name: "surprise", Description: "you will see my friend", ForSubcommand: q.SubcommandSet{"toto", "color"}, NotForRootCommand: true},
		},
		Function: Main,
		Subcommands: q.Subcommands{
			{Name: "color", Aliases: q.Aliases("co", "x"), Description: "print coloured message", Function: Color},
			{Name: "toto", Description: "??", Function: Toto},
			{Name: "🔥", Aliases: q.Aliases("🧯"), Description: "try me", Function: Titi},
		},
	}
	cli.RunWithSubcommand()
}
