package component

import (
	"context"
	"fmt"
	"log"
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

	prompt := "You are a research assistant. Summarize the following conversation into a concise bulleted list of 3-5 key takeaways for a research log. Focus on analytical insights, decisions, and factual findings.\n\n"
	prompt += "CRITICAL: If no significant research was discussed, state 'No significant research findings.' DO NOT return an empty response.\n\n"
	prompt += "Conversation Context:\n"

	var contentFound bool
	for _, msg := range history {
		if msg.Content == "" || msg.Role == schema.System {
			continue
		}
		prompt += fmt.Sprintf("[%s]: %s\n", msg.Role, msg.Content)
		contentFound = true
	}

	if !contentFound {
		return nil
	}

	// Logging the outgoing prompt for debugging
	log.Printf("[Reflection] Outgoing Prompt: %s", prompt)

	summary, err := model.Generate(ctx, []*schema.Message{
		schema.SystemMessage("You are a professional research summarizer. Always provide a text response. Never return an empty string."),
		schema.UserMessage(prompt),
	})
	if err != nil {
		return fmt.Errorf("failed to generate summary: %v", err)
	}

	// Logging the raw response from LLM
	log.Printf("[Reflection] Raw LLM Response: '%s'", summary.Content)

	content := summary.Content
	if content == "" {
		log.Printf("[Reflection] Warning: LLM returned empty content. Recording placeholder.")
		content = "Error: LLM returned empty summary for this session."
	}

	logEntry := fmt.Sprintf("\n## %s\n%s\n", time.Now().Format("2006-01-02 15:04:05"), content)

	f, err := os.OpenFile("memory/RESEARCH_LOGS.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open research logs: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(logEntry); err != nil {
		return fmt.Errorf("failed to write to research logs: %v", err)
	}

	return nil
}
