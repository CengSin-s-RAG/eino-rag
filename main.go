package main

import (
	"agent.article.fp/agent"
	"agent.article.fp/agent/component"
	"agent.article.fp/api"
	"agent.article.fp/client"
	_ "agent.article.fp/config"
	"agent.article.fp/service"
	"agent.article.fp/transport"
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	chatService, err := service.New()
	if err != nil {
		log.Fatal("new chatService fail, info: ", err)
	}

	log.Fatal("service start failed, info: ", transport.Start(chatService))

	client.Init()
	defer client.Close()

	temperature := float32(1)

	conf := openai.ChatModelConfig{
		APIKey:      os.Getenv("API_KEY"),
		BaseURL:     os.Getenv("BASE_URL"),
		Model:       os.Getenv("MODEL"),
		Temperature: &temperature,
	}
	chatAgent, err := agent.NewEinoChatAgent(ctx, conf)
	if err != nil {
		log.Fatalln(fmt.Errorf("NewEinoChatAgent: %v", err))
	}
	einoHandler := api.NewEinoChatAgentHandler(chatAgent)

	schemaDesc := jsonschema.Reflect(&component.RerankState{})
	rerankConf := openai.ChatModelConfig{
		APIKey:  os.Getenv("API_KEY"),
		BaseURL: os.Getenv("BASE_URL"),
		Model:   os.Getenv("MODEL"),
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				JSONSchema: schemaDesc,
			},
		},
		Temperature: &[]float32{0.01}[0],
	}
	client.RerankModel, err = openai.NewChatModel(ctx, &rerankConf)
	if err != nil {
		log.Fatalln(fmt.Errorf("NewChatModel: %v", err))
	}

	rewriteConf := openai.ChatModelConfig{
		APIKey:      os.Getenv("API_KEY"),
		BaseURL:     os.Getenv("BASE_URL"),
		Model:       os.Getenv("MODEL"),
		Temperature: &temperature,
	}
	client.QueryRewriteModel, err = openai.NewChatModel(ctx, &rewriteConf)
	if err != nil {
		log.Fatalln(fmt.Errorf("NewChatModel: %v", err))
	}

	session := api.NewChatSession()

	e := echo.New()
	e.Use(middleware.CORS())

	v2 := e.Group("/v2")
	v2.POST("/chat", einoHandler.HandleQuery)
	v2.POST("/rerank", einoHandler.HandleRerank)
	v2.POST("/rewrite", einoHandler.HandleRewrite)

	ses := v2.Group("/session")
	ses.POST("/new", session.NewSession)
	ses.GET("/list", session.List)
	ses.GET("/history", session.History)

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
