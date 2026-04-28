package quicli

import (
	"encoding/json"
	"flag"
	"time"
)

// JSONSchema returns the JSON Schema for the CLI's root flags.
func (c *Cli) JSONSchema() map[string]any {
	return buildSchema(c.Description, c.Flags)
}

// SubcommandSchemas returns a map of subcommand name to its JSON Schema.
// Each schema includes shared flags (from Cli.Flags) and exclusive flags.
func (c *Cli) SubcommandSchemas() map[string]map[string]any {
	schemas := make(map[string]map[string]any)
	for _, sub := range c.Subcommands {
		var flags []Flag
		// Include shared flags targeting this subcommand.
		for _, f := range c.Flags {
			if f.isForSubcommand(sub.Name) {
				flags = append(flags, f)
			}
		}
		// Include exclusive flags.
		flags = append(flags, sub.Flags...)
		schemas[sub.Name] = buildSchema(sub.Description, flags)
	}
	return schemas
}

// JSONSchemaString returns the root JSON Schema as pretty-printed JSON.
func (c *Cli) JSONSchemaString() string {
	schema := c.JSONSchema()

	// If there are subcommands, embed their schemas.
	if len(c.Subcommands) > 0 {
		subs := c.SubcommandSchemas()
		schema["x-quicli-subcommands"] = subs
	}

	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}

func buildSchema(description string, flags []Flag) map[string]any {
	schema := map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type":    "object",
	}
	if description != "" {
		schema["description"] = description
	}

	properties := map[string]any{}
	var required []string

	for _, f := range flags {
		properties[f.Name] = flagToSchemaProperty(f)
		if f.Required {
			required = append(required, f.Name)
		}
	}

	if len(properties) > 0 {
		schema["properties"] = properties
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

// jsonMarshalIndent is a thin wrapper around json.MarshalIndent.
func jsonMarshalIndent(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func flagToSchemaProperty(f Flag) map[string]any {
	prop := map[string]any{}

	if f.Description != "" {
		prop["description"] = f.Description
	}

	// Type mapping.
	switch d := f.Default.(type) {
	case int:
		prop["type"] = "integer"
		if d != 0 {
			prop["default"] = d
		}
	case string:
		prop["type"] = "string"
		if d != "" {
			prop["default"] = d
		}
	case bool:
		prop["type"] = "boolean"
		if d {
			prop["default"] = d
		}
	case float64:
		prop["type"] = "number"
		if d != 0 {
			prop["default"] = d
		}
	case []string:
		prop["type"] = "array"
		prop["items"] = map[string]any{"type": "string"}
	case time.Duration:
		prop["type"] = "string"
		prop["format"] = "duration"
		if d != 0 {
			prop["default"] = d.String()
		}
	default:
		if fv, ok := f.Default.(flag.Value); ok {
			prop["type"] = "string"
			s := fv.String()
			if s != "" {
				prop["default"] = s
			}
		}
	}

	// Enum from choices.
	if len(f.Choices) > 0 {
		prop["enum"] = f.Choices
	}

	// Env var metadata.
	ev := flagEnvVarDisplay(f)
	if ev != "" {
		prop["x-quicli-env-var"] = ev
	}

	return prop
}
