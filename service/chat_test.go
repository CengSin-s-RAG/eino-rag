package service

import (
	"agent.article.fp/runtime"
	storepkg "agent.article.fp/store"
	"context"
	"errors"
	"testing"
)

type fakeProvider struct {
	response runtime.Completion
	err      error
	requests []runtime.CompletionRequest
}

func (p *fakeProvider) Complete(_ context.Context, request runtime.CompletionRequest) (runtime.Completion, error) {
	p.requests = append(p.requests, request)
	return p.response, p.err
}

type memoryStore struct{ messages []runtime.Message }

func (s *memoryStore) Recent(_ context.Context, _ string, limit int) ([]runtime.Message, error) {
	if len(s.messages) <= limit {
		return append([]runtime.Message(nil), s.messages...), nil
	}
	return append([]runtime.Message(nil), s.messages[len(s.messages)-limit:]...), nil
}

func (s *memoryStore) Append(_ context.Context, _ string, messages ...runtime.Message) error {
	s.messages = append(s.messages, messages...)
	return nil
}

func (*memoryStore) Close() error { return nil }

func (s *memoryStore) CreateSession(_ context.Context) (storepkg.Session, error) {
	return storepkg.Session{ID: "s1", Title: "新对话"}, nil
}

func (*memoryStore) ListSessions(_ context.Context, _, _ int64) ([]storepkg.Session, error) {
	return nil, nil
}

func TestChatPersistsTranscript(t *testing.T) {
	provider := &fakeProvider{response: runtime.Completion{Message: runtime.Message{Role: runtime.RoleAssistant, Content: "answer"}}}
	store := &memoryStore{messages: []runtime.Message{{Role: runtime.RoleUser, Content: "old question"}}}
	service := NewChatService(provider, nil, nil, store, "system", 20, 2)

	response, err := service.Chat(context.Background(), ChatRequest{SessionID: "s1", Question: "new question"})
	if err != nil {
		t.Fatal(err)
	}
	if response.ReplyMessage != "answer" || response.SessionID != "s1" {
		t.Fatalf("unexpected response: %#v", response)
	}
	if len(provider.requests) != 1 || len(provider.requests[0].Messages) != 3 {
		t.Fatalf("unexpected completion request: %#v", provider.requests)
	}
	if len(store.messages) != 3 || store.messages[1].Content != "new question" || store.messages[2].Content != "answer" {
		t.Fatalf("unexpected stored messages: %#v", store.messages)
	}
}

func TestChatUsesFallback(t *testing.T) {
	primary := &fakeProvider{err: errors.New("primary unavailable")}
	fallback := &fakeProvider{response: runtime.Completion{Message: runtime.Message{Role: runtime.RoleAssistant, Content: "fallback answer"}}}
	service := NewChatService(primary, []runtime.Provider{fallback}, nil, &memoryStore{}, "system", 20, 2)

	response, err := service.Chat(context.Background(), ChatRequest{SessionID: "s1", Question: "question"})
	if err != nil {
		t.Fatal(err)
	}
	if response.ReplyMessage != "fallback answer" {
		t.Fatalf("unexpected response: %#v", response)
	}
	if len(primary.requests) != 1 || len(fallback.requests) != 1 {
		t.Fatalf("providers were not called once: %d, %d", len(primary.requests), len(fallback.requests))
	}
}
