// Package tools 实现 Tool、Registry、MCP/skill/exec/search 实现
package tools

import (
	"context"
	"encoding/json"
)

type ToolDefinition struct{}

type ToolSafety struct{}

type ToolResult struct{}

type Tool interface {
	Definition() ToolDefinition
	Safety() ToolSafety
	Run(ctx context.Context, args json.RawMessage) ToolResult
}

func New() ([]Tool, error) {
	return nil, nil
}
