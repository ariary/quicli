package quicli

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/ariary/go-utils/pkg/color"
)

// Subcommand
type Subcommand struct {
	Name        string
	Aliases     []string
	Description string
	Function    Runner
	Flags       []Flag // flags exclusive to this subcommand
}

func Aliases(aliases ...string) []string {
	return aliases
}

type Subcommands []Subcommand

type SubcommandSet []string

// RunWithSubcommand: equivalent of Run function when cli has subcommand defined
func (c *Cli) RunWithSubcommand() {
	var config Config
	usage := new(strings.Builder)
	wUsage := new(tabwriter.Writer)
	wUsage.Init(usage, 2, 8, 1, '\t', 1)
	var shorts []string
	config.Flags = make(map[string]any)
	fs := flag.NewFlagSet("parser", flag.ExitOnError)

	//Description
	if isRootCommand(c.Subcommands) {
		if len(c.Subcommands) > 0 {
			subcommandSet := []string{}
			for _, sub := range c.Subcommands {
				subcommandSet = append(subcommandSet, sub.Name)
				subcommandSet = append(subcommandSet, sub.Aliases...)
			}
			fmt.Fprintf(wUsage, color.Yellow(c.Description)+"\n\nUsage: "+c.Usage+"\nAvailable commands: "+strings.Join(subcommandSet, ", ")+"\n\n")
		} else {
			fmt.Fprintf(wUsage, color.Yellow(c.Description)+"\n\nUsage: "+c.Usage+"\n\n")
		}
	} else {
		sub := getSubcommandByName(c.Subcommands, os.Args[1])
		fmt.Fprintf(wUsage, c.Description+"\n\nUsage: "+c.Usage+"\n"+"Command "+color.Cyan(sub.Name)+": "+sub.Description+"\n\n")
	}

	// misspelled subcommand detection
	if isRootCommand(c.Subcommands) && len(c.Subcommands) > 0 && len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		if closest := findClosestSubcommand(c.Subcommands, os.Args[1]); closest != "" {
			fmt.Println(QUICLI_ERROR_PREFIX + "unknown subcommand '" + os.Args[1] + "', did you mean '" + closest + "'?")
			os.Exit(2)
		}
	}

	//Subcommands preliminary checks
	if len(c.Subcommands) > 0 {
		checkSubcommandFunctionIsDefined(c)
		checkSubcommandAliasesUniqueness(c)
	}

	//flags
	for _, f := range c.Flags {
		// prepation checks
		if len(f.Name) == 0 {
			fmt.Println(QUICLI_ERROR_PREFIX + "empty flag name defintion")
			os.Exit(2)
		}
		//check Default => if no value provided assume it is a bool flag
		if f.Default == nil {
			f.Default = false
		}
		for _, subcommandName := range f.SharedSubcommand {
			if getSubcommandByName(c.Subcommands, subcommandName).Name == "" {
				fmt.Println(QUICLI_ERROR_PREFIX+"subcommand", subcommandName, "specified for flag", f.Name, "is not defined")
				os.Exit(2)
			}
		}

		// before other stuff add subcommand aliases for the flag..
		var flagForSub SubcommandSet
		for _, scName := range f.SharedSubcommand {
			subcommand := getSubcommandByName(c.Subcommands, scName)
			flagForSub = append(flagForSub, subcommand.Aliases...)
		}
		f.SharedSubcommand = append(f.SharedSubcommand, flagForSub...)

		switch f.Default.(type) {
		case int:
			if isRootCommand(c.Subcommands) && !f.NotForRootCommand {
				createIntFlag(config, f, &shorts, wUsage, fs)
			} else if len(os.Args) > 1 && f.isForSubcommand(os.Args[1]) {
				createIntFlag(config, f, &shorts, wUsage, fs)
			}
		case string:
			if isRootCommand(c.Subcommands) && !f.NotForRootCommand {
				createStringFlag(config, f, &shorts, wUsage, fs)
			} else if len(os.Args) > 1 && f.isForSubcommand(os.Args[1]) {
				createStringFlag(config, f, &shorts, wUsage, fs)
			}
		case bool:
			if isRootCommand(c.Subcommands) && !f.NotForRootCommand {
				createBoolFlag(config, f, &shorts, wUsage, fs)
			} else if len(os.Args) > 1 && f.isForSubcommand(os.Args[1]) {
				createBoolFlag(config, f, &shorts, wUsage, fs)
			}
		case float64:
			if isRootCommand(c.Subcommands) && !f.NotForRootCommand {
				createFloatFlag(config, f, &shorts, wUsage, fs)
			} else if len(os.Args) > 1 && f.isForSubcommand(os.Args[1]) {
				createFloatFlag(config, f, &shorts, wUsage, fs)
			}
		case []string:
			if isRootCommand(c.Subcommands) && !f.NotForRootCommand {
				createStringSliceFlag(config, f, &shorts, wUsage, fs)
			} else if len(os.Args) > 1 && f.isForSubcommand(os.Args[1]) {
				createStringSliceFlag(config, f, &shorts, wUsage, fs)
			}
		default:
			fmt.Println(QUICLI_ERROR_PREFIX+"Unknown flag type:", f.Default)
			os.Exit(2)
		}
	}
	// Register exclusive flags for the active subcommand
	if !isRootCommand(c.Subcommands) {
		sub := getSubcommandByName(c.Subcommands, os.Args[1])
		for _, f := range sub.Flags {
			if len(f.Name) == 0 {
				fmt.Println(QUICLI_ERROR_PREFIX + "empty flag name definition in subcommand " + sub.Name)
				os.Exit(2)
			}
			if f.Default == nil {
				f.Default = false
			}
			switch f.Default.(type) {
			case int:
				createIntFlag(config, f, &shorts, wUsage, fs)
			case string:
				createStringFlag(config, f, &shorts, wUsage, fs)
			case bool:
				createBoolFlag(config, f, &shorts, wUsage, fs)
			case float64:
				createFloatFlag(config, f, &shorts, wUsage, fs)
			case []string:
				createStringSliceFlag(config, f, &shorts, wUsage, fs)
			default:
				fmt.Println(QUICLI_ERROR_PREFIX+"Unknown flag type:", f.Default)
				os.Exit(2)
			}
		}
	}

	fmt.Fprintf(wUsage, "\nUse \""+color.Yellow(os.Args[0])+" --help\" for more information about the command.\n")

	//cheat sheet pt1
	var cheatSheet bool
	if len(c.CheatSheet) > 0 {
		fmt.Fprintf(wUsage, "\nSee command examples with \""+os.Args[0]+" --cheat-sheet\"\n")
		fs.BoolVar(&cheatSheet, "cheat-sheet", false, "print cheat sheet")
		fs.BoolVar(&cheatSheet, "cs", false, "print cheat sheet")
	}

	var completionShell string
	fs.StringVar(&completionShell, "completion", "", "generate shell completion script (bash, zsh, fish)")

	wUsage.Flush()
	// Parse
	fs.Usage = func() { fmt.Print(usage.String()) }
	if isRootCommand(c.Subcommands) && len(os.Args) > 1 {
		fs.Parse(os.Args[1:])
	} else if len(os.Args) > 2 {
		fs.Parse(os.Args[2:])
	}
	config.Args = fs.Args()
	allFlags := c.Flags
	if !isRootCommand(c.Subcommands) {
		sub := getSubcommandByName(c.Subcommands, os.Args[1])
		allFlags = append(allFlags, sub.Flags...)
	}
	applyEnvVars(allFlags, fs)

	if completionShell != "" {
		script, err := generateCompletion(c, completionShell)
		if err != nil {
			fmt.Fprintln(os.Stderr, QUICLI_ERROR_PREFIX+err.Error())
			os.Exit(1)
		}
		fmt.Print(script)
		os.Exit(0)
	}

	//cheat sheet pt2
	if len(c.CheatSheet) > 0 && cheatSheet {
		c.PrintCheatSheet()
		os.Exit(0)
	}

	// Run
	if isRootCommand(c.Subcommands) {
		c.Function(config)
	} else {
		getSubcommandByName(c.Subcommands, os.Args[1]).Function(config)
	}
}

