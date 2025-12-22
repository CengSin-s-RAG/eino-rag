package agent

import (
	"agent.article.fp/client"
	"agent.article.fp/dao"
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	oa "github.com/sashabaranov/go-openai"
	"strings"
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

	// Node: LLM (这里我们将 build 逻辑其实应该提出来，避免每次 Invoke 都 NewModel)
	// 为了性能，我们先在外面初始化好 Model 和 Template
	// 1. 初始化组件 (生产环境建议在 build 时初始化一次，为了代码清晰这里仅做演示)
	// 实际项目中，chatModel 和 tmpl 应该在 NewEinoAgent 时创建好并闭包进来
	// 1. 定义 Chat Model (复用你的配置)
	chatModel, err := openai.NewChatModel(ctx, &config)
	if err != nil {
		return nil, err
	}

	// 2. 定义 Prompt 模版
	// 这里我们做一个简单的 Chat Template，支持 System Prompt 和 User Input
	// Placeholder 语法：{variable_name}
	tmpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个金融助手。参考以下上下文回答问题：\n\n{context}"), // 👈 新增 context 槽位
		schema.MessagesPlaceholder("history", false),
		schema.UserMessage("{query}"))

	lLMRunnerNode := func(ctx context.Context, input *ChatState) (*ChatState, error) {
		var sb strings.Builder
		for _, doc := range input.Documents {
			sb.WriteString(doc.Content + "\n---\n")
		}

		// 2. 准备组件需要的输入 (Map)
		inputMap := map[string]any{
			"context": sb.String(),
			"history": input.History,
			"query":   input.Query,
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

	embedder, err := NewEinoEmbedder(ctx)
	if err != nil {
		return nil, err
	}

	retrv, err := NewEinoRetriever(ctx, client.Qdrant, embedder)
	if err != nil {
		return nil, err
	}

	retrieverHandle := func(ctx context.Context, state *ChatState) (*ChatState, error) {
		return retrieverNode(ctx, retrv, state)
	}

	// 2. 添加节点 (AddNode)
	// 使用 LambdaNode 将我们的函数包装成 Graph 节点
	// Node: Loader
	if err := graph.AddLambdaNode("loader", compose.InvokableLambda(loadHistoryNode)); err != nil {
		return nil, err
	}

	if err = graph.AddLambdaNode("retriever", compose.InvokableLambda(retrieverHandle)); err != nil {
		return nil, err
	}

	if err = graph.AddLambdaNode("llm_runner", compose.InvokableLambda(lLMRunnerNode)); err != nil {
		return nil, err
	}

	if err = graph.AddLambdaNode("saver", compose.InvokableLambda(saveHistoryNode)); err != nil {
		return nil, err
	}

	// 3. 定义边 (AddEdge) - 决定执行顺序
	// START -> loader -> llm_runner -> saver -> END
	_ = graph.AddEdge(compose.START, "loader")
	_ = graph.AddEdge("loader", "retriever")
	_ = graph.AddEdge("retriever", "llm_runner")
	_ = graph.AddEdge("llm_runner", "saver")
	_ = graph.AddEdge("saver", compose.END)

	runner, err := graph.Compile(ctx)
	if err != nil {
		return nil, err
	}

	return &EinoChatAgent{Runnable: runner}, nil
}
