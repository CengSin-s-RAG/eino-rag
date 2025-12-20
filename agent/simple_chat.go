package agent

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// EinoChatAgent 封装编译好的 Runnable，对外提供服务
type EinoChatAgent struct {
	Runnable compose.Runnable[map[string]any, *schema.Message]
}

func NewEinoChatAgent(ctx context.Context, config openai.ChatModelConfig) (*EinoChatAgent, error) {
	// 1. 定义 Chat Model (复用你的配置)
	chatModel, err := openai.NewChatModel(ctx, &config)
	if err != nil {
		return nil, err
	}

	// 2. 定义 Prompt 模版
	// 这里我们做一个简单的 Chat Template，支持 System Prompt 和 User Input
	// Placeholder 语法：{variable_name}
	tmpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个金融助手，请用简练的语言回答问题。"),
		schema.UserMessage("{query}"))

	// 3. 编排 Chain: Input Map -> Prompt -> ChatModel -> Output Message
	chain := compose.NewChain[map[string]any, *schema.Message]()

	runner, err := chain.
		AppendChatTemplate(tmpl).
		AppendChatModel(chatModel).
		Compile(ctx)
	if err != nil {
		return nil, err
	}

	return &EinoChatAgent{Runnable: runner}, nil
}

func (a *EinoChatAgent) Run(ctx context.Context, query string) (*schema.StreamReader[*schema.Message], error) {
	// 调用 Eino Chain
	// Map 中的 Key 必须和 Prompt 模板里的占位符 "{query}" 对应
	return a.Runnable.Stream(ctx, map[string]any{
		"query": query,
	})
}
