package main

import (
	"encoding/json"
	"testing"
)

func TestParseSchemaRootOnly(t *testing.T) {
	raw := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"description": "A test tool",
		"properties": {
			"count": {"type": "integer", "description": "how many", "default": 1},
			"name": {"type": "string", "description": "a name"}
		},
		"required": ["name"]
	}`
	tools, err := parseSchema([]byte(raw), "mytool")
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
	if tools[0].Name != "mytool" {
		t.Errorf("name: got %q, want mytool", tools[0].Name)
	}
	if tools[0].Description != "A test tool" {
		t.Errorf("description: got %q", tools[0].Description)
	}
}

func TestParseSchemaWithSubcommands(t *testing.T) {
	raw := `{
		"type": "object",
		"description": "A tool",
		"x-quicli-subcommands": {
			"scan": {
				"type": "object",
				"description": "scan a target",
				"properties": {"target": {"type": "string"}},
				"required": ["target"]
			},
			"list": {
				"type": "object",
				"description": "list results"
			}
		}
	}`
	tools, err := parseSchema([]byte(raw), "scanner")
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
	names := map[string]bool{}
	for _, tool := range tools {
		names[tool.Name] = true
	}
	if !names["scanner_scan"] {
		t.Error("missing scanner_scan")
	}
	if !names["scanner_list"] {
		t.Error("missing scanner_list")
	}
}

func TestArgsToFlags(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"count":   map[string]any{"type": "integer"},
			"name":    map[string]any{"type": "string"},
			"verbose": map[string]any{"type": "boolean"},
			"tags":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"secret":  map[string]any{"type": "string", "x-quicli-input": "env-only", "x-quicli-env-var": "MY_SECRET"},
		},
	}
	args := map[string]any{
		"count":   float64(5),
		"name":    "hello",
		"verbose": true,
		"tags":    []any{"a", "b"},
		"secret":  "token123",
	}
	cliArgs, envVars := argsToFlags(args, schema)

	// Check env-only goes to envVars
	if envVars["MY_SECRET"] != "token123" {
		t.Errorf("env var: got %q, want token123", envVars["MY_SECRET"])
	}

	// Check CLI args contain expected flags (order is sorted by key)
	// count, name, tags, verbose (alphabetical, secret excluded)
	expectedContains := map[string]bool{"--count": true, "5": true, "--name": true, "hello": true, "--verbose": true, "--tags": true, "a": true, "b": true}
	for _, a := range cliArgs {
		delete(expectedContains, a)
	}
	if len(expectedContains) > 0 {
		t.Errorf("missing in cliArgs: %v, got: %v", expectedContains, cliArgs)
	}

	// secret should NOT be in cliArgs
	for _, a := range cliArgs {
		if a == "--secret" || a == "token123" {
			t.Errorf("env-only flag should not be in cliArgs: %v", cliArgs)
		}
	}
}

func TestArgsToFlagsIntegerFormatting(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"port": map[string]any{"type": "integer"},
		},
	}
	args := map[string]any{"port": float64(8080)}
	cliArgs, _ := argsToFlags(args, schema)
	found := false
	for i, a := range cliArgs {
		if a == "--port" && i+1 < len(cliArgs) {
			if cliArgs[i+1] != "8080" {
				t.Errorf("port value: got %q, want 8080", cliArgs[i+1])
			}
			found = true
		}
	}
	if !found {
		t.Error("--port not found")
	}
}

func TestArgsToFlagsBoolFalseOmitted(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"verbose": map[string]any{"type": "boolean"},
		},
	}
	args := map[string]any{"verbose": false}
	cliArgs, _ := argsToFlags(args, schema)
	if len(cliArgs) != 0 {
		t.Errorf("false boolean should produce no args, got %v", cliArgs)
	}
}

func TestHandleInitialize(t *testing.T) {
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
	}
	resp := handleInitialize(req)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

func TestHandleToolsList(t *testing.T) {
	tools := []mcpTool{
		{Name: "test", Description: "a test", InputSchema: map[string]any{"type": "object"}},
	}
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`2`),
		Method:  "tools/list",
	}
	resp := handleToolsList(req, tools)
	if resp.Error != nil {
		t.Fatal("unexpected error")
	}
}

func TestParseSchemaInvalidJSON(t *testing.T) {
	_, err := parseSchema([]byte("not json"), "test")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
