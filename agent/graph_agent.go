package agent

import (
	"agent.article.fp/agent/component"
	"agent.article.fp/client"
	"agent.article.fp/util"
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

// EinoChatAgent 封装编译好的 Runnable，对外提供服务
type EinoChatAgent struct {
	Runner *react.Agent
	Model  *openai.ChatModel
}

func NewEinoChatAgent(ctx context.Context, config openai.ChatModelConfig) (*EinoChatAgent, error) {
	// 1. 定义 Chat Model (复用你的配置)
	chatModel, err := openai.NewChatModel(ctx, &config)
	if err != nil {
		return nil, err
	}

	reactAgent, err := newReactLambdaAgent(ctx, chatModel)
	if err != nil {
		return nil, err
	}
	return &EinoChatAgent{Runner: reactAgent, Model: chatModel}, nil
}

func newReactLambdaAgent(ctx context.Context, chatModel *openai.ChatModel) (*react.Agent, error) {
	hybirdRetriever, err := component.BuildContentRetriever(ctx, &component.MyContentRetrieverConfig{
		ColName:    util.CollectionFupengshuo,
		Retrievers: []string{"vectors", "full_text"},
	})
	if err != nil {
		return nil, err
	}

	kbTool := client.NewRetrieverTool(hybirdRetriever,
		"search_financial_knowledge",
		"Use this tool to search for internal financial reports, news, and articles.",
	)
	info, err := kbTool.Info(ctx)
	if err != nil {
		return nil, err
	}
	client.ToolsInfo = append(client.ToolsInfo, info)
	client.EinoTools = append(client.EinoTools, kbTool)

	// Add Research Log Tool
	researchLogTool := component.NewResearchLogTool("memory/RESEARCH_LOGS.md")
	researchLogInfo, err := researchLogTool.Info(ctx)
	if err == nil {
		client.ToolsInfo = append(client.ToolsInfo, researchLogInfo)
		client.EinoTools = append(client.EinoTools, researchLogTool)
	}

	// Add HITL Correction Tool
	correctionTool := component.NewCorrectionTool("memory/CORRECTIONS.md")
	correctionInfo, err := correctionTool.Info(ctx)
	if err == nil {
		client.ToolsInfo = append(client.ToolsInfo, correctionInfo)
		client.EinoTools = append(client.EinoTools, correctionTool)
	}

	if err = chatModel.BindTools(client.ToolsInfo); err != nil {
		return nil, err
	}

	agentConfig := react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig:      compose.ToolsNodeConfig{Tools: client.EinoTools},
		MaxStep:          6,
		MessageModifier: func(ctx context.Context, input []*schema.Message) []*schema.Message {
			// Ensure history is not empty
			if len(input) == 0 {
				return input
			}

			// Force reload system prompt if corrections file changed
			// (GetSystemPrompt handles the reading logic)
			sysPrompt := util.GetSystemPrompt()

			// Check if we already have a system message at the start
			if input[0].Role == schema.System {
				// Create a copy to avoid mutating original state
				newInput := make([]*schema.Message, len(input))
				copy(newInput, input)
				newInput[0] = schema.SystemMessage(sysPrompt)
				input = newInput
			} else {
				// Create a new slice to avoid modifying the input slice directly
				newInput := make([]*schema.Message, 0, len(input)+1)
				newInput = append(newInput, schema.SystemMessage(sysPrompt))
				newInput = append(newInput, input...)
				input = newInput
			}

			// 使用小模型进行摘要 (if history is long)
			if len(input) > 20 && schema.User == input[len(input)-1].Role {
				newQuery, err := component.QueryRewriting(input[1:])
				if err != nil {
					return input
				}
				return []*schema.Message{schema.SystemMessage(sysPrompt), schema.UserMessage(newQuery)}
			}
			return input
		},
	}

	agent, err := react.NewAgent(ctx, &agentConfig)
	if err != nil {
		return nil, err
	}

	return agent, nil
}
