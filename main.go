package main

import (
	"agent.article.fp/agent"
	"agent.article.fp/api"
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	// 先初始化所需的 chatModel
	// 先初始化所需的 chatModel
	config := openai.ChatModelConfig{
		APIKey:  os.Getenv("OPENROUTER_API_KEY"),
		BaseURL: os.Getenv("OPENROUTER_API_BASE_URL"),
		Model:   os.Getenv("OPENROUTER_MODEL"),
	}
	chatAgent, err := agent.NewEinoChatAgent(ctx, config)
	if err != nil {
		log.Fatalln(fmt.Errorf("NewEinoChatAgent: %v", err))
	}

	einoHandler := api.NewEinoChatAgentHandler(chatAgent)

	e := echo.New()
	e.Use(middleware.CORS())

	e.POST("/v2/chat", sse(einoHandler.HandleQuery))

	if err := e.Start(":8086"); err != nil {
		log.Fatalln(err)
	}
}

func sse(handle func(c echo.Context) (*schema.StreamReader[*schema.Message], error)) echo.HandlerFunc {
	return func(c echo.Context) error {
		w := c.Response()
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		s, err := handle(c)
		if err != nil {
			return err
		}
		defer s.Close()

		for {
			recv, err := s.Recv()
			if err != nil && io.EOF == err {
				break
			}
			if recv.Content == "" {
				continue
			}
			_, err = fmt.Fprintf(w, "data: %s\n\n", recv.Content)
			if err != nil {
				return err
			}
			w.Flush()
		}
		return nil
	}
}
