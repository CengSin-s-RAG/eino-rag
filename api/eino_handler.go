package api

import (
	"agent.article.fp/agent"
	"agent.article.fp/client"
	"agent.article.fp/memory"
	"github.com/cloudwego/eino/schema"
	"github.com/labstack/echo/v4"
	"net/http"
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

func (h *EinoChatAgentHandler) HandleQuery(c echo.Context) error {
	ctx := c.Request().Context()

	// 1. 绑定请求
	// 复用了 workflow.ChatReq，你可以根据需要定义新的 Request 结构
	var req ChatReq
	if err := c.Bind(&req); err != nil {
		return err
	}

	input := agent.ChatState{
		SessionId: req.SessionID,
		Query:     req.Question,
	}

	store := memory.NewRedisStore(client.Redis)

	var messages []*schema.Message
	messages, err := store.Read(ctx, input.SessionId)
	if err != nil {
		return err
	}

	messages = append(messages, schema.UserMessage(input.Query))

	message, err := h.Agent.Runner.Generate(ctx, messages)
	if err != nil {
		return err
	}

	if err = store.Write(ctx, input.SessionId, append(messages, message)); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &ChatResp{SessionID: input.SessionId, ReplyMessage: message.Content})
}
