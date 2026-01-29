package component

import (
	"agent.article.fp/client"
	"agent.article.fp/model"
	"agent.article.fp/util"
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

type FullTextRetriever struct {
	colName        string
	topK           int
	scoreThreshold float64
}

// FullTextSearchResult 全文搜索结果（带相关性分数）
type FullTextSearchResult struct {
	model.VectorStore
	Rank float64 `gorm:"column:rank"`
}

func (f *FullTextRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// 使用 PostgreSQL 的全文搜索功能
	// plainto_tsquery 可以处理普通文本查询，自动处理空格和标点

	// 参数验证
	if f.colName == "" {
		return nil, fmt.Errorf("collection name is empty")
	}
	if query == "" {
		return nil, fmt.Errorf("query is empty")
	}

	// 使用 ts_rank 计算文本相关性分数
	// plainto_tsquery 会自动处理中文分词和查询优化
	var results []FullTextSearchResult
	err := client.IvankaContent.
		Raw(`
			SELECT 
				id, text_to_index, title, created_at, chunk_index, summary, embedding, collection,
				ts_rank(to_tsvector('english', text_to_index), plainto_tsquery('english', ?)) as rank
			FROM vector_store
			WHERE collection = ?
				AND to_tsvector('english', text_to_index) @@ plainto_tsquery('english', ?)
			ORDER BY rank DESC
			LIMIT ?
		`, query, f.colName, query, f.topK).
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("full text search failed: %w", err)
	}

	// 转换为 Document 格式
	var docs []*schema.Document
	for _, result := range results {
		// 过滤低于阈值的结果
		if result.Rank < f.scoreThreshold {
			continue
		}

		doc := &schema.Document{
			ID:      result.Id,
			Content: result.TextToIndex,
			MetaData: map[string]any{
				"textToIndex": result.TextToIndex,
				"title":       result.Title,
				"created_at":  result.CreatedAt,
				"id":          result.Id,
				"chunk_index": result.ChunkIndex,
				"summary":     result.Summary,
				"rank":        result.Rank,
			},
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// newFullTextRetriever 全文检索器
// 使用 PostgreSQL 的 tsvector 和 to_tsquery 进行全文搜索
func newFullTextRetriever(ctx context.Context, colName string) (retriever.Retriever, error) {
	threshold := 0.0
	if util.ScoreThreshold != nil {
		threshold = *util.ScoreThreshold
	}

	return &FullTextRetriever{
		colName:        colName,
		topK:           util.TopK,
		scoreThreshold: threshold,
	}, nil
}
