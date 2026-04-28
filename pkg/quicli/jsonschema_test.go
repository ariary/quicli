package quicli

import (
	"encoding/json"
	"testing"
	"time"
)

func TestJSONSchemaBasic(t *testing.T) {
	cli := Cli{
		Usage:       "prog [flags]",
		Description: "test tool",
		Flags: Flags{
			{Name: "count", Default: 1, Description: "how many"},
			{Name: "name", Default: "hello", Description: "a name"},
			{Name: "verbose", Default: false, Description: "verbose output"},
			{Name: "ratio", Default: float64(1.5), Description: "a ratio"},
		},
	}
	schema := cli.JSONSchema()
	if schema["type"] != "object" {
		t.Error("schema type should be object")
	}
	props := schema["properties"].(map[string]any)
	if len(props) != 4 {
		t.Errorf("expected 4 properties, got %d", len(props))
	}

	// Check type mappings
	countProp := props["count"].(map[string]any)
	if countProp["type"] != "integer" {
		t.Errorf("count type: got %v, want integer", countProp["type"])
	}
	nameProp := props["name"].(map[string]any)
	if nameProp["type"] != "string" {
		t.Errorf("name type: got %v, want string", nameProp["type"])
	}
	verboseProp := props["verbose"].(map[string]any)
	if verboseProp["type"] != "boolean" {
		t.Errorf("verbose type: got %v, want boolean", verboseProp["type"])
	}
	ratioProp := props["ratio"].(map[string]any)
	if ratioProp["type"] != "number" {
		t.Errorf("ratio type: got %v, want number", ratioProp["type"])
	}
}

func TestJSONSchemaRequired(t *testing.T) {
	cli := Cli{
		Flags: Flags{
			{Name: "output", Default: "", Description: "output file", Required: true},
			{Name: "verbose", Default: false, Description: "verbose"},
		},
	}
	schema := cli.JSONSchema()
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("schema should have required array")
	}
	if len(required) != 1 || required[0] != "output" {
		t.Errorf("required: got %v, want [output]", required)
	}
}

func TestJSONSchemaChoices(t *testing.T) {
	cli := Cli{
		Flags: Flags{
			{Name: "format", Default: "json", Description: "output format", Choices: []string{"json", "yaml", "csv"}},
		},
	}
	schema := cli.JSONSchema()
	props := schema["properties"].(map[string]any)
	fmtProp := props["format"].(map[string]any)
	enum, ok := fmtProp["enum"].([]string)
	if !ok {
		t.Fatal("format should have enum")
	}
	if len(enum) != 3 {
		t.Errorf("enum length: got %d, want 3", len(enum))
	}
}

func TestJSONSchemaDuration(t *testing.T) {
	cli := Cli{
		Flags: Flags{
			{Name: "timeout", Default: 30 * time.Second, Description: "request timeout"},
		},
	}
	schema := cli.JSONSchema()
	props := schema["properties"].(map[string]any)
	prop := props["timeout"].(map[string]any)
	if prop["type"] != "string" {
		t.Errorf("duration type: got %v, want string", prop["type"])
	}
	if prop["format"] != "duration" {
		t.Errorf("duration format: got %v, want duration", prop["format"])
	}
	if prop["default"] != "30s" {
		t.Errorf("duration default: got %v, want 30s", prop["default"])
	}
}

func TestJSONSchemaString(t *testing.T) {
	cli := Cli{
		Description: "my tool",
		Flags: Flags{
			{Name: "name", Default: "world", Description: "who"},
		},
	}
	s := cli.JSONSchemaString()
	var parsed map[string]any
	if err := json.Unmarshal([]byte(s), &parsed); err != nil {
		t.Fatalf("JSONSchemaString not valid JSON: %v", err)
	}
	if parsed["description"] != "my tool" {
		t.Error("description missing from JSON output")
	}
}

func TestSubcommandSchemas(t *testing.T) {
	cli := Cli{
		Description: "test",
		Flags: Flags{
			{Name: "verbose", Default: false, Description: "verbose", SharedSubcommand: SubcommandSet{"build"}},
		},
		Subcommands: Subcommands{
			{
				Name:        "build",
				Description: "build stuff",
				Function:    func(Config) {},
				Flags:       Flags{{Name: "output", Default: "", Description: "output dir"}},
			},
		},
	}
	schemas := cli.SubcommandSchemas()
	buildSchema, ok := schemas["build"]
	if !ok {
		t.Fatal("missing build schema")
	}
	props := buildSchema["properties"].(map[string]any)
	if _, ok := props["verbose"]; !ok {
		t.Error("build schema should include shared flag 'verbose'")
	}
	if _, ok := props["output"]; !ok {
		t.Error("build schema should include exclusive flag 'output'")
	}
}

func TestJSONSchemaFlagValue(t *testing.T) {
	lv := &testLevel{val: "info"}
	cli := Cli{
		Flags: Flags{
			{Name: "level", Default: lv, Description: "log level"},
		},
	}
	schema := cli.JSONSchema()
	props := schema["properties"].(map[string]any)
	prop := props["level"].(map[string]any)
	if prop["type"] != "string" {
		t.Errorf("flag.Value type: got %v, want string", prop["type"])
	}
	if prop["default"] != "info" {
		t.Errorf("flag.Value default: got %v, want info", prop["default"])
	}
}

func TestJSONSchemaSubcommandsIncluded(t *testing.T) {
	cli := Cli{
		Description: "my tool",
		Subcommands: Subcommands{
			{Name: "build", Description: "build it", Function: func(Config) {},
				Flags: Flags{{Name: "output", Default: "", Description: "output dir"}}},
		},
	}
	s := cli.JSONSchemaString()
	var parsed map[string]any
	if err := json.Unmarshal([]byte(s), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	subs, ok := parsed["x-quicli-subcommands"]
	if !ok {
		t.Fatal("missing x-quicli-subcommands in output")
	}
	subsMap := subs.(map[string]any)
	if _, ok := subsMap["build"]; !ok {
		t.Error("missing build subcommand in schema output")
	}
}

func TestJSONSchemaSlice(t *testing.T) {
	cli := Cli{
		Flags: Flags{
			{Name: "files", Default: []string{}, Description: "input files"},
		},
	}
	schema := cli.JSONSchema()
	props := schema["properties"].(map[string]any)
	prop := props["files"].(map[string]any)
	if prop["type"] != "array" {
		t.Errorf("slice type: got %v, want array", prop["type"])
	}
}
