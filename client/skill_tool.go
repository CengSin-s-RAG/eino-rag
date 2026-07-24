package client

import (
	"agent.article.fp/skill"
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"strings"
)

// NewActivateSkillTool 创建 activate_skill 工具，LLM 可通过此工具加载完整 Skill 指令
func NewActivateSkillTool(mgr *skill.Manager) tool.BaseTool {
	runFunc := func(ctx context.Context, input struct {
		Name string `json:"name" jsonschema:"description=The name of the skill to activate"`
	}) (string, error) {
		s, err := mgr.GetSkill(input.Name)
		if err != nil {
			return "", err
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("<skill_content name=\"%s\">\n", s.Name))
		sb.WriteString(s.Body)

		resources := mgr.ListResources(s.Name)
		if len(resources) > 0 {
			sb.WriteString("\n\n<skill_resources>\n")
			for _, r := range resources {
				sb.WriteString(fmt.Sprintf("  <file>%s</file>\n", r))
			}
			sb.WriteString("</skill_resources>\n")
		}

		sb.WriteString(fmt.Sprintf("\nSkill directory: %s\n", s.BaseDir))
		sb.WriteString("Relative paths in this skill are relative to the skill directory.\n")
		sb.WriteString("</skill_content>")
		return sb.String(), nil
	}

	newTool, _ := utils.InferTool("activate_skill", "Activate a skill by name to load its full instructions. Use when a task matches a skill's description.", runFunc)
	return newTool
}
