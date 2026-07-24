package util

import (
	"fmt"
	"io"
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
	ScoreTs        = 0.3
	ScoreThreshold = GetScoreThreshold(ScoreTs)
)

func GetScoreThreshold(n float64) *float64 {
	return &n
}

var (
	SystemPrompt       string
	RerankPrompt       string
	QueryRewritePrompt string
)

func InitPrompt(path string) string {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalln("open system prompt file failed, err ", err)
	}
	defer file.Close()

	contextBtys, err := io.ReadAll(file)
	if err != nil {
		log.Fatalln(fmt.Errorf("read system prompt file failed, err %v", err))
	}

	return string(contextBtys)
}
