package component

import (
	"agent.article.fp/client"
	"agent.article.fp/dao"
	"agent.article.fp/model"
	"agent.article.fp/util"
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"log"
)

type MyContentRetrieverConfig struct {
	ColName    string
	Retrievers []string
}

var (
	defaultRetrieverConfig = &MyContentRetrieverConfig{
		ColName:    util.CollectionFupengshuo,
		Retrievers: []string{"vectors", "full_text"},
	}
)

func BuildContentRetriever(ctx context.Context, config *MyContentRetrieverConfig) (retriever.Retriever, error) {
	if config == nil {
		config = defaultRetrieverConfig
	}

	var retrieverSlice []retriever.Retriever
	for _, s := range config.Retrievers {
		r, err := retrieverFactory(ctx, s, config.ColName)
		if err != nil {
			return nil, err
		}
		retrieverSlice = append(retrieverSlice, r)
	}
	return &hybridRetriever{collection: config.ColName, retrievers: retrieverSlice}, nil
}

func retrieverFactory(ctx context.Context, s string, colName string) (retriever.Retriever, error) {
	switch s {
	case "vectors":
		return newVectorRetriever(ctx, client.Qdrant, colName)
	case "full_text":
		return newFullTextRetriever(ctx, client.Qdrant, colName)
	default:
		return nil, fmt.Errorf("[defaultRetriever] invalid retriever type: %s", s)
	}
}

// 自定义混合检索器结构体
type hybridRetriever struct {
	retrievers []retriever.Retriever
	collection string
}

// 实现 Retriever 接口的 Retrieve 方法
func (h *hybridRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// 1. 后端思维：并发调用两路检索，提高性能
	type result struct {
		docs []*schema.Document
		err  error
	}

	resChan := make(chan result, len(h.retrievers))
	for i := 0; i < len(h.retrievers); i++ {
		go func() {
			docs, err := h.retrievers[i].Retrieve(ctx, query, opts...)
			resChan <- result{docs, err}
		}()
	}

	// 2. 收集并处理结果
	var allDocs []*schema.Document
	for i := 0; i < 2; i++ {
		r := <-resChan
		if r.err != nil {
			return nil, r.err
		}
		allDocs = append(allDocs, r.docs...)
	}

	if len(allDocs) == 0 {
		return allDocs, nil
	}

	// 3. 确定性逻辑：去重 (基于 Document.ID)
	ids, _ := h.unique(allDocs)
	// 4. 从数据库查询源文档
	msgs, err := dao.GetArticlesByIds(ids)
	if err != nil {
		return nil, err
	}
	var docs []string
	msgIdMap := make(map[int]*model.ArticleEntries)
	for _, msg := range msgs {
		doc, err := msg.ToRerankDoc()
		if err != nil {
			log.Println("[convertToString]", err)
			continue
		}
		docs = append(docs, fmt.Sprintf("【id】%d\n\n【名称】%s\n\n【描述】%s\n\n【内容】%s", msg.Id, msg.Name(), msg.Desc(), doc))
		msg.Content = doc
		msgIdMap[msg.Id] = msg
	}
	// 5. 把源文档给大模型并重新排序
	transform, err := Rerank(ctx, &RerankReq{
		Question:  query,
		Documents: docs,
	})
	if err != nil {
		return nil, err
	}

	var returnDocs []*schema.Document
	// 计算每篇文档的得分并过滤
	for _, resp := range transform {
		if resp.Score < util.ScoreTs {
			continue
		}
		entries := msgIdMap[resp.Id]
		returnDocs = append(returnDocs, &schema.Document{
			MetaData: map[string]any{
				"title":      entries.Title,
				"created_at": entries.CreatedAt,
				"id":         entries.Id,
				"summary":    entries.ContentShort,
			},
		})
	}
	return returnDocs, nil
}

func (h *hybridRetriever) unique(docs []*schema.Document) ([]int64, []*schema.Document) {
	seen := make(map[int64]bool)
	var ids []int64
	uniqueDocs := make([]*schema.Document, 0, len(docs))
	for _, doc := range docs {
		rawId, ok := doc.MetaData["id"]
		if !ok {
			continue
		}

		var id int64
		switch v := rawId.(type) {
		case int64:
			id = v
		case int:
			id = int64(v)
		case float64:
			id = int64(v)
		default:
			continue
		}

		if !seen[id] {
			seen[id] = true
			uniqueDocs = append(uniqueDocs, doc)
			ids = append(ids, id)
		}
	}
	return ids, uniqueDocs
}
