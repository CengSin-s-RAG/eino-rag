package util

import (
	"fmt"
	"time"
)

// GetSystemPrompt 返回系统提示词，skillCatalog 为 Skill 目录（可为空）
func GetSystemPrompt(skillCatalog string) string {
	prompt := SystemPrompt + fmt.Sprintf("\n\n 当前时间: %s", time.Now().Format(time.DateTime))
	if skillCatalog != "" {
		prompt += "\n\n" + skillCatalog
	}
	return prompt
}
