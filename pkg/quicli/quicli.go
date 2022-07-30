package quicli

import (
	"flag"
	"fmt"
	"os"
	"reflect"

	stringSlice "github.com/ariary/go-utils/pkg/stringSlice"
)

type Flag struct {
	Name        string
	Description string
	// UseDefault  bool
	Default interface{}
}

type Flags []Flag

type Config map[string]interface{}

type Cli struct {
	Name        string
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

func (c *Cli) Parse() (config Config) {
	var usage string
	var shorts []string
	config = make(map[string]interface{})
	//Description
	usage += c.Name + ": " + c.Description

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
			createIntFlag(config, fp[i], &shorts)
		case string:
			createStringFlag(config, fp[i], &shorts)
		case bool:
			createBoolFlag(config, fp[i], &shorts)
			//todo: add float
		}
		// if len(name) == 0 {
		// 	return config, errors.New("empty flag name")
		// }
		// shortFlag := name[0:1]
		// if stringSlice.Contains(shorts, shortFlag) {
		// 	createLongFlagOnly(config, &usage, fp[i])
		// } else {
		// 	createFlag(config, &usage, fp[i], shortFlag)
		// }
		// fmt.Println(fp[i].Name, ",", reflect.TypeOf(fp[i].Default))
		// flag.BoolVar(&cfg.Follow, "location", false, "Follow redirections")
		// flag.BoolVar(&cfg.Follow, "L", false, "Follow redirections")

	}
	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()
	return config
}

func createIntFlag(cfg Config, f Flag, shorts *[]string) {
	name := f.Name
	shortName := name[0:1]
	var intPtr int
	flag.IntVar(&intPtr, name, int(reflect.ValueOf(f.Default).Int()), f.Description)
	if !stringSlice.Contains(*shorts, shortName) {
		flag.IntVar(&intPtr, shortName, int(reflect.ValueOf(f.Default).Int()), f.Description)
		cfg[shortName] = &intPtr
		*shorts = append(*shorts, shortName)
	}
	cfg[name] = &intPtr
}

func createStringFlag(cfg Config, f Flag, shorts *[]string) {
	name := f.Name
	shortName := name[0:1]
	var strPtr string
	flag.StringVar(&strPtr, name, string(reflect.ValueOf(f.Default).String()), f.Description)
	if !stringSlice.Contains(*shorts, shortName) {
		flag.StringVar(&strPtr, shortName, string(reflect.ValueOf(f.Default).String()), f.Description)
		cfg[shortName] = &strPtr
		*shorts = append(*shorts, shortName)
	}
	cfg[name] = &strPtr
}

func createBoolFlag(cfg Config, f Flag, shorts *[]string) {
	name := f.Name
	shortName := name[0:1]
	var bPtr bool
	flag.BoolVar(&bPtr, name, bool(reflect.ValueOf(f.Default).Bool()), f.Description)
	cfg[name] = &bPtr
	if !stringSlice.Contains(*shorts, shortName) {
		flag.BoolVar(&bPtr, shortName, bool(reflect.ValueOf(f.Default).Bool()), f.Description)
		cfg[shortName] = &bPtr
		*shorts = append(*shorts, shortName)
	}
	cfg[name] = &bPtr
}
