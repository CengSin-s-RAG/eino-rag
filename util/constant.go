package util

import (
	_ "embed"
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
	ScoreTs        = 0.3
	ScoreThreshold = GetScoreThreshold(ScoreTs)
)

func GetScoreThreshold(n float64) *float64 {
	return &n
}

var (
	//go:embed systemPrompt.md
	SystemPrompt string

	//go:embed rerankSystemPrompt.txt
	RerankPrompt string

	//go:embed rerankSystemPrompt.txt
	QueryRewritePrompt string
)
