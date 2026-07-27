// Package retrieval Retriever、Qdrant vector/full-text/hybrid
package retrieval

import (
	"context"
)

type Document struct{}

type Retriever interface {
	Search(ctx context.Context, query string, limit int) ([]Document, error)
}
