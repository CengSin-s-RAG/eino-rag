package model

import (
	"fmt"
	"strings"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// VectorStore 向量存储模型，对应 vector_store 表
type VectorStore struct {
	Id          string          `gorm:"type:uuid;primaryKey"`
	Embedding   pgvector.Vector `gorm:"type:vector(1536);not null"`
	TextToIndex string          `gorm:"column:text_to_index;type:text"`
	Title       string          `gorm:"type:text"`
	CreatedAt   int64           `gorm:"column:created_at;type:bigint"`
	SourceId    int64           `gorm:"column:source_id;type:bigint"`
	ChunkIndex  int             `gorm:"column:chunk_index;type:int"`
	Summary     string          `gorm:"type:text"`
}

func (VectorStore) TableName() string { return "vector_stores" }

// CreateTableIfNotExists 创建表（如果不存在）
func CreateTableIfNotExists(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS vector_store (
			id BIGSERIAL PRIMARY KEY,
			text_to_index TEXT,
			title VARCHAR(500),
			created_at VARCHAR(50),
			chunk_index INTEGER,
			summary TEXT,
			embedding vector(1536),
			collection VARCHAR(100)
		)
	`).Error
}

// sanitizeIndexName 将 collection 转为合法索引名（避免连字符等导致 SQL 解析错误）
func sanitizeIndexName(collection string) string {
	s := strings.ReplaceAll(collection, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

// CreateIndexIfNotExists 创建按 collection 的向量索引（如果不存在）
func CreateIndexIfNotExists(db *gorm.DB, collection string) error {
	indexName := "idx_vector_store_" + sanitizeIndexName(collection) + "_embedding"
	return db.Exec(fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS %s ON vector_store 
		USING ivfflat (embedding vector_cosine_ops)
		WITH (lists = 100)
		WHERE collection = '%s'
	`, indexName, collection)).Error
}

// CreateGlobalVectorIndexIfNotExists 创建全局向量索引（不区分 collection，供所有向量查询使用）
func CreateGlobalVectorIndexIfNotExists(db *gorm.DB) error {
	return db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_vector_store_embedding ON vector_store 
		USING ivfflat (embedding vector_cosine_ops)
		WITH (lists = 100)
	`).Error
}

// CreateFullTextIndexIfNotExists 创建全文搜索索引（如果不存在）
func CreateFullTextIndexIfNotExists(db *gorm.DB) error {
	return db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_vector_store_text_search 
		ON vector_store USING gin(to_tsvector('english', text_to_index))
	`).Error
}
