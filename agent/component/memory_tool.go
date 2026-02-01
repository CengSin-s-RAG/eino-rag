package component

import (
	"context"
	"os"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// NewResearchLogTool returns a tool for reading the RESEARCH_LOGS.md file.
func NewResearchLogTool(path string) tool.BaseTool {
	newTool, _ := utils.InferTool(
		"read_research_log",
		"Read the research log to understand previous findings and progress.",
		func(ctx context.Context, input struct{}) (string, error) {
			data, err := os.ReadFile(path)
			if err != nil {
				if os.IsNotExist(err) {
					return "Research log is empty.", nil
				}
				return "", err
			}
			return string(data), nil
		},
	)
	return newTool
}
