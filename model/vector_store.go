package model

import (
	"fmt"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// VectorStore 向量存储模型
type VectorStore struct {
	ID          int64          `gorm:"primaryKey" json:"id"`
	TextToIndex string         `gorm:"type:text" json:"textToIndex"`
	Title       string         `gorm:"type:varchar(500)" json:"title"`
	CreatedAt   string         `gorm:"type:varchar(50)" json:"created_at"`
	ChunkIndex  int            `gorm:"type:int" json:"chunk_index"`
	Summary     string         `gorm:"type:text" json:"summary"`
	Embedding   pgvector.Vector `gorm:"type:vector(1536)" json:"embedding"` // 1536 是 embedding 维度
	Collection  string         `gorm:"type:varchar(100);index" json:"collection"`
}

func (v *VectorStore) TableName() string {
	return "vector_store"
}

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

// CreateIndexIfNotExists 创建向量索引（如果不存在）
func CreateIndexIfNotExists(db *gorm.DB, collection string) error {
	indexName := fmt.Sprintf("idx_vector_store_%s_embedding", collection)
	return db.Exec(fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS %s ON vector_store 
		USING ivfflat (embedding vector_cosine_ops)
		WITH (lists = 100)
		WHERE collection = '%s'
	`, indexName, collection)).Error
}

// CreateFullTextIndexIfNotExists 创建全文搜索索引（如果不存在）
func CreateFullTextIndexIfNotExists(db *gorm.DB) error {
	return db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_vector_store_text_search 
		ON vector_store USING gin(to_tsvector('english', text_to_index))
	`).Error
}
