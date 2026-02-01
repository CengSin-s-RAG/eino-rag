package component

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type CorrectionInput struct {
	Correction string `json:"correction" jsonschema:"description=The specific logic correction or rule provided by the human"`
}

// NewCorrectionTool returns a tool for the human to submit logic corrections.
func NewCorrectionTool(path string) tool.BaseTool {
	newTool, _ := utils.InferTool(
		"submit_correction",
		"Use this tool ONLY when the human explicitly provides a correction to your analytical logic or a new rule to follow. These rules will be prioritized in all future analyses.",
		func(ctx context.Context, input *CorrectionInput) (string, error) {
			if input.Correction == "" {
				return "Error: correction content cannot be empty", nil
			}

			entry := fmt.Sprintf("- [%s]: %s\n", time.Now().Format("2006-01-02"), input.Correction)

			f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return "", fmt.Errorf("failed to open corrections file: %v", err)
			}
			defer f.Close()

			if _, err := f.WriteString(entry); err != nil {
				return "", fmt.Errorf("failed to write correction: %v", err)
			}

			return "Correction successfully recorded and will be applied to future turns.", nil
		},
	)
	return newTool
}
