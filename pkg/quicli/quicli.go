package quicli

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/ariary/go-utils/pkg/color"
)

const QUICLI_ERROR_PREFIX = "quicli error: "

// struct representing a cli flag
type Flag struct {
	Name        string
	Description string
	//Default is use to determine the flag value type and must be defined
	Default           any
	NoShortName       bool
	ShortName         string        // overrides auto first-letter derivation
	NotForRootCommand bool
	SharedSubcommand  SubcommandSet
	EnvVar            string        // env var override (activated in PR2)
}

type Flags []Flag

type Config struct {
	Flags map[string]any
	Args  []string
}

// Cheat Sheet
type Example struct {
	Title       string
	CommandLine string
}

type Examples []Example

type Runner func(Config)

// struct representing CLI
type Cli struct {
	Usage       string
	Description string
	Flags       []Flag
	Function    Runner
	CheatSheet  []Example
	Subcommands []Subcommand
}

// return the int value of an interger flag
func (c Config) GetIntFlag(name string) int {
	elem := c.Flags[name]
	if elem == nil {
		fmt.Println(QUICLI_ERROR_PREFIX, "failed to retrieve value for flag:", name)
		os.Exit(92)
	}
	i := reflect.ValueOf(elem).Interface().(*int)
	return *i
}

// return the string value of a string flag
func (c Config) GetStringFlag(name string) string {
	elem := c.Flags[name]
	if elem == nil {
		fmt.Println(QUICLI_ERROR_PREFIX, "failed to retrieve value for flag:", name)
		os.Exit(92)
	}
	str := reflect.ValueOf(elem).Interface().(*string)
	return *str
}

// return the string value of a string flag
func (c Config) GetBoolFlag(name string) bool {
	elem := c.Flags[name]
	if elem == nil {
		fmt.Println(QUICLI_ERROR_PREFIX, "failed to retrieve value for flag:", name)
		os.Exit(92)
	}
	boolean := reflect.ValueOf(elem).Interface().(*bool)
	return *boolean
}

// GetFloatFlag returns the float64 value of a float64 flag.
func (c Config) GetFloatFlag(name string) float64 {
	elem := c.Flags[name]
	if elem == nil {
		fmt.Println(QUICLI_ERROR_PREFIX, "failed to retrieve value for flag:", name)
		os.Exit(92)
	}
	f := reflect.ValueOf(elem).Interface().(*float64)
	return *f
}

// Parse: parse the different flags and return the struct containing the flag values.
// This is the core of the library. All the logic is within
func (c *Cli) Parse() (config Config) {
	usage := new(strings.Builder)
	wUsage := new(tabwriter.Writer)
	wUsage.Init(usage, 2, 8, 1, '\t', 1)
	var shorts []string
	config.Flags = make(map[string]any)

	//Description
	// usage += c.Description + "\n\nUsage: " + c.Usage + "\n\n"
	fmt.Fprintf(wUsage, color.Yellow(c.Description)+"\n\nUsage: "+c.Usage+"\n\n")

	//flags
	fs := flag.CommandLine
	fp := c.Flags
	for i := 0; i < len(fp); i++ {
		flag := fp[i]
		// prepation checks
		if len(flag.Name) == 0 {
			fmt.Println(QUICLI_ERROR_PREFIX + "empty flag name defintion")
			os.Exit(2)
		}
		//check Default => if no value provided assume it is a bool flag
		if flag.Default == nil {
			flag.Default = false
		}

		switch flag.Default.(type) {
		case int:
			createIntFlag(config, flag, &shorts, wUsage, fs)
		case string:
			createStringFlag(config, flag, &shorts, wUsage, fs)
		case bool:
			createBoolFlag(config, flag, &shorts, wUsage, fs)
		case float64:
			createFloatFlag(config, flag, &shorts, wUsage, fs)
			//todo: add float64;multiple value
		default:
			fmt.Println(QUICLI_ERROR_PREFIX+"Unknown flag type:", flag.Default)
			os.Exit(2)
		}
	}
	fmt.Fprintf(wUsage, "\nUse \""+color.Yellow(os.Args[0])+" --help\" for more information about the command.\n")

	//cheat sheet pt1
	var cheatSheet bool
	if len(c.CheatSheet) > 0 {
		fmt.Fprintf(wUsage, "\nSee command examples with \""+color.Yellow(os.Args[0])+" --cheat-sheet\"\n")
		flag.BoolVar(&cheatSheet, "cheat-sheet", false, "print cheat sheet")
		flag.BoolVar(&cheatSheet, "cs", false, "print cheat sheet")
	}

	wUsage.Flush()
	flag.Usage = func() { fmt.Print(usage.String()) }
	flag.Parse()
	config.Args = flag.Args()

	//cheat sheet pt2
	if len(c.CheatSheet) > 0 && cheatSheet {
		c.PrintCheatSheet()
		os.Exit(0)
	}

	return config
}

// Run: parse the different flags and run the function of the cli. Users have to define it, this is the core/logic of their application
func (c *Cli) Run() {
	config := c.Parse()

	// run
	if c.Function != nil {
		c.Function(config)
	} else {
		fmt.Println(QUICLI_ERROR_PREFIX + "you must define Function attribute for the Cli struct if you use Run function, otherwise use Parse")
	}

}

// Parse: parse the CLI given in parameter
func Parse(c Cli) (config Config) {
	return c.Parse()
}

// Run: run the application corresponding of the CLI given as parameter
func Run(c Cli) {
	c.Run()
}

// PrintCheatSheet: print the cheat sheet of the command
func (c *Cli) PrintCheatSheet() {
	fmt.Println(color.BlueForeground("Cheat Sheet") + "\n")
	examples := c.CheatSheet
	for i := 0; i < len(examples); i++ {
		example := examples[i]
		fmt.Println(color.Teal("~> " + example.Title + ":"))
		fmt.Println(example.CommandLine)
		fmt.Println()
	}
}

