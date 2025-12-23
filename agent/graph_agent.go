package agent

import (
	"agent.article.fp/client"
	"agent.article.fp/dao"
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	oa "github.com/sashabaranov/go-openai"
)

// EinoChatAgent 封装编译好的 Runnable，对外提供服务
type EinoChatAgent struct {
	Runnable compose.Runnable[*ChatState, *ChatState]
}

// ----------------------------------------------------------------
// 节点 1: History Loader
// 职责：读取数据库 -> 填充 State.History
// ----------------------------------------------------------------
func loadHistoryNode(ctx context.Context, state *ChatState) (*ChatState, error) {
	// 使用之前的工具函数读取 DAO
	// 注意：这里我们直接操作 State
	msgs, err := dao.QueryMessagesBySessionId(ctx, state.SessionId)
	if err != nil {
		return nil, fmt.Errorf("load history failed: %w", err)
	}

	// 转换格式并赋值给 State
	state.History = ToEinoMessages(msgs)

	// 2. 将当前 Query 作为一条新消息，追加到 History 末尾
	// 这样顺序就锁定为：[旧历史..., User(当前提问)]
	state.History = append(state.History, schema.UserMessage(state.Query))
	return state, nil
}

// ----------------------------------------------------------------
// 节点 2: LLM Runner
// 职责：State -> Prompt Template -> Chat Model -> 填充 State.Response
// ----------------------------------------------------------------
// 这是一个“组合节点”的逻辑。
// 为了复用 Eino 的 Prompt 和 ChatModel 组件，我们在函数内部构建一个小 Chain，或者直接调用组件。
// “正统”做法是将 Prompt 和 Model 作为 Graph 的独立节点，
// 但由于它们的输入输出 (Map -> Message) 和我们的 State 流 (*ChatState) 不匹配，
// 我们需要在这里做一个 "Adapter" (适配器)。

// ----------------------------------------------------------------
// 节点 3: History Saver
// 职责：读取 State (SessionID, Query, Response) -> 存库 -> 返回 State
// ----------------------------------------------------------------
func saveHistoryNode(ctx context.Context, state *ChatState) (*ChatState, error) {
	if state.Response == nil {
		return state, nil // 没生成结果就不存
	}

	// 组装要保存的消息
	var newMsgs []oa.ChatCompletionMessage
	for _, message := range state.History {
		newMsgs = append(newMsgs, ToOpenAIMessages(message))
	}

	newMsgs = append(newMsgs, oa.ChatCompletionMessage{Role: oa.ChatMessageRoleUser, Content: state.Query})
	newMsgs = append(newMsgs, ToOpenAIMessages(state.Response))

	// 调用 DAO 落库
	// 这是一个阻塞操作，确保数据一致性
	if err := dao.UpdateMessages(ctx, state.SessionId, newMsgs); err != nil {
		// 在这里，我们可以选择 log 错误但不中断流程，或者返回 error
		fmt.Printf("Warning: failed to save history: %v\n", err)
	}
	return state, nil
}

func retrieverNode(ctx context.Context, r retriever.Retriever, state *ChatState) (*ChatState, error) {
	documents, err := r.Retrieve(ctx, state.Query)
	if err != nil {
		return nil, err
	}

	state.Documents = documents
	return state, nil
}

