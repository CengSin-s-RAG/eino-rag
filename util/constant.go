package util

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	ChatHistoryFormat        = "chatHistory:%s"
	CollectionName           = "financial_articles"
	CollectionFupengshuoName = "fupengshuo_articles"
	NewsCollectionName       = "724_news_col"
	ModelName                = "openai/gpt-5"
)

var (
	SystemPrompt string
)

func InitSystemPrompt(path string) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalln("open system prompt file failed, err ", err)
	}
	defer file.Close()

	contextBtys, err := io.ReadAll(file)
	if err != nil {
		log.Fatalln(fmt.Errorf("read system prompt file failed, err %v", err))
	}

	SystemPrompt = string(contextBtys)
}
