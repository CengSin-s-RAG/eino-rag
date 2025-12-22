package agent

import (
	"agent.article.fp/util"
	"context"
	einoQdrant "github.com/cloudwego/eino-ext/components/retriever/qdrant"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/qdrant/go-client/qdrant"
)

func NewEinoRetriever(ctx context.Context, client *qdrant.Client, embedder embedding.Embedder) (retriever.Retriever, error) {
	return einoQdrant.NewRetriever(ctx, &einoQdrant.Config{
		Client:     client,
		Collection: util.CollectionName,
		Embedding:  embedder, // Eino 会自动用这个 embedder 把 query 转向量
		// 搜索参数
		TopK: 5, // 每次取回 5 篇文章
		// ScoreThreshold: 0.7, // 可选：相似度阈值，过滤掉不相关的
	})
}
