package api

import (
	"agent.article.fp/agent"
	"agent.article.fp/client"
	"agent.article.fp/memory"
	"agent.article.fp/util"
	"context"
	"fmt"
	"github.com/cloudwego/eino/compose"
	einoAgent "github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/schema"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"time"
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

const ContextWindowSize = 20

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
	messages, err := store.GetRecentMessages(ctx, input.SessionId, ContextWindowSize)
	if err != nil {
		return err
	}
	userMessage := schema.UserMessage(input.Query)
	messages = append(messages, userMessage)

	message, err := h.Agent.Runner.Generate(ctx, messages, einoAgent.WithComposeOptions(compose.WithCallbacks(&util.SimpleLogger{})))
	if err != nil {
		return err
	}

	go func() {
		aCtx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
		defer cancelFunc()
		if err = store.AddMessage(aCtx, input.SessionId, userMessage); err != nil {
			log.Println(fmt.Sprintf("store add message error: %v", err))
		}
		if err = store.AddMessage(aCtx, input.SessionId, message); err != nil {
			log.Println(fmt.Sprintf("store add message error: %v", err))
		}
	}()

	return c.JSON(http.StatusOK, &ChatResp{SessionID: input.SessionId, ReplyMessage: message.Content})
}
