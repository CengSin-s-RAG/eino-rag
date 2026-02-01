package tests

import (
	"context"
	"os"
	"strings"
	"testing"

	"agent.article.fp/agent/component"
	"github.com/cloudwego/eino-ext/components/model/openai"
)

// TestMorningReportGeneration ensures that the proactive inspection logic
// can synthesize raw news into a strategy report following Fu Peng's logic.
func TestMorningReportGeneration(t *testing.T) {
	ctx := context.Background()

	// 1. Initialize the Chat Model for the test
	// We use OpenRouter as configured in the main project
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		// Fallback for local dev if not in env but in config
		// For this test, we expect the user to provide it or have it in the env
		t.Skip("Skipping test: OPENROUTER_API_KEY not set")
	}

	conf := openai.ChatModelConfig{
		APIKey:  apiKey,
		BaseURL: "https://openrouter.ai/api/v1",
		Model:   "google/gemini-2.0-flash-001", // Use a fast/cheap model for testing
	}
	model, err := openai.NewChatModel(ctx, &conf)
	if err != nil {
		t.Fatalf("Failed to create chat model: %v", err)
	}

	// 2. Mock Raw News Data
	mockNews := []string{
		"Fed signals 'higher for longer' interest rates as inflation remains sticky.",
		"PBoC maintains steady rates, focus shifts to liquidity injection.",
		"Global capital flows show shift towards stable yield assets in the US.",
		"Scarcity of high-quality corporate bonds noted in European markets.",
	}

	// 3. Call the implemented logic
	report, err := component.GenerateMorningReport(ctx, model, mockNews)
	if err != nil {
		t.Fatalf("Failed to generate morning report: %v", err)
	}

	// 4. Assertions based on "Boss's" requirements
	t.Logf("Generated Report: \n%s", report)

	if len(report) < 200 {
		t.Errorf("Report is too short (%d characters), expected at least 200", len(report))
	}

	// Logic Checks: Must contain core philosophy keywords (Chinese expected)
	requiredKeywords := []string{"资产荒", "利率平价", "资本流动"}
	foundKeyword := false
	for _, kw := range requiredKeywords {
		if strings.Contains(report, kw) {
			foundKeyword = true
			break
		}
	}

	if !foundKeyword {
		t.Errorf("Report does not seem to follow Fu Peng's core analytical logic (missing keywords like %v)", requiredKeywords)
	}
}
