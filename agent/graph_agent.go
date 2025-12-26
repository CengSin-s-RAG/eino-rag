package agent

import (
	"agent.article.fp/client"
	"agent.article.fp/util"
	"agent.article.fp/visualize"
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"time"
)

// EinoChatAgent 封装编译好的 Runnable，对外提供服务
type EinoChatAgent struct {
	Runner *react.Agent
}

func NewEinoChatAgent(ctx context.Context, config openai.ChatModelConfig) (*EinoChatAgent, error) {
	reactAgent, err := newReactLambdaAgent(ctx, config)
	if err != nil {
		return nil, err
	}

	anyGraph, opts := reactAgent.ExportGraph()
	genAgentConfigImage := visualize.NewMermaidGenerator("/Users/cengsin/my_projects/Eino-Projects/fupeng-article-agent")

	// 1. 创建 Graph 容器
	// 泛型明确指定了输入输出都是 *ChatState
	graph := compose.NewGraph[*ChatState, *ChatState]()

	_ = graph.AddGraphNode("react_agent", anyGraph, opts...)
	// 3. 定义边 (AddEdge) - 决定执行顺序
	_ = graph.AddEdge(compose.START, "react_agent")
	_ = graph.AddEdge("react_agent", compose.END)
	_, _ = graph.Compile(ctx, compose.WithGraphCompileCallbacks(genAgentConfigImage))
	return &EinoChatAgent{Runner: reactAgent}, nil
}

func newReactLambdaAgent(ctx context.Context, config openai.ChatModelConfig) (*react.Agent, error) {
	// Node: LLM (这里我们将 build 逻辑其实应该提出来，避免每次 Invoke 都 NewModel)
	// 为了性能，我们先在外面初始化好 Model 和 Template
	// 1. 初始化组件 (生产环境建议在 build 时初始化一次，为了代码清晰这里仅做演示)
	// 实际项目中，chatModel 和 tmpl 应该在 NewEinoAgent 时创建好并闭包进来
	// 1. 定义 Chat Model (复用你的配置)
	chatModel, err := openai.NewChatModel(ctx, &config)
	if err != nil {
		return nil, err
	}

	embedder, err := NewEinoEmbedder(ctx)
	if err != nil {
		return nil, err
	}

	retrv, err := NewEinoRetriever(ctx, client.Qdrant, embedder)
	if err != nil {
		return nil, err
	}

	kbTool := client.NewRetrieverTool(retrv,
		"search_financial_knowledge",
		"Use this tool to search for internal financial reports, news, and articles.",
	)

	info, err := kbTool.Info(ctx)
	if err != nil {
		return nil, err
	}
	client.ToolsInfo = append(client.ToolsInfo, info)
	client.EinoTools = append(client.EinoTools, kbTool)
	if err = chatModel.BindTools(client.ToolsInfo); err != nil {
		return nil, err
	}

	agentConfig := react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig:      compose.ToolsNodeConfig{Tools: client.EinoTools},
		MaxStep:          20,
		MessageModifier: func(ctx context.Context, input []*schema.Message) []*schema.Message {
			if len(input) > 3 { // 滑动窗口，系统提示词和最近的19条信息
				input = append(input[:1], input[len(input)-2:]...)
			}

			if len(input) > 0 && input[0].Role != schema.System {
				input = append([]*schema.Message{schema.SystemMessage(util.SystemPrompt + fmt.Sprintf("\n\n 当前时间: %s", time.Now().Format(time.DateTime)))}, input...)
			}
			return input
		},
		MessageRewriter: func(ctx context.Context, input []*schema.Message) []*schema.Message {
			return input
		},
	}

	agent, err := react.NewAgent(ctx, &agentConfig)
	if err != nil {
		return nil, err
	}

	return agent, nil
}
