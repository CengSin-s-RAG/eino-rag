// Package tools contains model-visible tools and their execution registry.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
)

type Safety struct {
	SideEffect string `json:"side_effect"`
	Permission string `json:"permission"`
	Reason     string `json:"reason"`
}

type Result struct {
	Content string `json:"content"`
	IsError bool   `json:"is_error"`
}

type Tool interface {
	Definition() ToolDefinition
	Safety() Safety
	Run(ctx context.Context, args json.RawMessage) Result
}

type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type Registry struct {
	tools map[string]Tool
}

func NewRegistry(entries ...Tool) (*Registry, error) {
	r := &Registry{tools: make(map[string]Tool, len(entries))}
	for _, tool := range entries {
		if err := r.Register(tool); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *Registry) Register(tool Tool) error {
	if tool == nil {
		return fmt.Errorf("tool is nil")
	}
	name := tool.Definition().Name
	if name == "" {
		return fmt.Errorf("tool name is required")
	}
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %q is already registered", name)
	}
	r.tools[name] = tool
	return nil
}

func (r *Registry) Definitions() []ToolDefinition {
	definitions := make([]ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		definitions = append(definitions, tool.Definition())
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Name < definitions[j].Name })
	return definitions
}

func (r *Registry) Run(ctx context.Context, name string, args json.RawMessage) Result {
	tool, ok := r.tools[name]
	if !ok {
		return Result{Content: fmt.Sprintf("unknown tool %q", name), IsError: true}
	}
	return tool.Run(ctx, args)
}

func (r *Registry) Close() error { return nil }
