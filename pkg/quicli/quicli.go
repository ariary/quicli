package quicli

import (
	"flag"
	"fmt"
	"os"
	"reflect"

	stringSlice "github.com/ariary/go-utils/pkg/stringSlice"
)

//struct representing a cli flag
type Flag struct {
	Name        string
	Description string
	// UseDefault  bool
	Default interface{}
}

type Flags []Flag

type Config map[string]interface{}

// struct representing CLI
type Cli struct {
	Usage       string
	Description string
	Flags       []Flag
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

//Parse: parse the different flags and return the struct containing the flag values
func (c *Cli) Parse() (config Config) {
	var usage string
	var shorts []string
	config = make(map[string]interface{})
	//Description
	usage += "Usage: " + c.Usage + "\nDescription: " + c.Description + "\n\n"

	//flags
	fp := c.Flags
	for i := 0; i < len(fp); i++ {
		name := fp[i].Name
		if len(name) == 0 {
			fmt.Println("Error: empty flag name defintion")
			os.Exit(1)
		}

		switch fp[i].Default.(type) {
		case int:
			createIntFlag(config, fp[i], &shorts, &usage)
		case string:
			createStringFlag(config, fp[i], &shorts, &usage)
		case bool:
			createBoolFlag(config, fp[i], &shorts, &usage)
			//todo: add float64;multiple value
		default:
			fmt.Println("Unknown flag type:", fp[i].Default)
			os.Exit(1)
		}
	}
	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()
	return config
}

func createIntFlag(cfg Config, f Flag, shorts *[]string, usage *string) {
	name := f.Name
	shortName := name[0:1]
	var intPtr int
	flag.IntVar(&intPtr, name, int(reflect.ValueOf(f.Default).Int()), f.Description)
	if !stringSlice.Contains(*shorts, shortName) {
		flag.IntVar(&intPtr, shortName, int(reflect.ValueOf(f.Default).Int()), f.Description)
		*usage += "--" + name + "\t-" + shortName + "\t" + f.Description + "\n"
		cfg[shortName] = &intPtr
		*shorts = append(*shorts, shortName)
	} else {
		*usage += "--" + name + "\t\t" + f.Description + "\n"
	}
	cfg[name] = &intPtr
}

func createStringFlag(cfg Config, f Flag, shorts *[]string, usage *string) {
	name := f.Name
	shortName := name[0:1]
	var strPtr string
	flag.StringVar(&strPtr, name, string(reflect.ValueOf(f.Default).String()), f.Description)
	if !stringSlice.Contains(*shorts, shortName) {
		flag.StringVar(&strPtr, shortName, string(reflect.ValueOf(f.Default).String()), f.Description)
		*usage += "--" + name + "\t-" + shortName + "\t" + f.Description + "\n"
		cfg[shortName] = &strPtr
		*shorts = append(*shorts, shortName)
	} else {
		*usage += "--" + name + "\t\t" + f.Description + "\n"
	}
	cfg[name] = &strPtr
}

func createBoolFlag(cfg Config, f Flag, shorts *[]string, usage *string) {
	name := f.Name
	shortName := name[0:1]
	var bPtr bool
	flag.BoolVar(&bPtr, name, bool(reflect.ValueOf(f.Default).Bool()), f.Description)
	cfg[name] = &bPtr
	if !stringSlice.Contains(*shorts, shortName) {
		flag.BoolVar(&bPtr, shortName, bool(reflect.ValueOf(f.Default).Bool()), f.Description)
		*usage += "--" + name + "\t-" + shortName + "\t" + f.Description + "\n"
		cfg[shortName] = &bPtr
		*shorts = append(*shorts, shortName)
	} else {
		*usage += "--" + name + "\t\t" + f.Description + "\n"
	}
	cfg[name] = &bPtr
}
