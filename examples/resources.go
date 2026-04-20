//go:build ignore

// resources — a fake resource manager demonstrating the subcommand pattern:
// get, list, delete with shared and exclusive flags, aliases, and typo detection.
//
// Try:
//   go run examples/resources.go --help
//   go run examples/resources.go get --help
//   go run examples/resources.go get --id abc123
//   go run examples/resources.go g --id abc123 --verbose
//   go run examples/resources.go list --output json
//   go run examples/resources.go ls
//   go run examples/resources.go delete --id abc123
//   go run examples/resources.go delet    # typo detection
//   RESOURCES_VERBOSE=true go run examples/resources.go list

package main

import (
	"fmt"
	"strings"

	q "github.com/ariary/quicli/pkg/quicli"
)

var db = map[string]string{
	"abc123": "widget",
	"def456": "gadget",
	"ghi789": "doohickey",
}

func cmdGet(cfg q.Config) {
	id := cfg.GetStringFlag("id")
	if id == "" {
		fmt.Println("error: --id is required")
		return
	}
	name, ok := db[id]
	if !ok {
		fmt.Printf("not found: %s\n", id)
		return
	}
	if cfg.GetBoolFlag("verbose") {
		fmt.Printf("id:   %s\nname: %s\n", id, name)
	} else {
		fmt.Println(name)
	}
}

func cmdList(cfg q.Config) {
	output := cfg.GetStringFlag("output")
	verbose := cfg.GetBoolFlag("verbose")
	switch output {
	case "json":
		parts := []string{}
		for id, name := range db {
			parts = append(parts, fmt.Sprintf(`{"id":%q,"name":%q}`, id, name))
		}
		fmt.Println("[" + strings.Join(parts, ",") + "]")
	default:
		for id, name := range db {
			if verbose {
				fmt.Printf("%-10s %s\n", id, name)
			} else {
				fmt.Println(name)
			}
		}
	}
}

func cmdDelete(cfg q.Config) {
	id := cfg.GetStringFlag("id")
	if id == "" {
		fmt.Println("error: --id is required")
		return
	}
	if _, ok := db[id]; !ok {
		fmt.Printf("not found: %s\n", id)
		return
	}
	delete(db, id)
	fmt.Printf("deleted %s\n", id)
}

func cmdRoot(cfg q.Config) {
	fmt.Println("use a subcommand: get, list, delete")
}

func main() {
	cli := q.Cli{
		Usage:       "resources [command] [flags]",
		Description: "Manage resources",
		Flags: q.Flags{
			{Name: "verbose", Description: "verbose output",
				SharedSubcommand: q.SubcommandSet{"get", "list", "delete"}},
			{Name: "output", Default: "text", Description: "output format (text, json)",
				SharedSubcommand: q.SubcommandSet{"list"}},
		},
		Function: cmdRoot,
		Subcommands: q.Subcommands{
			{
				Name: "get", Aliases: q.Aliases("g"),
				Description: "get a resource by id",
				Function:    cmdGet,
				Flags: q.Flags{
					{Name: "id", Default: "", Description: "resource id"},
				},
			},
			{
				Name: "list", Aliases: q.Aliases("ls"),
				Description: "list all resources",
				Function:    cmdList,
			},
			{
				Name:        "delete",
				Description: "delete a resource by id",
				Function:    cmdDelete,
				Flags: q.Flags{
					{Name: "id", Default: "", Description: "resource id"},
				},
			},
		},
	}
	cli.RunWithSubcommand()
}
