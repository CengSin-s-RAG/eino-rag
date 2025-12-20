package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"log"
	"os"
)

func main() {
	ctx := context.Background()
	// ⚠️ 记得设置环境变量 OPENAI_API_KEY，或者替换为你的 OpenRouter 配置

	// 先初始化所需的 chatModel
	config := openai.ChatModelConfig{
		APIKey:  os.Getenv("OPENROUTER_API_KEY"),
		BaseURL: os.Getenv("OPENROUTER_API_BASE_URL"),
		Model:   os.Getenv("OPENROUTER_MODEL"),
	}
	// 1. 定义组件：ChatModel (后端核心)
	// 这里使用 OpenAI 作为示例，Eino 支持通过配置切换底层实现
	chatModel, err := openai.NewChatModel(ctx, &config)
	if err != nil {
		log.Fatalln(fmt.Errorf("NewChatModel error: %v", err))
	}

	// 2. 定义组件：Prompt Template (请求组装)
	// 它的作用是把 input map 转换成 message 列表
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage(`你是一个资深的 Golang 工程师，只用简短的代码回答问题。`),
		schema.UserMessage(`如何实现{query}`)) // {query} 是占位符

	// 3. 编排：构建 Chain (流水线)
	// input 是 map[string]any (为了填充模板), output 是 *schema.Message (模型回复)
	chain := compose.NewChain[map[string]any, *schema.Message]()

	// 串联: 输入 -> 模板 -> 模型 -> 输出
	// AppendChatModel 会自动处理从模板输出(Message list)到模型输入的转换
	runner, err := chain.AppendChatTemplate(template).AppendChatModel(chatModel).Compile(ctx)
	if err != nil {
		log.Fatalln(fmt.Errorf("chain compile error: %v", err))
	}

	output, err := runner.Invoke(ctx, map[string]any{
		"query": "实现一个输出hello world的程序",
	})
	if err != nil {
		log.Fatalln(fmt.Errorf("chain invoke error: %v", err))
	}
	fmt.Println("模型回复:", output.Content)
}
