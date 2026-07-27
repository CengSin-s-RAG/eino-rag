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
	if cfg.APIKeyEnv == "" && apiKey == "" {
		return nil, fmt.Errorf("provider API key environment variable %q is empty", cfg.APIKeyEnv)
	}
	return &Provider{baseURL: strings.TrimRight(cfg.BaseURL, "/"), apiKey: apiKey, model: cfg.Model, temperature: cfg.Temperature, client: http.DefaultClient}, nil
}

func (p *Provider) Complete(ctx context.Context, request runtime.CompletionRequest) (runtime.Completion, error) {
	body, err := json.Marshal(chatRequest{Model: p.model, Messages: request.Messages, Tools: request.Tools, Temperature: p.temperature})
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
	return runtime.Completion{Message: decoded.Choices[0].Message}, nil
}

func (p *Provider) Close() error { return nil }

type chatRequest struct {
	Model       string                   `json:"model"`
	Messages    []runtime.Message        `json:"messages"`
	Tools       []runtime.ToolDefinition `json:"tools,omitempty"`
	Temperature float64                  `json:"temperature"`
}

type chatResponse struct {
	Choices []struct {
		Message runtime.Message `json:"message"`
	} `json:"choices"`
}
