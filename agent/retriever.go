package agent

import (
	"agent.article.fp/agent/component"
	"agent.article.fp/client"
	"agent.article.fp/dao"
	"agent.article.fp/model"
	"agent.article.fp/util"
	"context"
	"fmt"
	einoQdrant "github.com/cloudwego/eino-ext/components/retriever/qdrant"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/qdrant/go-client/qdrant"
	"log"
)

// newVectorRetriever 向量检索器
func newVectorRetriever(ctx context.Context, client *qdrant.Client, embedder embedding.Embedder) (retriever.Retriever, error) {
	return einoQdrant.NewRetriever(ctx, &einoQdrant.Config{
		Client:     client,
		Collection: util.CollectionFupengshuoName,
		Embedding:  embedder, // Eino 会自动用这个 embedder 把 query 转向量
		// 搜索参数
		TopK:           util.TopK,           // 每次取回 5 篇文章
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

type FullTextRetriever struct {
	client       *qdrant.Client
	returnFields []string
}

func (f *FullTextRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	filter := qdrant.Filter{
		Must: []*qdrant.Condition{
			qdrant.NewMatchText("textToIndex", query),
		},
	}
	points, err := f.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: util.CollectionFupengshuoName,
		Filter:         &filter,
		Limit:          &[]uint64{util.TopK}[0],
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}

	var docs []*schema.Document
	for _, point := range points {
		doc := &schema.Document{
			Content:  "",
			MetaData: map[string]any{},
		}
		if point.Id != nil {
			if uuid := point.Id.GetUuid(); uuid != "" {
				doc.ID = uuid
			} else {
				doc.ID = fmt.Sprintf("%d", point.Id.GetNum())
			}
		}

		for _, field := range f.returnFields {
			val, found := point.Payload[field]
			if !found {
				return nil, fmt.Errorf("[defaultResultParser] field=%s not found in payload, point=%v", field, point)
			}

			if field == "content" {
				doc.Content = val.GetStringValue()
			} else if field == "metadata" {
				doc.MetaData["metadata"] = val.GetStructValue().Fields
			} else {
				switch val.GetKind().(type) {
				case *qdrant.Value_NullValue:
					doc.MetaData[field] = val.GetNullValue()
				case *qdrant.Value_DoubleValue:
					doc.MetaData[field] = val.GetDoubleValue()
				case *qdrant.Value_IntegerValue:
					doc.MetaData[field] = val.GetIntegerValue()
				case *qdrant.Value_StringValue:
					doc.MetaData[field] = val.GetStringValue()
				case *qdrant.Value_BoolValue:
					doc.MetaData[field] = val.GetBoolValue()
				case *qdrant.Value_StructValue:
					doc.MetaData[field] = val.GetStructValue().Fields
				case *qdrant.Value_ListValue:
					doc.MetaData[field] = val.GetListValue()
				}
			}
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// newFullTextRetriever 全文检索器 (新增)
// 注意：这要求你在 Qdrant 的 textToIndex 字段上已经建立了 'text' 类型的索引
func newFullTextRetriever(ctx context.Context, client *qdrant.Client) (retriever.Retriever, error) {
	return &FullTextRetriever{
		client: client,
		returnFields: []string{"textToIndex",
			"title",
			"created_at",
			"id",
			"chunk_index",
			"summary",
		},
	}, nil
}

func BuildHybridRetriever(ctx context.Context) (retriever.Retriever, error) {
	embedder, err := newEinoEmbedder(ctx)
	if err != nil {
		return nil, err
	}

	vectors, err := newVectorRetriever(ctx, client.Qdrant, embedder)
	if err != nil {
		return nil, err
	}

	fullText, err := newFullTextRetriever(ctx, client.Qdrant)
	if err != nil {
		return nil, err
	}
	return &hybridRetriever{vector: vectors, fullText: fullText}, nil
}

// 自定义混合检索器结构体
type hybridRetriever struct {
	vector   retriever.Retriever
	fullText retriever.Retriever
}

// 实现 Retriever 接口的 Retrieve 方法
func (h *hybridRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// 1. 后端思维：并发调用两路检索，提高性能
	type result struct {
		docs []*schema.Document
		err  error
	}
	resChan := make(chan result, 2)

	go func() {
		docs, err := h.vector.Retrieve(ctx, query, opts...)
		resChan <- result{docs, err}
	}()
	go func() {
		docs, err := h.fullText.Retrieve(ctx, query, opts...)
		resChan <- result{docs, err}
	}()

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
	transform, err := component.Rerank(ctx, &component.RerankReq{
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

func (h *hybridRetriever) unique(docs []*schema.Document) ([]string, []*schema.Document) {
	seen := make(map[string]bool)
	var ids []string
	uniqueDocs := make([]*schema.Document, 0, len(docs))
	for _, doc := range docs {
		id := doc.MetaData["id"].(string)
		if !seen[id] {
			seen[id] = true
			uniqueDocs = append(uniqueDocs, doc)
			ids = append(ids, id)
		}
	}
	return ids, uniqueDocs
}
