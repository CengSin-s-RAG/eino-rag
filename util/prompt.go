package util

import (
	"fmt"
	"os"
	"time"
)

func GetSystemPrompt() string {
	prompt := SoulPrompt + "\n\n" + SystemPrompt

	// Inject Human-Correction Rules
	if corrections, err := os.ReadFile("memory/CORRECTIONS.md"); err == nil && len(corrections) > 0 {
		prompt += "\n\n### CRITICAL: HUMAN-CORRECTION RULES\n"
		prompt += "The following rules were provided by the human and MUST take precedence over any retrieved documents or internal model knowledge:\n"
		prompt += string(corrections)
	}

	return prompt + fmt.Sprintf("\n\n 当前时间: %s", time.Now().Format(time.DateTime))
}