// isRootCommand: return true if the command line is targetting the root command, false if it is targgeting a subcommand
func isRootCommand(subcommands Subcommands) bool {
	if len(os.Args) < 2 {
		return true
	} else {
		sub := getSubcommandByName(subcommands, os.Args[1])
		return sub.Name == ""
	}
}

// getSubcommandByName: return the subcommand with name (take into account aliases)
func getSubcommandByName(subcommands Subcommands, subcommandName string) (sub Subcommand) {
	for _, s := range subcommands {
		if subcommandName == s.Name {
			return s
		}
		if slices.Contains(s.Aliases, subcommandName) {
			return s
		}
	}
	return sub
}

// isForSubcommand: return true if the subcommand is concerned by the flag
func (f *Flag) isForSubcommand(subcommandName string) bool {
	return slices.Contains(f.SharedSubcommand, subcommandName)
}

// checkSubcommandFunctionIsDefined: assert the subcommmand Function is filled, exit otherwise
func checkSubcommandFunctionIsDefined(c *Cli) {
	for _, sub := range c.Subcommands {
		if sub.Function == nil {
			fmt.Println(QUICLI_ERROR_PREFIX+"subcommand", sub.Name, "does not define mandatory 'Function' attribute")
			os.Exit(2)
		}
	}
}

// checkSubcommandAliasesUniqueness: assert the subcommand Aliases are unique (ie not same alias for two different subcommands), exit otherwise
func checkSubcommandAliasesUniqueness(c *Cli) {
	seen := map[string]bool{}
	for _, sub := range c.Subcommands {
		var duplicates []string
		for _, a := range sub.Aliases {
			if seen[a] {
				duplicates = append(duplicates, a)
			} else {
				seen[a] = true
			}
		}
		if len(duplicates) > 0 {
			fmt.Println(QUICLI_ERROR_PREFIX+"subcommand", sub.Name, "define some already defined aliases ('", strings.Join(duplicates, ","), "')")
			os.Exit(2)
		}
	}
}
