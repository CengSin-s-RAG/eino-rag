package api

import (
	"agent.article.fp/agent"
	"github.com/cloudwego/eino/schema"
	"github.com/labstack/echo/v4"
)

// EinoChatAgentHandler 结构体用于持有 Agent 实例
type EinoChatAgentHandler struct {
	Agent *agent.EinoChatAgent
}

func NewEinoChatAgentHandler(agent *agent.EinoChatAgent) *EinoChatAgentHandler {
	return &EinoChatAgentHandler{
		Agent: agent,
	}
}

func (h *EinoChatAgentHandler) HandleQuery(c echo.Context) (*schema.StreamReader[*schema.Message], error) {
	ctx := c.Request().Context()

	// 1. 绑定请求
	// 复用了 workflow.ChatReq，你可以根据需要定义新的 Request 结构
	var req ChatReq
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	// 2. 调用 Eino Agent
	// 注意：这里是同步调用，会阻塞直到 LLM 返回。
	// 对于超长推理，生产环境通常还是会结合 Temporal 或 SSE (Server-Sent Events)
	return h.Agent.Run(ctx, req.Question)
}
