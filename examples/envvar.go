//go:build ignore

// envvar — demonstrates environment variable features:
//   - automatic env var fallback (PROGNAME_FLAGNAME)
//   - custom env var name via env:"MY_VAR" tag
//   - env-only flags for secrets (never shown in help or accepted on CLI)
//   - opting out of env var with env:"-"
//
// Try:
//   go run examples/envvar.go --help
//   go run examples/envvar.go --host localhost --port 5432
//   DB_HOST=prod-db go run examples/envvar.go --port 5432
//   DB_HOST=prod-db DB_TOKEN=secret123 go run examples/envvar.go
//   go run examples/envvar.go --host localhost --debug-options

package main

import (
	"fmt"

	q "github.com/ariary/quicli/pkg/quicli"
)

type ConnectOpts struct {
	Host  string `cli:"database host"          default:"localhost" env:"DB_HOST"`
	Port  int    `cli:"database port"          default:"5432"      env:"DB_PORT"`
	Token string `cli:"auth token (env only)"  default:""          env:"only:DB_TOKEN"`
	Query string `cli:"SQL query to run"       default:"SELECT 1"  env:"-"`
}

func main() {
	q.RunFunc("dbping [flags]", "Connect to a database and run a query", func(o ConnectOpts) {
		masked := "***"
		if o.Token == "" {
			masked = "(none)"
		}
		fmt.Printf("host:  %s\n", o.Host)
		fmt.Printf("port:  %d\n", o.Port)
		fmt.Printf("token: %s\n", masked)
		fmt.Printf("query: %s\n", o.Query)
	})
}
