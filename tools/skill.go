package tools

import (
	"agent.article.fp/skill"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type skillTool struct{ manager *skill.Manager }

func NewSkill(manager *skill.Manager) Tool { return skillTool{manager: manager} }

func (skillTool) Definition() ToolDefinition {
	return ToolDefinition{
		Name: "activate_skill", Description: "Load the full instructions for an available skill.",
		Parameters: objectSchema(map[string]any{"name": stringSchema("The skill name.")}, "name"),
	}
}

func (skillTool) Safety() Safety {
	return Safety{SideEffect: "read", Permission: "allow", Reason: "reads local skill instructions"}
}

func (t skillTool) Run(_ context.Context, raw json.RawMessage) Result {
	var input struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &input); err != nil || strings.TrimSpace(input.Name) == "" {
		return Result{Content: "activate_skill requires a non-empty name", IsError: true}
	}
	s, err := t.manager.GetSkill(input.Name)
	if err != nil {
		return Result{Content: err.Error(), IsError: true}
	}
	var out strings.Builder
	fmt.Fprintf(&out, "<skill_content name=%q>\n%s", s.Name, s.Body)
	if resources := t.manager.ListResources(s.Name); len(resources) > 0 {
		out.WriteString("\n<skill_resources>\n")
		for _, resource := range resources {
			fmt.Fprintf(&out, "  <file>%s</file>\n", resource)
		}
		out.WriteString("</skill_resources>")
	}
	fmt.Fprintf(&out, "\nSkill directory: %s\n</skill_content>", s.BaseDir)
	return Result{Content: out.String()}
}

func objectSchema(properties map[string]any, required ...string) map[string]any {
	return map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
}

func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}
