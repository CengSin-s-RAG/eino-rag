package component

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// GenerateMorningReport takes raw news items and synthesizes them into a macro report
// based on the Fu Peng persona's analytical logic.
func GenerateMorningReport(ctx context.Context, model *openai.ChatModel, news []string) (string, error) {
	if len(news) == 0 {
		return "No news data available for the morning report.", nil
	}

	newsContext := strings.Join(news, "\n- ")

	prompt := fmt.Sprintf(`You are Fu Peng's Research Assistant. Your goal is to analyze the following overnight global macro news and provide a strategic morning report.

Your analysis MUST strictly follow these principles:
1. Focus on **Interest Rate Parity (利率平价)** and its impact on global capital flows.
2. Identify **Asset Scarcity (资产荒)** signals in various markets.
3. Observe **Global Capital Flows (资本流动)** shifts and their drivers.
4. Maintain a professional, insightful, and concise tone.

Raw News Data:
- %s

Provide the report in Chinese. Ensure it is detailed (at least 200 characters) and clearly identifies the core macro logic.`, newsContext)

	resp, err := model.Generate(ctx, []*schema.Message{
		schema.SystemMessage("You are a professional macro researcher specializing in global capital flows and yield analysis."),
		schema.UserMessage(prompt),
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate morning report: %w", err)
	}

	return resp.Content, nil
}
