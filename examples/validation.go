//go:build ignore

// validation — demonstrates required flags and choices constraints.
// quicli validates after all sources (CLI, env vars) have been applied
// and prints clear error messages on failure.
//
// Try:
//   go run examples/validation.go --help
//   go run examples/validation.go --name Alice --env prod
//   go run examples/validation.go --env prod                    # error: --name is required
//   go run examples/validation.go --name Alice --env staging    # error: invalid choice
//   go run examples/validation.go --name Alice                  # uses default env "dev"

package main

import (
	"fmt"

	q "github.com/ariary/quicli/pkg/quicli"
)

type DeployOpts struct {
	Name    string `cli:"service name to deploy"         required:"true"`
	Env     string `cli:"target environment"             default:"dev"  choices:"dev,prod"`
	Replicas int   `cli:"number of replicas"             default:"1"`
	DryRun  bool   `cli:"simulate without applying"`
}

func main() {
	q.RunFunc("deploy [flags]", "Deploy a service to an environment", func(o DeployOpts) {
		if o.DryRun {
			fmt.Printf("[dry-run] would deploy %s to %s with %d replica(s)\n", o.Name, o.Env, o.Replicas)
			return
		}
		fmt.Printf("deploying %s to %s with %d replica(s)...\n", o.Name, o.Env, o.Replicas)
		fmt.Println("done")
	})
}
