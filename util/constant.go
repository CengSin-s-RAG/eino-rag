package util

import (
	"fmt"
	"log"
	"os"
)

const (
	ChatHistoryPrefix        = "chatHistory"
	CollectionName           = "financial_articles"
	CollectionFupengshuo     = "fupengshuo-contents"
	CollectionFupengshuoName = "fupengshuo_articles"
	NewsCollectionName       = "724_news_col"
	ModelName                = "openai/gpt-5"
)

const (
	TopK = 10
)

var (
	ScoreTs        = 0.38
	ScoreThreshold = GetScoreThreshold(ScoreTs)
)

func GetScoreThreshold(n float64) *float64 {
	return &n
}

var (
	SoulPrompt         string
	SystemPrompt       string
	RerankPrompt       string
	QueryRewritePrompt string
)

func InitPrompt(path string) string {
	contextBtys, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln(fmt.Errorf("read prompt file %s failed: %v", path, err))
	}

	return string(contextBtys)
}
