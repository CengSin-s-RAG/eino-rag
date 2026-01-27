package component

import (
	"context"
	"github.com/cloudwego/eino-ext/libs/acl/openai"
	"github.com/cloudwego/eino/components/embedding"
	"net/http"
	"os"
)

// newEinoEmbedder 在项目中，Embedding 主要是为了给 Retriever 用的。Eino 有一个非常强大的概念叫 Retriever，它会自动调用 Embedder 把用户的 Query 变成向量，然后去向量数据库查。
func newEinoEmbedder(ctx context.Context) (embedding.Embedder, error) {
	// todo 修改为可配置化的模型调用
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseURL := os.Getenv("OPENROUTER_API_BASE_URL")
	model := "qwen/qwen3-embedding-8b"
	dimensions := 1536
	format := openai.EmbeddingEncodingFormatFloat
	return openai.NewEmbeddingClient(ctx, &openai.EmbeddingConfig{
		HTTPClient:     http.DefaultClient,
		APIKey:         apiKey,
		BaseURL:        baseURL,
		Model:          model,
		EncodingFormat: &format,
		Dimensions:     &dimensions,
	})
}
