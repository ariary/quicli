package quicli

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/ariary/go-utils/pkg/color"
	stringSlice "github.com/ariary/go-utils/pkg/stringSlice"
	mapset "github.com/deckarep/golang-set/v2"
)

// Subcommand
type Subcommand struct {
	Name        string
	Aliases     mapset.Set[string]
	Description string
	Function    Runner
	// Flags       []Flag
}

func Aliases(aliases ...string) (aliasesSet mapset.Set[string]) {
	aliasesSet = mapset.NewSet[string]()
	aliasesSet.Append(aliases...)

	return aliasesSet
}

type Subcommands []Subcommand

type SubcommandSet []string

var AllAliases mapset.Set[string]

// RunWithSubcommand: equivalent of Run function when cli has subcommand defined
func (c *Cli) RunWithSubcommand() {
	var config Config
	usage := new(strings.Builder)
	wUsage := new(tabwriter.Writer)
	wUsage.Init(usage, 2, 8, 1, '\t', 1)
	var shorts []string
	config.Flags = make(map[string]interface{})
	fs := flag.NewFlagSet("parser", flag.ExitOnError)

	//Description
	if isRootCommand(c.Subcommands) {
		if len(c.Subcommands) > 0 {
			subcommandSet := []string{}
			for i := 0; i < len(c.Subcommands); i++ {
				//to do check that there isn't duplicate: subcommandSet in set
				subcommandSet = append(subcommandSet, c.Subcommands[i].Name) // add name
				if c.Subcommands[i].Aliases != nil {                         //add aliases
					subcommandSet = append(subcommandSet, c.Subcommands[i].Aliases.ToSlice()...)
				}
			}
			fmt.Fprintf(wUsage, color.Yellow(c.Description)+"\n\nUsage: "+c.Usage+"\nAvailable commands: "+strings.Join(subcommandSet, ", ")+"\n\n")
		} else {
			fmt.Fprintf(wUsage, color.Yellow(c.Description)+"\n\nUsage: "+c.Usage+"\n\n")
		}
	} else {
		//TODO: check if subcommand is misspelled
		sub := getSubcommandByName(c.Subcommands, os.Args[1])
		fmt.Fprintf(wUsage, c.Description+"\n\nUsage: "+c.Usage+"\n"+"Command "+color.Cyan(sub.Name)+": "+sub.Description+"\n\n")
	}

	//Subcommands preliminary checks
	if len(c.Subcommands) > 0 {
		checkSubcommandFunctionIsDefined(c)
		AllAliases = mapset.NewSet[string]()
		checkSubcommandAliasesUniqueness(c)
	}

	//flags
	fp := c.Flags
	for i := 0; i < len(fp); i++ {
		f := fp[i]
		// prepation checks
		if len(f.Name) == 0 {
			fmt.Println(QUICLI_ERROR_PREFIX + "empty flag name defintion")
			os.Exit(2)
		}
		//check Default => if no value provided assume it is a bool flag
		if f.Default == nil {
			f.Default = false
		}
		for i := 0; i < len(f.ForSubcommand); i++ {
			subcommandName := f.ForSubcommand[i]
			if getSubcommandByName(c.Subcommands, subcommandName).Name == "" {
				fmt.Println(QUICLI_ERROR_PREFIX+"subcommand", subcommandName, "specified for flag", f.Name, "is not defined")
				os.Exit(2)
			}
		}

		// before other stuff add subcommand aliases for the flag..
		var flagForSub SubcommandSet
		for i := 0; i < len(f.ForSubcommand); i++ {
			subcommand := getSubcommandByName(c.Subcommands, f.ForSubcommand[i])
			if subcommand.Aliases != nil {
				flagForSub = append(flagForSub, subcommand.Aliases.ToSlice()...)
			}

		}
		f.ForSubcommand = append(f.ForSubcommand, flagForSub...)

		switch f.Default.(type) {
		case int:
			if isRootCommand(c.Subcommands) && !f.NotForRootCommand {
				createIntFlagFs(config, f, &shorts, wUsage, fs)
			} else if len(os.Args) > 1 && f.isForSubcommand(os.Args[1]) {
				createIntFlagFs(config, f, &shorts, wUsage, fs)
			}
		case string:
			if isRootCommand(c.Subcommands) && !f.NotForRootCommand {
				createStringFlagFs(config, f, &shorts, wUsage, fs)
			} else if len(os.Args) > 1 && f.isForSubcommand(os.Args[1]) {
				createStringFlagFs(config, f, &shorts, wUsage, fs)
			}
		case bool:
			if isRootCommand(c.Subcommands) && !f.NotForRootCommand {
				createBoolFlagFs(config, f, &shorts, wUsage, fs)
			} else if len(os.Args) > 1 && f.isForSubcommand(os.Args[1]) {
				createBoolFlagFs(config, f, &shorts, wUsage, fs)
			}
		case float64:
			if isRootCommand(c.Subcommands) && !f.NotForRootCommand {
				createFloatFlagFs(config, f, &shorts, wUsage, fs)
			} else if len(os.Args) > 1 && f.isForSubcommand(os.Args[1]) {
				createFloatFlagFs(config, f, &shorts, wUsage, fs)
			}
		default:
			fmt.Println(QUICLI_ERROR_PREFIX+"Unknown flag type:", f.Default)
			os.Exit(2)
		}
	}
	fmt.Fprintf(wUsage, "\nUse \""+color.Yellow(os.Args[0])+" --help\" for more information about the command.\n")

	//cheat sheet pt1
	var cheatSheet bool
	if len(c.CheatSheet) > 0 {
		fmt.Fprintf(wUsage, "\nSee command examples with \""+os.Args[0]+" --cheat-sheet\"\n")
		flag.BoolVar(&cheatSheet, "cheat-sheet", false, "print cheat sheet")
		flag.BoolVar(&cheatSheet, "cs", false, "print cheat sheet")
	}

	wUsage.Flush()
	// Parse
	fs.Usage = func() { fmt.Print(usage.String()) }
	if isRootCommand(c.Subcommands) && len(os.Args) > 1 {
		fs.Parse(os.Args[1:])
	} else if len(os.Args) > 2 {
		fs.Parse(os.Args[2:])
	}
	config.Args = fs.Args()

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
	for i := 0; i < len(subcommands); i++ {
		if subcommandName == subcommands[i].Name {
			return subcommands[i]
		}
		if subcommands[i].Aliases != nil && subcommands[i].Aliases.Contains(subcommandName) {
			return subcommands[i]
		}
	}
	return sub
}

