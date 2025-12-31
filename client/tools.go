package client

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
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
		for _, doc := range docs {
			result += fmt.Sprintf("文章ID:[%d] 来源:%s (发布时间:%s)\n内容:%v文章链接:%s\n---\n",
				doc.MetaData["id"], doc.MetaData["title"], doc.MetaData["created_at"], doc.MetaData["textToIndex"], fmt.Sprintf("https://wallstreetcn.com/premium/articles/%d", doc.MetaData["id"]))
		}
		return result, nil
	}

	newTool, _ := utils.InferTool(name, desc, runFunc)
	return newTool
}
