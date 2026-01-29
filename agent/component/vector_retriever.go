package component

import (
	"agent.article.fp/client"
	"agent.article.fp/model"
	"agent.article.fp/util"
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/pgvector/pgvector-go"
)

// PgVectorRetriever PostgreSQL + pgvector 向量检索器
type PgVectorRetriever struct {
	colName        string
	embedder       func(ctx context.Context, text string) ([]float32, error)
	topK           int
	scoreThreshold float64
}

// VectorStoreWithSimilarity 带相似度分数的向量存储结果
type VectorStoreWithSimilarity struct {
	model.VectorStore
	Similarity float64 `gorm:"column:similarity"`
}

func (p *PgVectorRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// 1. 将查询文本转换为向量
	embedding, err := p.embedder(ctx, query)
	if err != nil {
		return nil, err
	}

	// 2. 使用 pgvector 进行相似度搜索
	queryVector := pgvector.NewVector(embedding)

	// 使用原生 SQL 进行向量相似度搜索（余弦距离转换为相似度）
	var results []VectorStoreWithSimilarity
	err = client.IvankaContent.
		Raw(`
			SELECT 
				id, text_to_index, title, created_at, chunk_index, summary, embedding, collection,
				1 - (embedding <=> ?::vector) as similarity
			FROM vector_store
			WHERE collection = ?
			ORDER BY embedding <=> ?::vector
			LIMIT ?
		`, queryVector.String(), p.colName, queryVector.String(), p.topK).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// 3. 转换为 Document 格式
	var docs []*schema.Document
	for _, result := range results {
		// 过滤低于阈值的结果
		if result.Similarity < p.scoreThreshold {
			continue
		}

		doc := &schema.Document{
			ID:      fmt.Sprintf("%d", result.Id),
			Content: result.TextToIndex,
			MetaData: map[string]any{
				"textToIndex": result.TextToIndex,
				"title":       result.Title,
				"created_at":  result.CreatedAt,
				"id":          result.Id,
				"chunk_index": result.ChunkIndex,
				"summary":     result.Summary,
			},
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// newVectorRetriever 向量检索器
func newVectorRetriever(ctx context.Context, colName string) (retriever.Retriever, error) {
	embedder, err := newEinoEmbedder(ctx)
	if err != nil {
		return nil, err
	}

	// 包装 embedder 为函数
	embedFunc := func(ctx context.Context, text string) ([]float32, error) {
		embeddings, err := embedder.EmbedStrings(ctx, []string{text})
		if err != nil {
			return nil, err
		}
		if len(embeddings) == 0 {
			return nil, fmt.Errorf("empty embedding result")
		}
		// 转换为 float32 切片
		result := make([]float32, len(embeddings[0]))
		for i, v := range embeddings[0] {
			result[i] = float32(v)
		}
		return result, nil
	}

	threshold := 0.0
	if util.ScoreThreshold != nil {
		threshold = *util.ScoreThreshold
	}

	return &PgVectorRetriever{
		colName:        colName,
		embedder:       embedFunc,
		topK:           util.TopK,
		scoreThreshold: threshold,
	}, nil
}
