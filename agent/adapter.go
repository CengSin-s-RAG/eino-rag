package agent

import (
	"github.com/cloudwego/eino/schema"
	"github.com/sashabaranov/go-openai"
)

// ToEinoMessages 将 DAO 的 OpenAI 消息转换为 Eino 消息
func ToEinoMessages(oaMsgs []openai.ChatCompletionMessage) []*schema.Message {
	var einoMsgs []*schema.Message
	for _, m := range oaMsgs {
		switch m.Role {
		case openai.ChatMessageRoleSystem:
			einoMsgs = append(einoMsgs, schema.SystemMessage(m.Content))
		case openai.ChatMessageRoleUser:
			einoMsgs = append(einoMsgs, schema.UserMessage(m.Content))
		case openai.ChatMessageRoleAssistant:
			einoMsgs = append(einoMsgs, schema.AssistantMessage(m.Content, nil)) // ToolCalls 暂时置空
			// TODO: 处理 Tool 消息
		}
	}
	return einoMsgs
}

// ToOpenAIMessages 将 Eino 消息转回 OpenAI 格式以保存
func ToOpenAIMessages(einoMsg *schema.Message) openai.ChatCompletionMessage {
	role := openai.ChatMessageRoleUser
	switch einoMsg.Role {
	case schema.System:
		role = openai.ChatMessageRoleSystem
	case schema.Assistant:
		role = openai.ChatMessageRoleAssistant
	case schema.User:
		role = openai.ChatMessageRoleUser
	}
	return openai.ChatCompletionMessage{
		Role:    role,
		Content: einoMsg.Content,
	}
}