// isForSubcommand: return true if the subcommand is concerned by the flag
func (f *Flag) isForSubcommand(subcommandName string) bool {
	for i := 0; i < len(f.ForSubcommand); i++ {
		if subcommandName == f.ForSubcommand[i] { // look subcommand name
			return true
		}
	}
	return false
}

func createIntFlagFs(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet) {
	name := f.Name
	shortName := name[0:1]
	var intPtr int
	fs.IntVar(&intPtr, name, int(reflect.ValueOf(f.Default).Int()), f.Description)
	if !stringSlice.Contains(*shorts, shortName) && !f.NoShortName {
		fs.IntVar(&intPtr, shortName, int(reflect.ValueOf(f.Default).Int()), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName))
		cfg.Flags[shortName] = &intPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, ""))
	}
	cfg.Flags[name] = &intPtr
}

func createStringFlagFs(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet) {
	name := f.Name
	shortName := name[0:1]
	var strPtr string
	fs.StringVar(&strPtr, name, string(reflect.ValueOf(f.Default).String()), f.Description)
	if !stringSlice.Contains(*shorts, shortName) && !f.NoShortName {
		fs.StringVar(&strPtr, shortName, string(reflect.ValueOf(f.Default).String()), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName))
		cfg.Flags[shortName] = &strPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, ""))
	}
	cfg.Flags[name] = &strPtr
}

func createBoolFlagFs(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet) {
	name := f.Name
	shortName := name[0:1]
	var bPtr bool
	fs.BoolVar(&bPtr, name, bool(reflect.ValueOf(f.Default).Bool()), f.Description)
	cfg.Flags[name] = &bPtr
	if !stringSlice.Contains(*shorts, shortName) && !f.NoShortName {
		fs.BoolVar(&bPtr, shortName, bool(reflect.ValueOf(f.Default).Bool()), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName))
		cfg.Flags[shortName] = &bPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, ""))
	}
	cfg.Flags[name] = &bPtr
}

func createFloatFlagFs(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer, fs *flag.FlagSet) {
	name := f.Name
	shortName := name[0:1]
	var floatPtr float64
	fs.Float64Var(&floatPtr, name, float64(reflect.ValueOf(f.Default).Float()), f.Description)
	cfg.Flags[name] = &floatPtr
	if !stringSlice.Contains(*shorts, shortName) && !f.NoShortName {
		fs.Float64Var(&floatPtr, shortName, float64(reflect.ValueOf(f.Default).Float()), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName))
		cfg.Flags[shortName] = &floatPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, ""))
	}
	cfg.Flags[name] = &floatPtr
}

// checkSubcommandFunctionIsDefined: assert the subcommmand Function is filled, exit otherwise
func checkSubcommandFunctionIsDefined(c *Cli) {
	for i := 0; i < len(c.Subcommands); i++ {
		sub := c.Subcommands[i]
		if sub.Function == nil {
			fmt.Println(QUICLI_ERROR_PREFIX+"subcommand", sub.Name, "does not define mandatory 'Function' attribute")
			os.Exit(2)
		}
	}
}

// checkSubcommandFunctionIsDefined: assert the subcommmand Aliases are unique (ie not same alias for two different subcommands), exit otherwise
func checkSubcommandAliasesUniqueness(c *Cli) {
	for i := 0; i < len(c.Subcommands); i++ {
		subcommandAliases := c.Subcommands[i].Aliases
		if subcommandAliases != nil {
			commonAliases := AllAliases.Intersect(subcommandAliases)
			if commonAliases.Cardinality() == 0 {
				AllAliases.Append(subcommandAliases.ToSlice()...)
			} else {
				fmt.Println(QUICLI_ERROR_PREFIX+"subcommand", c.Subcommands[i].Name, "define some already defined aliases ('", strings.Join(commonAliases.ToSlice(), ","), "')")
				os.Exit(2)
			}
		}
	}
}
