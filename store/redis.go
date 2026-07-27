package store

import (
	"agent.article.fp/config"
	"agent.article.fp/runtime"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Redis struct{ client *redis.Client }

func NewRedis(cfg config.RedisConfig) *Redis {
	return &Redis{client: redis.NewClient(&redis.Options{Addr: cfg.Addr, Password: cfg.Password, DB: cfg.DB})}
}

func (s *Redis) Recent(ctx context.Context, sessionID string, limit int) ([]runtime.Message, error) {
	if limit <= 0 {
		return nil, nil
	}
	items, err := s.client.LRange(ctx, key(sessionID), int64(-limit), -1).Result()
	if err != nil {
		return nil, err
	}
	messages := make([]runtime.Message, 0, len(items))
	for _, item := range items {
		var message runtime.Message
		if err := json.Unmarshal([]byte(item), &message); err != nil {
			return nil, fmt.Errorf("decode message: %w", err)
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func (s *Redis) Append(ctx context.Context, sessionID string, messages ...runtime.Message) error {
	if len(messages) == 0 {
		return nil
	}
	values := make([]any, 0, len(messages))
	for _, message := range messages {
		encoded, err := json.Marshal(message)
		if err != nil {
			return err
		}
		values = append(values, encoded)
	}
	now := time.Now().Unix()
	pipe := s.client.TxPipeline()
	pipe.RPush(ctx, key(sessionID), values...)
	pipe.HSetNX(ctx, metaKey(sessionID), "created_at", now)
	pipe.HSet(ctx, metaKey(sessionID), "updated_at", now)
	if title := firstUserMessage(messages); title != "" {
		pipe.HSetNX(ctx, metaKey(sessionID), "title", title)
	}
	pipe.ZAdd(ctx, sessionsKey(), redis.Z{Score: float64(now), Member: sessionID})
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Redis) CreateSession(ctx context.Context) (Session, error) {
	now := time.Now().Unix()
	session := Session{ID: uuid.NewString(), Title: "新对话", CreatedAt: now, UpdatedAt: now}
	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, metaKey(session.ID), map[string]any{"title": session.Title, "created_at": now, "updated_at": now})
	pipe.ZAdd(ctx, sessionsKey(), redis.Z{Score: float64(now), Member: session.ID})
	_, err := pipe.Exec(ctx)
	return session, err
}

func (s *Redis) ListSessions(ctx context.Context, offset, limit int64) ([]Session, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	ids, err := s.client.ZRevRange(ctx, sessionsKey(), offset, offset+limit-1).Result()
	if err != nil {
		return nil, err
	}
	result := make([]Session, 0, len(ids))
	for _, id := range ids {
		meta, err := s.client.HGetAll(ctx, metaKey(id)).Result()
		if err != nil {
			return nil, err
		}
		session := Session{ID: id, Title: meta["title"]}
		fmt.Sscan(meta["created_at"], &session.CreatedAt)
		fmt.Sscan(meta["updated_at"], &session.UpdatedAt)
		if session.Title == "" {
			session.Title = "对话 " + id[:min(8, len(id))]
		}
		result = append(result, session)
	}
	return result, nil
}

func (s *Redis) Close() error { return s.client.Close() }

func key(sessionID string) string     { return "chatHistory:" + sessionID }
func metaKey(sessionID string) string { return "chatSession:" + sessionID }
func sessionsKey() string             { return "chatSessions" }
func firstUserMessage(messages []runtime.Message) string {
	for _, message := range messages {
		if message.Role == runtime.RoleUser && message.Content != "" {
			runes := []rune(message.Content)
			return strings.TrimSpace(string(runes[:min(40, len(runes))]))
		}
	}
	return ""
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