func NewEinoChatAgent(ctx context.Context, config openai.ChatModelConfig) (*EinoChatAgent, error) {
	// 1. 创建 Graph 容器
	// 泛型明确指定了输入输出都是 *ChatState
	graph := compose.NewGraph[*ChatState, *ChatState]()

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

	// Node: LLM (这里我们将 build 逻辑其实应该提出来，避免每次 Invoke 都 NewModel)
	// 为了性能，我们先在外面初始化好 Model 和 Template
	// 1. 初始化组件 (生产环境建议在 build 时初始化一次，为了代码清晰这里仅做演示)
	// 实际项目中，chatModel 和 tmpl 应该在 NewEinoAgent 时创建好并闭包进来
	// 1. 定义 Chat Model (复用你的配置)
	chatModel, err := openai.NewChatModel(ctx, &config)
	if err != nil {
		return nil, err
	}

	info, err := kbTool.Info(ctx)
	if err != nil {
		return nil, err
	}
	client.ToolsInfo = append(client.ToolsInfo, info)
	if err = chatModel.BindTools(client.ToolsInfo); err != nil {
		return nil, err
	}

	// 2. 定义 Prompt 模版
	// 这里我们做一个简单的 Chat Template，支持 System Prompt 和 User Input
	// Placeholder 语法：{variable_name}
	tmpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个专业、高效、多功能的问题解决引擎。你的主要职责是利用你被赋予的外部工具和可参考的上下文来准确、简洁地回答用户的问题并完成指令。"), // 👈 新增 context 槽位
		schema.MessagesPlaceholder("messages", false)) // 修复多轮对话的bug

	lLMRunnerNode := func(ctx context.Context, input *ChatState) (*ChatState, error) {
		// 2. 准备组件需要的输入 (Map)
		inputMap := map[string]any{
			"messages": input.History,
		}

		// 3. 执行 Template -> Messages
		msgs, err := tmpl.Format(ctx, inputMap)
		if err != nil {
			return nil, err
		}

		// 4. 执行 Model -> Message
		resp, err := chatModel.Generate(ctx, msgs)
		if err != nil {
			return nil, err
		}

		// 5. 写入 State
		input.Response = resp
		return input, nil
	}

	var newTools []tool.BaseTool
	newTools = append(newTools, client.EinoTools...)
	newTools = append(newTools, kbTool)
	// 3. 创建 ToolsNode (Eino 提供的标准工具执行节点)
	// 它会自动处理：接收 ToolCalls -> 找到对应 Tool -> 执行 -> 返回 ToolMessage
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{Tools: newTools}) // 把 MCP 工具列表塞进去
	if err != nil {
		return nil, err
	}

	// 2. 添加节点 (AddNode)
	// 使用 LambdaNode 将我们的函数包装成 Graph 节点
	// Node: Loader
	if err := graph.AddLambdaNode("loader", compose.InvokableLambda(loadHistoryNode)); err != nil {
		return nil, err
	}

	if err = graph.AddLambdaNode("llm_runner", compose.InvokableLambda(lLMRunnerNode)); err != nil {
		return nil, err
	}

	if err = graph.AddLambdaNode("saver", compose.InvokableLambda(saveHistoryNode)); err != nil {
		return nil, err
	}

	// 🔥 添加 ToolsNode
	// 注意：ToolsNode 的标准输入是 []*schema.Message，输出也是 []*schema.Message
	// 但我们的 Graph 流动的是 *ChatState。
	// 所以我们需要把 ToolsNode 包装一下，适配 State
	_ = graph.AddLambdaNode("tools", compose.InvokableLambda(func(ctx context.Context, state *ChatState) (*ChatState, error) {
		// 1. 这里的输入应该是 LLM 刚刚生成的带 ToolCalls 的那条消息
		// 也就是 state.Response
		inputMsg := state.Response

		// 2. 调用 ToolsNode
		// Eino 的 ToolsNode.Invoke 需要包含 ToolCalls 的消息作为 input
		resMsgs, err := toolsNode.Invoke(ctx, inputMsg)
		if err != nil {
			return nil, err
		}

		// 3. 更新 State
		// resMsgs 包含了执行结果 (ToolMessage)
		// 我们先要把 LLM 的 ToolCalls 消息存入历史
		state.History = append(state.History, state.Response)

		// 再把工具执行结果存入历史
		for _, m := range resMsgs {
			state.History = append(state.History, m)
		}

		// 清空 Response 以便下一轮生成
		state.Response = nil
		return state, nil
	}))

	// 3. 定义边 (AddEdge) - 决定执行顺序
	// START -> loader -> llm_runner -> saver -> END
	_ = graph.AddEdge(compose.START, "loader")
	_ = graph.AddEdge("loader", "llm_runner")
	// 闭环：工具执行完 -> 回到 LLM
	_ = graph.AddEdge("tools", "llm_runner")
	// 分支：有 ToolCalls 吗？
	_ = graph.AddBranch("llm_runner", compose.NewGraphBranch(func(ctx context.Context, state *ChatState) (string, error) {
		if len(state.Response.ToolCalls) > 0 {
			return "tools", nil // -> 去执行工具
		}
		return "saver", nil // -> 结束
	}, map[string]bool{"saver": true, "tools": true}))
	//_ = graph.AddEdge("llm_runner", "saver")
	_ = graph.AddEdge("saver", compose.END)

	runner, err := graph.Compile(ctx, compose.WithMaxRunSteps(20))
	if err != nil {
		return nil, err
	}

	return &EinoChatAgent{Runnable: runner}, nil
}
