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
	colName string
}

func (f *FullTextRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// 使用 PostgreSQL 的全文搜索功能
	// plainto_tsquery 可以处理普通文本查询，自动处理空格和标点
	
	var results []model.VectorStore
	err := client.IvankaContent.
		Raw(`
			SELECT *
			FROM vector_store
			WHERE collection = $1
			AND to_tsvector('english', text_to_index) @@ plainto_tsquery('english', $2)
			ORDER BY ts_rank(to_tsvector('english', text_to_index), plainto_tsquery('english', $3)) DESC
			LIMIT $4
		`, f.colName, query, query, util.TopK).
		Scan(&results).Error
	
	if err != nil {
		return nil, err
	}

	var docs []*schema.Document
	for _, result := range results {
		doc := &schema.Document{
			ID:      fmt.Sprintf("%d", result.ID),
			Content: result.TextToIndex,
			MetaData: map[string]any{
				"textToIndex": result.TextToIndex,
				"title":      result.Title,
				"created_at": result.CreatedAt,
				"id":         result.ID,
				"chunk_index": result.ChunkIndex,
				"summary":    result.Summary,
			},
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// newFullTextRetriever 全文检索器
// 使用 PostgreSQL 的 tsvector 和 to_tsquery 进行全文搜索
func newFullTextRetriever(ctx context.Context, colName string) (retriever.Retriever, error) {
	return &FullTextRetriever{
		colName: colName,
	}, nil
}
