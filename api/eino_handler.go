package api

import (
	"agent.article.fp/agent"
	"agent.article.fp/agent/component"
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

// EinoChatAgentHandler 结构体用于持有 Chat 实例
type EinoChatAgentHandler struct {
	Chat *agent.EinoChatAgent
}

func NewEinoChatAgentHandler(chat *agent.EinoChatAgent) *EinoChatAgentHandler {
	return &EinoChatAgentHandler{
		Chat: chat,
	}
}

const ContextWindowSize = 20

func (h *EinoChatAgentHandler) HandleRewrite(c echo.Context) error {
	ctx := c.Request().Context()
	var req component.RewriteReq
	if err := c.Bind(&req); err != nil {
		return err
	}

	userMsg := fmt.Sprintf("【历史对话记录】\n %s \n【当前用户问题】%s", req.GetHistory(), req.Query)

	msg, err := client.QueryRewriteModel.Generate(ctx, append([]*schema.Message{schema.SystemMessage(util.QueryRewritePrompt)}, schema.UserMessage(userMsg)))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, msg.Content)
}

func (h *EinoChatAgentHandler) HandleRerank(c echo.Context) error {
	ctx := c.Request().Context()

	var req component.RerankReq
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// todo 基于BM25算法对文档进行重排序 qdrant go-sdk目前暂未支持

	// 让大模型对文档进行排序
	userMsg := fmt.Sprintf("【用户问题】\n %s \n【候选文档列表】", req.Question)
	for _, document := range req.Documents {
		userMsg += fmt.Sprintf("\n%s\n", document)
	}

	input := append([]*schema.Message{schema.SystemMessage(util.RerankPrompt)}, schema.UserMessage(userMsg))

	message, err := client.RerankModel.Generate(ctx, input)
	if err != nil {
		return err
	}

	// 创建解析器
	parser := schema.NewMessageJSONParser[component.RerankState](&schema.MessageJSONParseConfig{
		ParseFrom: schema.MessageParseFromContent,
	})
	items, err := parser.Parse(ctx, message)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// todo 需要在接口中对结果进行分数纬度的降序排序，目前是在调用接口的逻辑中处理的

	// todo 根据文档来源、发布时间进行排序
	return c.JSON(http.StatusOK, items)
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

	// todo 输入消毒 识别用户输入是否危险
	if err := h.CheckInput(req.Question); err != nil {
		return c.JSON(http.StatusOK, err.Error())
	}

	store := memory.NewRedisStore(client.Redis)
	var messages []*schema.Message
	messages, err := store.GetRecentMessages(ctx, input.SessionId, -ContextWindowSize, -1)
	if err != nil {
		return err
	}
	userMessage := schema.UserMessage(input.Query)
	messages = append(messages, userMessage)

	chatContent := append([]*schema.Message{schema.SystemMessage(util.GetSystemPrompt())}, messages...)
	message, err := h.Chat.Runner.Generate(ctx, chatContent, einoAgent.WithComposeOptions(compose.WithCallbacks(&util.SimpleLogger{})))
	if err != nil {
		return err
	}

	// todo 输出审查 识别并拦截仇恨言论、偏见内容、毒性语言或违反公司政策的信息
	if err = h.CheckOutput(message.Content); err != nil {
		return c.JSON(http.StatusOK, err.Error())
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

func (h *EinoChatAgentHandler) CheckInput(question string) error {
	// todo 增加输入内容的校验
	return nil
}

func (h *EinoChatAgentHandler) CheckOutput(content string) error {
	// todo 增加输出内容的校验
	return nil
}
