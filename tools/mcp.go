package tools

import (
	"agent.article.fp/config"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type mcpTool struct {
	definition ToolDefinition
	client     *client.Client
}

// NewMCP connects, negotiates and discovers tools during startup. A caller can
// therefore treat a failed MCP connection as a startup failure, not a delayed
// model-facing error.
func NewMCP(ctx context.Context, cfg config.MCPConfig) ([]Tool, *MCPClient, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, nil, fmt.Errorf("MCP %q endpoint is required", cfg.Name)
	}
	cli, err := client.NewStreamableHttpClient(cfg.Endpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("create MCP client %q: %w", cfg.Name, err)
	}
	if _, err := cli.Initialize(ctx, mcp.InitializeRequest{Params: mcp.InitializeParams{ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION, Capabilities: mcp.ClientCapabilities{}, ClientInfo: mcp.Implementation{Name: "z-agent", Version: "1.0"}}}); err != nil {
		cli.Close()
		return nil, nil, fmt.Errorf("initialize MCP %q: %w", cfg.Name, err)
	}
	listed, err := cli.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		cli.Close()
		return nil, nil, fmt.Errorf("list MCP tools for %q: %w", cfg.Name, err)
	}
	tools := make([]Tool, 0, len(listed.Tools))
	for _, remote := range listed.Tools {
		parameters := map[string]any{"type": "object"}
		if len(remote.RawInputSchema) > 0 {
			if err := json.Unmarshal(remote.RawInputSchema, &parameters); err != nil {
				cli.Close()
				return nil, nil, fmt.Errorf("decode MCP tool schema %q: %w", remote.Name, err)
			}
		}
		tools = append(tools, &mcpTool{definition: ToolDefinition{Name: remote.Name, Description: remote.Description, Parameters: parameters}, client: cli})
	}
	return tools, &MCPClient{client: cli}, nil
}

type MCPClient struct{ client *client.Client }

func (c *MCPClient) Close() error { return c.client.Close() }

func (t *mcpTool) Definition() ToolDefinition { return t.definition }
func (*mcpTool) Safety() Safety {
	return Safety{SideEffect: "network", Permission: "allow", Reason: "calls configured MCP service"}
}
func (t *mcpTool) Run(ctx context.Context, raw json.RawMessage) Result {
	var args map[string]any
	if len(raw) > 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &args); err != nil {
			return Result{Content: "invalid tool arguments: " + err.Error(), IsError: true}
		}
	}
	result, err := t.client.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: t.definition.Name, Arguments: args}})
	if err != nil {
		return Result{Content: err.Error(), IsError: true}
	}
	var content strings.Builder
	for _, item := range result.Content {
		if text, ok := item.(mcp.TextContent); ok {
			content.WriteString(text.Text)
			content.WriteByte('\n')
		} else {
			encoded, _ := json.Marshal(item)
			content.Write(encoded)
			content.WriteByte('\n')
		}
	}
	if content.Len() == 0 && result.StructuredContent != nil {
		encoded, _ := json.Marshal(result.StructuredContent)
		content.Write(encoded)
	}
	return Result{Content: strings.TrimSpace(content.String()), IsError: result.IsError}
}
