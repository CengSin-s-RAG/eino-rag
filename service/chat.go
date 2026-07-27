package service

import (
	"agent.article.fp/runtime"
	"agent.article.fp/store"
	"agent.article.fp/tools"
	"context"
)

type ChatService struct {
	primary  runtime.Provider
	fallback []runtime.Provider
	tools    []tools.Tool
	store    store.Store
}

type ChatRequest struct {
	Message string `json:"message"`
	Session string `json:"session"`
}

type ChatResponse struct {
	Session      string `json:"session"`
	ReplyMessage string `json:"reply_message"`
}

type Handle func(ctx context.Context, req ChatRequest) (*ChatResponse, error)

func New() (*ChatService, error) {
	primary, fallback, err := runtime.New()
	if err != nil {
		return nil, err
	}

	zTools, err := tools.New()
	if err != nil {
		return nil, err
	}

	zStore, err := store.New()
	if err != nil {
		return nil, err
	}

	return &ChatService{
		primary:  primary,
		fallback: fallback,
		tools:    zTools,
		store:    zStore,
	}, nil
}
