// Package store 实现 Redis 会话实现
package store

import (
	"context"
)

type Message struct{}

type Store interface {
	Recent(ctx context.Context, sessionID string, limit int) ([]Message, error)
	Append(ctx context.Context, sessionID string, messages ...Message) error
}

func New() (Store, error) {
	return nil, nil
}
