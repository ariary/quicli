//go:build ignore

// duration — demonstrates time.Duration flag support.
// Duration flags accept Go duration strings like "5s", "2m30s", "1h".
//
// Try:
//   go run examples/duration.go --help
//   go run examples/duration.go --timeout 5s --interval 500ms
//   go run examples/duration.go --timeout 2m --retries 5
//   go run examples/duration.go                                 # uses defaults

package main

import (
	"fmt"
	"time"

	q "github.com/ariary/quicli/pkg/quicli"
)

type PollOpts struct {
	URL      string        `cli:"endpoint to poll"             default:"http://localhost:8080/health"`
	Timeout  time.Duration `cli:"max time to wait"             default:"30s"`
	Interval time.Duration `cli:"time between attempts"        default:"2s"`
	Retries  int           `cli:"max number of retries"        default:"10"`
}

func main() {
	q.RunFunc("poll [flags]", "Poll an endpoint until healthy", func(o PollOpts) {
		fmt.Printf("polling %s\n", o.URL)
		fmt.Printf("  timeout:  %s\n", o.Timeout)
		fmt.Printf("  interval: %s\n", o.Interval)
		fmt.Printf("  retries:  %d\n", o.Retries)

		deadline := time.Now().Add(o.Timeout)
		for i := 1; i <= o.Retries; i++ {
			if time.Now().After(deadline) {
				fmt.Println("timed out")
				return
			}
			fmt.Printf("  attempt %d/%d... ok\n", i, o.Retries)
			if i < o.Retries {
				time.Sleep(o.Interval)
			}
		}
		fmt.Println("healthy")
	})
}
