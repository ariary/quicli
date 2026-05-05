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

// --- Task 9: Edge Case Tests ---

func TestArgsToFlagsEmptyArgs(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
	}
	args := map[string]any{}
	cliArgs, envVars := argsToFlags(args, schema)
	if len(cliArgs) != 0 {
		t.Errorf("expected no cliArgs, got %v", cliArgs)
	}
	if len(envVars) != 0 {
		t.Errorf("expected no envVars, got %v", envVars)
	}
}

func TestArgsToFlagsNilProperties(t *testing.T) {
	schema := map[string]any{
		"type": "object",
	}
	args := map[string]any{"foo": "bar"}
	cliArgs, envVars := argsToFlags(args, schema)
	if cliArgs != nil {
		t.Errorf("expected nil cliArgs for nil properties, got %v", cliArgs)
	}
	if len(envVars) != 0 {
		t.Errorf("expected no envVars, got %v", envVars)
	}
}

func TestArgsToFlagsUnknownArgIgnored(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
	}
	args := map[string]any{
		"name":    "hello",
		"unknown": "ignored",
	}
	cliArgs, _ := argsToFlags(args, schema)
	for _, a := range cliArgs {
		if a == "--unknown" || a == "ignored" {
			t.Errorf("unknown arg should be ignored, got %v", cliArgs)
		}
	}
	if len(cliArgs) != 2 {
		t.Errorf("expected 2 cliArgs (--name hello), got %v", cliArgs)
	}
}

func TestArgsToFlagsFloatNonInteger(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"rate": map[string]any{"type": "number"},
		},
	}
	args := map[string]any{"rate": float64(3.14)}
	cliArgs, _ := argsToFlags(args, schema)
	found := false
	for i, a := range cliArgs {
		if a == "--rate" && i+1 < len(cliArgs) {
			if cliArgs[i+1] != "3.14" {
				t.Errorf("rate value: got %q, want 3.14", cliArgs[i+1])
			}
			found = true
		}
	}
	if !found {
		t.Error("--rate not found in cliArgs")
	}
}

func TestArgsToFlagsEmptyArray(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"tags": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
	}
	args := map[string]any{"tags": []any{}}
	cliArgs, _ := argsToFlags(args, schema)
	if len(cliArgs) != 0 {
		t.Errorf("empty array should produce no args, got %v", cliArgs)
	}
}

func TestArgsToFlagsEmptyString(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
	}
	args := map[string]any{"name": ""}
	cliArgs, _ := argsToFlags(args, schema)
	if len(cliArgs) != 2 || cliArgs[0] != "--name" || cliArgs[1] != "" {
		t.Errorf("empty string should produce --name '', got %v", cliArgs)
	}
}

func TestArgsToFlagsEnvOnlyWithoutEnvVar(t *testing.T) {
	// env-only field without x-quicli-env-var should be silently ignored
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"secret": map[string]any{"type": "string", "x-quicli-input": "env-only"},
		},
	}
	args := map[string]any{"secret": "value"}
	cliArgs, envVars := argsToFlags(args, schema)
	if len(cliArgs) != 0 {
		t.Errorf("env-only should not produce cliArgs, got %v", cliArgs)
	}
	if len(envVars) != 0 {
		t.Errorf("missing env-var key should not produce envVars, got %v", envVars)
	}
}

func TestParseSchemaSubcommandsSorted(t *testing.T) {
	raw := `{
		"type": "object",
		"x-quicli-subcommands": {
			"zeta": {"type": "object", "description": "z"},
			"alpha": {"type": "object", "description": "a"},
			"middle": {"type": "object", "description": "m"}
		}
	}`
	tools, err := parseSchema([]byte(raw), "app")
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(tools))
	}
	expectedOrder := []string{"app_alpha", "app_middle", "app_zeta"}
	for i, tool := range tools {
		if tool.Name != expectedOrder[i] {
			t.Errorf("tool[%d]: got %q, want %q", i, tool.Name, expectedOrder[i])
		}
	}
}

