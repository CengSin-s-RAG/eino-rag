package agent

import (
	"github.com/cloudwego/eino/schema"
)

type ChatState struct {
	SessionId string
	Query     string
	// 中间状态字段
	History []*schema.Message // 由 Loader 填充
	// 输出字段
	Response *schema.Message // 由 LLM 填充
}
