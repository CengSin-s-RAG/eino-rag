// Package openai implements the OpenAI-compatible chat-completions protocol.
package openai

import (
	"agent.article.fp/config"
	"agent.article.fp/runtime"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Provider struct {
	baseURL     string
	apiKey      string
	model       string
	temperature float64
	client      *http.Client
}

func New(cfg config.ProviderConfig) (*Provider, error) {
	if cfg.Type != "" && cfg.Type != "openai-compatible" {
		return nil, fmt.Errorf("unsupported provider type %q", cfg.Type)
	}
	if strings.TrimSpace(cfg.BaseURL) == "" || strings.TrimSpace(cfg.Model) == "" {
		return nil, fmt.Errorf("provider baseURL and model are required")
	}
	apiKey := os.Getenv(cfg.APIKeyEnv)
	if cfg.APIKeyEnv == "" || apiKey == "" {
		return nil, fmt.Errorf("provider API key environment variable %q is empty", cfg.APIKeyEnv)
	}
	return &Provider{baseURL: strings.TrimRight(cfg.BaseURL, "/"), apiKey: apiKey, model: cfg.Model, temperature: cfg.Temperature, client: http.DefaultClient}, nil
}

func (p *Provider) Complete(ctx context.Context, request runtime.CompletionRequest) (runtime.Completion, error) {
	payload := chatRequest{Model: p.model, Messages: encodeMessages(request.Messages), Tools: encodeTools(request.Tools), Temperature: p.temperature}
	if len(payload.Tools) > 0 {
		payload.ToolChoice = "auto"
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return runtime.Completion{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return runtime.Completion{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return runtime.Completion{}, err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return runtime.Completion{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return runtime.Completion{}, fmt.Errorf("chat completion returned %s: %s", resp.Status, strings.TrimSpace(string(responseBody)))
	}
	var decoded chatResponse
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return runtime.Completion{}, fmt.Errorf("decode chat completion: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return runtime.Completion{}, fmt.Errorf("chat completion returned no choices")
	}
	return runtime.Completion{Message: decodeMessage(decoded.Choices[0].Message)}, nil
}

func (p *Provider) Close() error { return nil }

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []wireMessage `json:"messages"`
	Tools       []wireTool    `json:"tools,omitempty"`
	ToolChoice  string        `json:"tool_choice,omitempty"`
	Temperature float64       `json:"temperature"`
}

type chatResponse struct {
	Choices []struct {
		Message wireMessage `json:"message"`
	} `json:"choices"`
}

// These wire structures intentionally mirror ChatCompletionNewParams from
// openai-go. In particular, function tools and tool calls are nested under a
// "function" object; runtime types stay provider-neutral.
type wireMessage struct {
	Role       runtime.Role   `json:"role"`
	Content    *string        `json:"content"`
	ToolCalls  []wireToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
}

type wireTool struct {
	Type     string       `json:"type"`
	Function wireFunction `json:"function"`
}

type wireFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
	Arguments   string         `json:"arguments,omitempty"`
}

type wireToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function wireFunction `json:"function"`
}

func encodeMessages(messages []runtime.Message) []wireMessage {
	encoded := make([]wireMessage, 0, len(messages))
	for _, message := range messages {
		item := wireMessage{Role: message.Role, ToolCallID: message.ToolCallID}
		// OpenAI-compatible APIs expect content=null for an assistant tool-call
		// message. Text messages keep their content exactly as supplied.
		if message.Content != "" || len(message.ToolCalls) == 0 {
			content := message.Content
			item.Content = &content
		}
		for _, call := range message.ToolCalls {
			item.ToolCalls = append(item.ToolCalls, wireToolCall{ID: call.ID, Type: "function", Function: wireFunction{Name: call.Name, Arguments: call.Arguments}})
		}
		encoded = append(encoded, item)
	}
	return encoded
}

func encodeTools(tools []runtime.ToolDefinition) []wireTool {
	encoded := make([]wireTool, 0, len(tools))
	for _, tool := range tools {
		encoded = append(encoded, wireTool{Type: "function", Function: wireFunction{Name: tool.Name, Description: tool.Description, Parameters: tool.Parameters}})
	}
	return encoded
}

func decodeMessage(message wireMessage) runtime.Message {
	decoded := runtime.Message{Role: message.Role, ToolCallID: message.ToolCallID}
	if message.Content != nil {
		decoded.Content = *message.Content
	}
	for _, call := range message.ToolCalls {
		decoded.ToolCalls = append(decoded.ToolCalls, runtime.ToolCall{ID: call.ID, Name: call.Function.Name, Arguments: call.Function.Arguments})
	}
	return decoded
}