func TestParseSchemaNoDescription(t *testing.T) {
	raw := `{"type": "object"}`
	tools, err := parseSchema([]byte(raw), "bare")
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
	if tools[0].Description != "" {
		t.Errorf("expected empty description, got %q", tools[0].Description)
	}
}

func TestParseSchemaNoProperties(t *testing.T) {
	raw := `{"type": "object", "description": "minimal"}`
	tools, err := parseSchema([]byte(raw), "min")
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
	// inputSchema should still have "type": "object" but no "properties" key
	if tools[0].InputSchema["type"] != "object" {
		t.Error("expected type object in inputSchema")
	}
	if _, has := tools[0].InputSchema["properties"]; has {
		t.Error("expected no properties key in inputSchema")
	}
}

func TestParseSchemaPreservesRequired(t *testing.T) {
	raw := `{
		"type": "object",
		"description": "tool",
		"properties": {"name": {"type": "string"}},
		"required": ["name"]
	}`
	tools, err := parseSchema([]byte(raw), "t")
	if err != nil {
		t.Fatal(err)
	}
	req, ok := tools[0].InputSchema["required"].([]any)
	if !ok {
		t.Fatal("required field missing or wrong type")
	}
	if len(req) != 1 || req[0] != "name" {
		t.Errorf("expected required [name], got %v", req)
	}
}

func TestHandleInitializeProtocolVersion(t *testing.T) {
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`"abc"`),
		Method:  "initialize",
	}
	resp := handleInitialize(req)
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("result is not a map")
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("protocol version: got %v", result["protocolVersion"])
	}
	info, ok := result["serverInfo"].(map[string]any)
	if !ok {
		t.Fatal("serverInfo is not a map")
	}
	if info["name"] != "quicli-mcp" {
		t.Errorf("server name: got %v", info["name"])
	}
	if info["version"] != "0.1.0" {
		t.Errorf("server version: got %v", info["version"])
	}
	// ID should be preserved (string ID)
	if string(resp.ID) != `"abc"` {
		t.Errorf("ID: got %s, want \"abc\"", resp.ID)
	}
}

func TestHandleToolsCallUnknownTool(t *testing.T) {
	tools := []mcpTool{
		{Name: "app_scan", Description: "scan"},
	}
	fullSchema := map[string]any{}
	params, _ := json.Marshal(map[string]any{
		"name":      "app_nonexistent",
		"arguments": map[string]any{},
	})
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`3`),
		Method:  "tools/call",
		Params:  params,
	}
	resp := handleToolsCall(req, "/usr/bin/app", tools, fullSchema)
	if resp.Error == nil {
		t.Fatal("expected error for unknown tool")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("error code: got %d, want -32602", resp.Error.Code)
	}
}

func TestHandleToolsCallInvalidParams(t *testing.T) {
	tools := []mcpTool{}
	fullSchema := map[string]any{}
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`4`),
		Method:  "tools/call",
		Params:  json.RawMessage(`not valid json`),
	}
	resp := handleToolsCall(req, "/usr/bin/app", tools, fullSchema)
	if resp.Error == nil {
		t.Fatal("expected error for invalid params")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("error code: got %d, want -32602", resp.Error.Code)
	}
}

func TestHandleToolsCallExecutesBinary(t *testing.T) {
	// Use echo as a real binary to test actual execution
	tools := []mcpTool{
		{Name: "echo", Description: "echo", InputSchema: map[string]any{"type": "object"}},
	}
	fullSchema := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
	params, _ := json.Marshal(map[string]any{
		"name":      "echo",
		"arguments": map[string]any{},
	})
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`5`),
		Method:  "tools/call",
		Params:  params,
	}
	resp := handleToolsCall(req, "/bin/echo", tools, fullSchema)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("result is not a map")
	}
	// isError should not be set for successful execution
	if _, has := result["isError"]; has {
		t.Error("isError should not be set for successful command")
	}
}

