package quicli

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/ariary/go-utils/pkg/color"
	stringSlice "github.com/ariary/go-utils/pkg/stringSlice"
)

//struct representing a cli flag
type Flag struct {
	Name        string
	Description string
	//Default is use to determine the flag value type and must be defined
	Default interface{}
}

type Flags []Flag

type Config map[string]interface{}

// Cheat Sheet
type Example struct {
	Title       string
	CommandLine string
}

type Examples []Example

// struct representing CLI
type Cli struct {
	Usage       string
	Description string
	Flags       []Flag
	CheatSheet  []Example
}

// return the int value of an interger flag
func (c Config) GetIntFlag(name string) int {
	elem := c[name]
	i := reflect.ValueOf(elem).Interface().(*int)
	return *i
}

// return the string value of a string flag
func (c Config) GetStringFlag(name string) string {
	elem := c[name]
	str := reflect.ValueOf(elem).Interface().(*string)
	return *str
}

// return the string value of a string flag
func (c Config) GetBoolFlag(name string) bool {
	elem := c[name]
	boolean := reflect.ValueOf(elem).Interface().(*bool)
	return *boolean
}

//Parse: parse the different flags and return the struct containing the flag values.
// This is the core of the library. All the logic is within
func (c *Cli) Parse() (config Config) {
	usage := new(strings.Builder)
	wUsage := new(tabwriter.Writer)
	wUsage.Init(usage, 2, 8, 1, '\t', 1)
	var shorts []string
	config = make(map[string]interface{})

	//Description
	// usage += c.Description + "\n\nUsage: " + c.Usage + "\n\n"
	fmt.Fprintf(wUsage, c.Description+"\n\nUsage: "+c.Usage+"\n\n")

	//flags
	fp := c.Flags
	for i := 0; i < len(fp); i++ {
		flag := fp[i]
		// prepation checks
		if len(flag.Name) == 0 {
			fmt.Println("Error: empty flag name defintion")
			os.Exit(2)
		}
		//check Default => if no value provided assume it is a bool flag
		if flag.Default == nil {
			flag.Default = false
		}

		switch flag.Default.(type) {
		case int:
			createIntFlag(config, flag, &shorts, wUsage)
		case string:
			createStringFlag(config, flag, &shorts, wUsage)
		case bool:
			createBoolFlag(config, flag, &shorts, wUsage)
		case float64:
			createFloatFlag(config, flag, &shorts, wUsage)
			//todo: add float64;multiple value
		default:
			fmt.Println("Unknown flag type:", flag.Default)
			os.Exit(2)
		}
	}
	fmt.Fprintf(wUsage, "\nUse \""+os.Args[0]+" --help\" for more information about the command.\n")

	//cheat sheet pt1
	var cheatSheet bool
	if len(c.CheatSheet) > 0 {
		fmt.Fprintf(wUsage, "\nSee command examples with \""+os.Args[0]+" --cheat-sheet\"\n")
		flag.BoolVar(&cheatSheet, "cheat-sheet", false, "print cheat sheet")
		flag.BoolVar(&cheatSheet, "cs", false, "print cheat sheet")
	}

	wUsage.Flush()
	flag.Usage = func() { fmt.Print(usage.String()) }
	flag.Parse()

	//cheat sheet pt2
	if len(c.CheatSheet) > 0 && cheatSheet {
		c.PrintCheatSheet()
		os.Exit(0)
	}

	return config
}

//PrintCheatSheet: print the cheat sheet of the command
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

func createIntFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer) {
	name := f.Name
	shortName := name[0:1]
	var intPtr int
	flag.IntVar(&intPtr, name, int(reflect.ValueOf(f.Default).Int()), f.Description)
	if !stringSlice.Contains(*shorts, shortName) {
		flag.IntVar(&intPtr, shortName, int(reflect.ValueOf(f.Default).Int()), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName))
		cfg[shortName] = &intPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, ""))
	}
	cfg[name] = &intPtr
}

func createStringFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer) {
	name := f.Name
	shortName := name[0:1]
	var strPtr string
	flag.StringVar(&strPtr, name, string(reflect.ValueOf(f.Default).String()), f.Description)
	if !stringSlice.Contains(*shorts, shortName) {
		flag.StringVar(&strPtr, shortName, string(reflect.ValueOf(f.Default).String()), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName))
		cfg[shortName] = &strPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, ""))
	}
	cfg[name] = &strPtr
}

func createBoolFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer) {
	name := f.Name
	shortName := name[0:1]
	var bPtr bool
	flag.BoolVar(&bPtr, name, bool(reflect.ValueOf(f.Default).Bool()), f.Description)
	cfg[name] = &bPtr
	if !stringSlice.Contains(*shorts, shortName) {
		flag.BoolVar(&bPtr, shortName, bool(reflect.ValueOf(f.Default).Bool()), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName))
		cfg[shortName] = &bPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, ""))
	}
	cfg[name] = &bPtr
}

func createFloatFlag(cfg Config, f Flag, shorts *[]string, wUsage *tabwriter.Writer) {
	name := f.Name
	shortName := name[0:1]
	var floatPtr float64
	flag.Float64Var(&floatPtr, name, float64(reflect.ValueOf(f.Default).Float()), f.Description)
	cfg[name] = &floatPtr
	if !stringSlice.Contains(*shorts, shortName) {
		flag.Float64Var(&floatPtr, shortName, float64(reflect.ValueOf(f.Default).Float()), f.Description)
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, shortName))
		cfg[shortName] = &floatPtr
		*shorts = append(*shorts, shortName)
	} else {
		fmt.Fprintf(wUsage, getFlagLine(f.Description, f.Default, name, ""))
	}
	cfg[name] = &floatPtr
}

//getFlagLine: return the string representing the flag line in help message. If short is empty, only long will be include in string
func getFlagLine(description string, defaultValue interface{}, long string, short string) (line string) {
	defaultValueStr := ". (default: "
	switch defaultValue.(type) {
	case int:
		defaultValueStr += strconv.Itoa(int(reflect.ValueOf(defaultValue).Int())) + ")\n"
	case string:
		defaultValueStr += "\"" + string(reflect.ValueOf(defaultValue).String()) + "\")\n"
	case bool:
		defaultValueStr += strconv.FormatBool(reflect.ValueOf(defaultValue).Bool()) + ")\n"
	case float64:
		defaultValueStr += strconv.FormatFloat(float64(reflect.ValueOf(defaultValue).Float()), 'f', -1, 64) + ")\n"
	default:
		fmt.Println("Unknown type for default value:", defaultValue)
		os.Exit(1)
	}

	if short == "" {
		line = "--" + long + "\t\t\t" + description + defaultValueStr
	} else {
		line = "--" + long + "\t-" + short + "\t\t" + description + defaultValueStr
	}
	return line
}
