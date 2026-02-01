package component

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// Reflect summarizes the session and appends to the research log.
func Reflect(ctx context.Context, model *openai.ChatModel, history []*schema.Message) error {
	if len(history) == 0 {
		return nil
	}

	prompt := "Please summarize the key findings and progress from the following conversation for a research log. Focus on factual information and decisions made.\n\nConversation:\n"
	for _, msg := range history {
		prompt += fmt.Sprintf("[%s]: %s\n", msg.Role, msg.Content)
	}

	summary, err := model.Generate(ctx, []*schema.Message{
		schema.UserMessage(prompt),
	})
	if err != nil {
		return err
	}

	logEntry := fmt.Sprintf("\n## %s\n%s\n", time.Now().Format("2006-01-02 15:04:05"), summary.Content)

	f, err := os.OpenFile("memory/RESEARCH_LOGS.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(logEntry); err != nil {
		return err
	}

	return nil
}