func TestHandleToolsCallNonZeroExit(t *testing.T) {
	// Use 'false' command which always exits with code 1
	tools := []mcpTool{
		{Name: "false", Description: "always fails", InputSchema: map[string]any{"type": "object"}},
	}
	fullSchema := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
	params, _ := json.Marshal(map[string]any{
		"name":      "false",
		"arguments": map[string]any{},
	})
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`6`),
		Method:  "tools/call",
		Params:  params,
	}
	resp := handleToolsCall(req, "/usr/bin/false", tools, fullSchema)
	if resp.Error != nil {
		t.Fatalf("unexpected RPC error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("result is not a map")
	}
	isErr, has := result["isError"]
	if !has {
		t.Fatal("isError should be set for failed command")
	}
	if isErr != true {
		t.Errorf("isError: got %v, want true", isErr)
	}
}

func TestWriteResponseJSON(t *testing.T) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Result:  map[string]any{"ok": true},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed["jsonrpc"] != "2.0" {
		t.Errorf("jsonrpc: got %v", parsed["jsonrpc"])
	}
	// Error should be omitted when nil
	if _, has := parsed["error"]; has {
		t.Error("error field should be omitted when nil")
	}
}

func TestWriteResponseErrorOmitsResult(t *testing.T) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Error:   &rpcError{Code: -32601, Message: "not found"},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	if _, has := parsed["result"]; has {
		t.Error("result field should be omitted when nil")
	}
	errObj, ok := parsed["error"].(map[string]any)
	if !ok {
		t.Fatal("error is not an object")
	}
	if errObj["code"].(float64) != -32601 {
		t.Errorf("error code: got %v", errObj["code"])
	}
}

func TestArgsToFlagsSortOrder(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"zebra": map[string]any{"type": "string"},
			"alpha": map[string]any{"type": "string"},
			"mid":   map[string]any{"type": "string"},
		},
	}
	args := map[string]any{
		"zebra": "z",
		"alpha": "a",
		"mid":   "m",
	}
	cliArgs, _ := argsToFlags(args, schema)
	// Should be sorted: alpha, mid, zebra
	expected := []string{"--alpha", "a", "--mid", "m", "--zebra", "z"}
	if len(cliArgs) != len(expected) {
		t.Fatalf("expected %d args, got %d: %v", len(expected), len(cliArgs), cliArgs)
	}
	for i, e := range expected {
		if cliArgs[i] != e {
			t.Errorf("cliArgs[%d]: got %q, want %q", i, cliArgs[i], e)
		}
	}
}

func TestArgsToFlagsLargeInteger(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"big": map[string]any{"type": "integer"},
		},
	}
	args := map[string]any{"big": float64(1000000000)}
	cliArgs, _ := argsToFlags(args, schema)
	found := false
	for i, a := range cliArgs {
		if a == "--big" && i+1 < len(cliArgs) {
			if cliArgs[i+1] != "1000000000" {
				t.Errorf("big value: got %q, want 1000000000", cliArgs[i+1])
			}
			found = true
		}
	}
	if !found {
		t.Error("--big not found")
	}
}

func TestArgsToFlagsZeroInteger(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"count": map[string]any{"type": "integer"},
		},
	}
	args := map[string]any{"count": float64(0)}
	cliArgs, _ := argsToFlags(args, schema)
	found := false
	for i, a := range cliArgs {
		if a == "--count" && i+1 < len(cliArgs) {
			if cliArgs[i+1] != "0" {
				t.Errorf("count value: got %q, want 0", cliArgs[i+1])
			}
			found = true
		}
	}
	if !found {
		t.Error("--count not found")
	}
}

func TestArgsToFlagsNegativeInteger(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"offset": map[string]any{"type": "integer"},
		},
	}
	args := map[string]any{"offset": float64(-10)}
	cliArgs, _ := argsToFlags(args, schema)
	found := false
	for i, a := range cliArgs {
		if a == "--offset" && i+1 < len(cliArgs) {
			if cliArgs[i+1] != "-10" {
				t.Errorf("offset value: got %q, want -10", cliArgs[i+1])
			}
			found = true
		}
	}
	if !found {
		t.Error("--offset not found")
	}
}
