// Package store persists the framework-neutral chat transcript.
package store

import (
	"agent.article.fp/runtime"
	"context"
)

type Store interface {
	Recent(ctx context.Context, sessionID string, limit int) ([]runtime.Message, error)
	Append(ctx context.Context, sessionID string, messages ...runtime.Message) error
	CreateSession(ctx context.Context) (Session, error)
	ListSessions(ctx context.Context, offset, limit int64) ([]Session, error)
	Close() error
}

type Session struct {
	ID        string `json:"session_id"`
	Title     string `json:"title"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}
