package component

import (
	"agent.article.fp/util"
	"context"
	einoQdrant "github.com/cloudwego/eino-ext/components/retriever/qdrant"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/qdrant/go-client/qdrant"
)

// newVectorRetriever 向量检索器
func newVectorRetriever(ctx context.Context, client *qdrant.Client, colName string) (retriever.Retriever, error) {
	embedder, err := newEinoEmbedder(ctx)
	if err != nil {
		return nil, err
	}
	return einoQdrant.NewRetriever(ctx, &einoQdrant.Config{
		Client:         client,
		Collection:     colName,
		Embedding:      embedder,            // Eino 会自动用这个 embedder 把 query 转向量
		TopK:           util.TopK,           // 搜索参数 每次取回 5 篇文章
		ScoreThreshold: util.ScoreThreshold, // 可选：相似度阈值，过滤掉不相关的
		ReturnFields: []string{
			"textToIndex",
			"title",
			"created_at",
			"id",
			"chunk_index",
			"summary",
		},
	})
}
