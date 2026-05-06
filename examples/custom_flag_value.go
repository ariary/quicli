//go:build ignore

// custom_flag_value — demonstrates using a custom type that implements
// flag.Value for structured parsing. Any type whose pointer receiver
// implements Set(string) error and String() string works out of the box.
//
// Try:
//   go run examples/custom_flag_value.go --help
//   go run examples/custom_flag_value.go --listen 0.0.0.0:9090
//   go run examples/custom_flag_value.go --listen :3000 --prefix /api

package main

import (
	"fmt"
	"strings"

	q "github.com/ariary/quicli/pkg/quicli"
)

// Addr is a custom flag type for host:port pairs.
type Addr struct {
	Host string
	Port string
}

func (a *Addr) String() string {
	if a.Host == "" && a.Port == "" {
		return ""
	}
	return a.Host + ":" + a.Port
}

func (a *Addr) Set(s string) error {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("expected host:port, got %q", s)
	}
	a.Host = parts[0]
	a.Port = parts[1]
	return nil
}

type ServeOpts struct {
	Listen Addr   `cli:"address to listen on" default:"localhost:8080"`
	Prefix string `cli:"URL path prefix"      default:"/"`
}

func main() {
	q.RunFunc("serve [flags]", "Start an HTTP server", func(o ServeOpts) {
		fmt.Printf("listening on %s\n", o.Listen.String())
		fmt.Printf("prefix: %s\n", o.Prefix)
	})
}
