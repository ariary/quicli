//go:build ignore

// sayhello_runfunc — demonstrates RunFunc: define a struct with cli tags,
// pass a function, done. No flag registration, no GetXxxFlag calls.
//
// Try:
//   go run examples/sayhello_runfunc.go --help
//   go run examples/sayhello_runfunc.go --count 3 --say "bonjour"
//   go run examples/sayhello_runfunc.go -w --tags en --tags fr
//   go run examples/sayhello_runfunc.go --volume 0.3 --say "whisper"
//   SAY_HELLO_COUNT=5 go run examples/sayhello_runfunc.go

package main

import (
	"fmt"
	"strings"

	q "github.com/ariary/quicli/pkg/quicli"
)

// Opts defines the full CLI through struct tags — no Flags slice needed.
type Opts struct {
	Count  int      `cli:"how many times to say it"    default:"1"`
	Say    string   `cli:"what to say"                 default:"hello"`
	World  bool     `cli:"prefix with a world message"`
	Tags   []string `cli:"filter output by tags"`
	Volume float64  `cli:"output volume: 0.0 quiet, 1.0 loud" default:"1.0" short:"V"`
}

func main() {
	q.RunFunc("say-hello [flags]", "Say Hello to the world", func(o Opts) {
		prefix := ""
		if o.World {
			prefix = "🌍 "
		}

		for i := 0; i < o.Count; i++ {
			line := prefix + o.Say
			if len(o.Tags) > 0 {
				line += " [" + strings.Join(o.Tags, ", ") + "]"
			}
			switch {
			case o.Volume == 0:
				// silent
			case o.Volume < 0.5:
				fmt.Println(strings.ToLower(line))
			case o.Volume == 1.0:
				fmt.Println(strings.ToUpper(line))
			default:
				fmt.Println(line)
			}
		}
	})
}
