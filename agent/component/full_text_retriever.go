package component

import (
	"agent.article.fp/util"
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/qdrant/go-client/qdrant"
)

type FullTextRetriever struct {
	client       *qdrant.Client
	returnFields []string
	colName      string
}

func (f *FullTextRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	filter := qdrant.Filter{
		Must: []*qdrant.Condition{
			qdrant.NewMatchText("textToIndex", query),
		},
	}
	points, err := f.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: f.colName,
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
func newFullTextRetriever(ctx context.Context, client *qdrant.Client, colName string) (retriever.Retriever, error) {
	return &FullTextRetriever{
		colName: colName,
		client:  client,
		returnFields: []string{"textToIndex",
			"title",
			"created_at",
			"id",
			"chunk_index",
			"summary",
		},
	}, nil
}
