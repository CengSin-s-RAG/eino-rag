package client

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

func NewRetrieverTool(r retriever.Retriever, name, desc string) tool.BaseTool {
	runFunc := func(ctx context.Context, input struct {
		Query string `json:"query" jsonschema:"description=The search query to find relevant documents"`
	}) (string, error) {
		docs, err := r.Retrieve(ctx, input.Query)
		if err != nil {
			return "", fmt.Errorf("retrieval failed: %w", err)
		}

		if len(docs) == 0 {
			return "No relevant documents found.", nil
		}

		// 2. 格式化文档
		var result string
		for i, doc := range docs {
			result += fmt.Sprintf("Document %d:\n%s\n---\n", i+1, doc.Content)
		}
		return result, nil
	}

	newTool := utils.NewTool(&schema.ToolInfo{Name: name, Desc: desc}, runFunc)
	return newTool
}
