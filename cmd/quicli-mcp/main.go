package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// parseSchema parses a JSON Schema produced by quicli's --json-schema flag
// and returns a list of MCP tools. If x-quicli-subcommands exists, each
// subcommand becomes a separate tool named {binaryName}_{subcommand}.
// Otherwise the root schema becomes a single tool named {binaryName}.
func parseSchema(data []byte, binaryName string) ([]mcpTool, error) {
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("invalid JSON schema: %w", err)
	}

	subcmds, hasSub := schema["x-quicli-subcommands"]
	if hasSub {
		subcmdMap, ok := subcmds.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("x-quicli-subcommands is not an object")
		}

		// Sort subcommand names for deterministic output
		names := make([]string, 0, len(subcmdMap))
		for name := range subcmdMap {
			names = append(names, name)
		}
		sort.Strings(names)

		tools := make([]mcpTool, 0, len(names))
		for _, name := range names {
			sub := subcmdMap[name]
			subSchema, ok := sub.(map[string]any)
			if !ok {
				continue
			}
			desc, _ := subSchema["description"].(string)
			inputSchema := buildInputSchema(subSchema)
			tools = append(tools, mcpTool{
				Name:        binaryName + "_" + name,
				Description: desc,
				InputSchema: inputSchema,
			})
		}
		return tools, nil
	}

	// No subcommands — root becomes a single tool
	desc, _ := schema["description"].(string)
	inputSchema := buildInputSchema(schema)
	return []mcpTool{
		{
			Name:        binaryName,
			Description: desc,
			InputSchema: inputSchema,
		},
	}, nil
}

// buildInputSchema creates an MCP-compatible input schema from a quicli JSON Schema object.
func buildInputSchema(schema map[string]any) map[string]any {
	result := map[string]any{
		"type": "object",
	}
	if props, ok := schema["properties"]; ok {
		result["properties"] = props
	}
	if req, ok := schema["required"]; ok {
		result["required"] = req
	}
	return result
}

// argsToFlags converts MCP tool call arguments into CLI flags and environment
// variables. Fields marked with x-quicli-input: "env-only" are placed in
// envVars using their x-quicli-env-var key. Keys are sorted for deterministic output.
func argsToFlags(args map[string]any, schema map[string]any) (cliArgs []string, envVars map[string]string) {
	envVars = make(map[string]string)

	props, _ := schema["properties"].(map[string]any)
	if props == nil {
		return nil, envVars
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		val := args[key]
		propDef, ok := props[key]
		if !ok {
			continue
		}
		propMap, ok := propDef.(map[string]any)
		if !ok {
			continue
		}

		// Check if this is an env-only field
		if input, _ := propMap["x-quicli-input"].(string); input == "env-only" {
			envKey, _ := propMap["x-quicli-env-var"].(string)
			if envKey != "" {
				envVars[envKey] = fmt.Sprintf("%v", val)
			}
			continue
		}

		flag := "--" + key

		switch v := val.(type) {
		case bool:
			if v {
				cliArgs = append(cliArgs, flag)
			}
			// false → omit
		case float64:
			// Check if it's an integer value
			if v == float64(int64(v)) {
				cliArgs = append(cliArgs, flag, strconv.FormatInt(int64(v), 10))
			} else {
				cliArgs = append(cliArgs, flag, strconv.FormatFloat(v, 'f', -1, 64))
			}
		case []any:
			for _, item := range v {
				cliArgs = append(cliArgs, flag, fmt.Sprintf("%v", item))
			}
		case string:
			cliArgs = append(cliArgs, flag, v)
		default:
			cliArgs = append(cliArgs, flag, fmt.Sprintf("%v", v))
		}
	}

	return cliArgs, envVars
}

func handleInitialize(req jsonRPCRequest) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo": map[string]any{
				"name":    "quicli-mcp",
				"version": "0.1.0",
			},
		},
	}
}

func handleToolsList(req jsonRPCRequest, tools []mcpTool) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]any{"tools": tools},
	}
}

func handleToolsCall(req jsonRPCRequest, binaryPath string, tools []mcpTool, fullSchema map[string]any) jsonRPCResponse {
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32602, Message: "invalid params: " + err.Error()},
		}
	}

	// Find matching tool
	var matched *mcpTool
	for i := range tools {
		if tools[i].Name == params.Name {
			matched = &tools[i]
			break
		}
	}
	if matched == nil {
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32602, Message: "unknown tool: " + params.Name},
		}
	}

	// Determine subcommand and schema
	binaryName := filepath.Base(binaryPath)
	var subcommand string
	var toolSchema map[string]any

	prefix := binaryName + "_"
	if strings.HasPrefix(params.Name, prefix) {
		subcommand = strings.TrimPrefix(params.Name, prefix)
		// Get subcommand schema from fullSchema
		if subcmds, ok := fullSchema["x-quicli-subcommands"].(map[string]any); ok {
			if subSchema, ok := subcmds[subcommand].(map[string]any); ok {
				toolSchema = subSchema
			}
		}
	} else {
		// Root tool
		toolSchema = fullSchema
	}

	if toolSchema == nil {
		toolSchema = map[string]any{}
	}

	// Convert arguments to flags
	cliArgs, envVars := argsToFlags(params.Arguments, toolSchema)

	// Build command arguments
	var cmdArgs []string
	if subcommand != "" {
		cmdArgs = append(cmdArgs, subcommand)
	}
	cmdArgs = append(cmdArgs, cliArgs...)

	// Execute the command
	cmd := exec.Command(binaryPath, cmdArgs...)
	cmd.Env = os.Environ()
	for k, v := range envVars {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	output, err := cmd.CombinedOutput()

	result := map[string]any{
		"content": []mcpContent{
			{Type: "text", Text: string(output)},
		},
	}

	if err != nil {
		result["isError"] = true
	}

	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func writeResponse(resp jsonRPCResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling response: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: quicli-mcp <binary>")
		os.Exit(1)
	}

	binaryPath := os.Args[1]

	// Run <binary> --json-schema to discover the CLI contract
	cmd := exec.Command(binaryPath, "--json-schema")
	schemaOutput, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running %s --json-schema: %v\n", binaryPath, err)
		os.Exit(1)
	}

	binaryName := filepath.Base(binaryPath)

	tools, err := parseSchema(schemaOutput, binaryName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing schema: %v\n", err)
		os.Exit(1)
	}

	// Parse the full schema for use in tools/call handler
	var fullSchema map[string]any
	if err := json.Unmarshal(schemaOutput, &fullSchema); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing full schema: %v\n", err)
		os.Exit(1)
	}

	// Enter stdin loop — read one JSON object per line
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var req jsonRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			writeResponse(jsonRPCResponse{
				JSONRPC: "2.0",
				Error:   &rpcError{Code: -32700, Message: "parse error: " + err.Error()},
			})
			continue
		}

		switch req.Method {
		case "initialize":
			writeResponse(handleInitialize(req))
		case "notifications/initialized":
			// Silent — no response for notifications
		case "tools/list":
			writeResponse(handleToolsList(req, tools))
		case "tools/call":
			writeResponse(handleToolsCall(req, binaryPath, tools, fullSchema))
		default:
			writeResponse(jsonRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &rpcError{Code: -32601, Message: "method not found: " + req.Method},
			})
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
		os.Exit(1)
	}
}
