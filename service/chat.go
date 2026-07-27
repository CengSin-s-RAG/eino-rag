// Package service owns the agent loop and is independent of HTTP/Echo.
package service

import (
	"agent.article.fp/runtime"
	"agent.article.fp/store"
	"agent.article.fp/tools"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type ChatRequest struct {
	SessionID string `json:"session_id"`
	Question  string `json:"question"`
}

type ChatResponse struct {
	SessionID    string `json:"session_id"`
	ReplyMessage string `json:"reply_message"`
}

// Handle is retained as a small transport extension seam. New HTTP routes
// should normally call a concrete service method directly.
type Handle func(context.Context, ChatRequest) (*ChatResponse, error)

type ChatService struct {
	primary  runtime.Provider
	fallback []runtime.Provider
	tools    *tools.Registry
	store    store.Store
	system   string
	history  int
	maxSteps int
}

func NewChatService(primary runtime.Provider, fallback []runtime.Provider, registry *tools.Registry, chatStore store.Store, systemPrompt string, history, maxSteps int) *ChatService {
	if history <= 0 {
		history = 20
	}
	if maxSteps <= 0 {
		maxSteps = 15
	}
	return &ChatService{primary: primary, fallback: fallback, tools: registry, store: chatStore, system: systemPrompt, history: history, maxSteps: maxSteps}
}

func (s *ChatService) Chat(ctx context.Context, request ChatRequest) (ChatResponse, error) {
	if request.Question == "" {
		return ChatResponse{}, fmt.Errorf("question is required")
	}
	if request.SessionID == "" {
		request.SessionID = uuid.NewString()
	}
	history, err := s.store.Recent(ctx, request.SessionID, s.history)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("read chat history: %w", err)
	}
	messages := make([]runtime.Message, 0, len(history)+2)
	messages = append(messages, runtime.Message{Role: runtime.RoleSystem, Content: s.system})
	messages = append(messages, history...)
	userMessage := runtime.Message{Role: runtime.RoleUser, Content: request.Question}
	messages = append(messages, userMessage)

	completion, err := s.complete(ctx, runtime.CompletionRequest{Messages: messages, Tools: s.definitions()})
	if err != nil {
		return ChatResponse{}, err
	}
	for step := 0; len(completion.Message.ToolCalls) > 0; step++ {
		if step >= s.maxSteps {
			return ChatResponse{}, fmt.Errorf("agent reached tool-call limit (%d)", s.maxSteps)
		}
		messages = append(messages, completion.Message)
		for _, call := range completion.Message.ToolCalls {
			result := s.tools.Run(ctx, call.Name, []byte(call.Arguments))
			messages = append(messages, runtime.Message{Role: runtime.RoleTool, ToolCallID: call.ID, Content: result.Content})
		}
		completion, err = s.complete(ctx, runtime.CompletionRequest{Messages: messages, Tools: s.definitions()})
		if err != nil {
			return ChatResponse{}, err
		}
	}
	if err := s.store.Append(ctx, request.SessionID, userMessage, completion.Message); err != nil {
		return ChatResponse{}, fmt.Errorf("write chat history: %w", err)
	}
	return ChatResponse{SessionID: request.SessionID, ReplyMessage: completion.Message.Content}, nil
}

func (s *ChatService) complete(ctx context.Context, request runtime.CompletionRequest) (runtime.Completion, error) {
	completion, err := s.primary.Complete(ctx, request)
	if err == nil {
		return completion, nil
	}
	primaryErr := err
	for _, fallback := range s.fallback {
		completion, err = fallback.Complete(ctx, request)
		if err == nil {
			return completion, nil
		}
	}
	return runtime.Completion{}, fmt.Errorf("all providers failed (primary: %w)", primaryErr)
}

func (s *ChatService) definitions() []runtime.ToolDefinition {
	if s.tools == nil {
		return nil
	}
	definitions := s.tools.Definitions()
	result := make([]runtime.ToolDefinition, 0, len(definitions))
	for _, definition := range definitions {
		result = append(result, runtime.ToolDefinition{Name: definition.Name, Description: definition.Description, Parameters: definition.Parameters})
	}
	return result
}
