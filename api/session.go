package api

import (
	"agent.article.fp/client"
	"agent.article.fp/memory"
	"agent.article.fp/util"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Session struct {
}

func NewChatSession() *Session {
	return &Session{}
}

type SessionQuery struct {
	Cursor int64 `query:"cursor"`
	Limit  int64 `query:"limit"`
}

type SessionList struct {
	Sessions   []string `json:"sessions"`
	NextCursor int64    `json:"next_cursor"`
}

func (s *Session) List(c echo.Context) error {
	ctx := c.Request().Context()
	var query SessionQuery
	if err := c.Bind(&query); err != nil {
		return err
	}
	pageSize := query.Limit
	cursor := query.Cursor
	// scanKeysByPrefix 返回当前页的 keys 和下一页的游标
	var keys []string
	count := pageSize * 2 // COUNT 设置大一点，确保能收集够一页（SCAN 不保证精确数量）

	for {
		scan := client.Redis.Scan(ctx, uint64(cursor), util.ChatHistoryPrefix+"*", count)
		currKeys, nextCursor, err := scan.Result()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		keys = append(keys, currKeys...)

		// 如果已经收集够一页，直接返回
		if int64(len(keys)) >= pageSize {
			if int64(len(keys)) > pageSize {
				// 多余的留给下一页（实际可缓存或丢弃，这里简单截取）
				keys = keys[:pageSize]
			}
			break
		}

		cursor = int64(nextCursor)
		// 如果游标为 0，已遍历完
		if cursor == 0 {
			break
		}
	}

	return c.JSON(http.StatusOK, SessionList{
		Sessions:   keys,
		NextCursor: cursor,
	})
}

type SessionHistory struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (s *Session) History(c echo.Context) error {
	ctx := c.Request().Context()
	sessionId := c.QueryParam("session_id")
	store := memory.NewRedisStore(client.Redis)
	messages, err := store.GetRecentMessages(ctx, sessionId, -ContextWindowSize, -1)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	var history []SessionHistory
	for _, message := range messages {
		history = append(history, SessionHistory{
			Role:    string(message.Role),
			Content: message.Content,
		})
	}

	return c.JSON(http.StatusOK, history)
}
