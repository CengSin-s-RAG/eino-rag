package main

import (
	"agent.article.fp/agent"
	"agent.article.fp/api"
	"agent.article.fp/client"
	"agent.article.fp/config"
	"agent.article.fp/util"
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/callbacks/apmplus"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/devops"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/schema"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io"
	"log"
	"os"
)

func main() {
	ctx := context.Background()
	if err := devops.Init(ctx); err != nil {
		log.Fatalln(fmt.Errorf("init devops error: %v", err))
	}

	// 创建apmplus handler
	cbh, shutdown, err := apmplus.NewApmplusHandler(&apmplus.Config{
		Host:        "apmplus-cn-beijing.volces.com:4317",
		AppKey:      os.Getenv("AMP_PLUS_API_KEY"),
		ServiceName: "fp-article-agent",
		Release:     "release/v0.0.1",
	})
	if err != nil {
		log.Fatalln(fmt.Errorf("init apmplus error: %v", err))
	}

	// 设置apmplus为全局callback
	callbacks.AppendGlobalHandlers(cbh)

	defer func() {
		if err = shutdown(ctx); err != nil {
			log.Fatalln(fmt.Errorf("shutdown error: %v", err))
		}
	}()

	util.InitSystemPrompt()

	var cfg config.Config
	_ = cleanenv.ReadConfig("./config/config.yaml", &cfg)

	client.InitRedis(cfg.Redis)
	client.InitQdrant(cfg.Qdrant)
	//client.InitTemporal(cfg.Temporal, &client.Temporal)
	//client.InitTemporal(cfg.SyncTemporal, &client.SyncTemporal)
	client.InitMcpClient(cfg.McpServer)
	client.InitTools()
	client.InitMysql(cfg.Mysql)
	defer client.Close()

	// 先初始化所需的 chatModel
	// 先初始化所需的 chatModel
	conf := openai.ChatModelConfig{
		APIKey:  os.Getenv("OPENROUTER_API_KEY"),
		BaseURL: os.Getenv("OPENROUTER_API_BASE_URL"),
		Model:   os.Getenv("OPENROUTER_MODEL"),
	}
	chatAgent, err := agent.NewEinoChatAgent(ctx, conf)
	if err != nil {
		log.Fatalln(fmt.Errorf("NewEinoChatAgent: %v", err))
	}

	einoHandler := api.NewEinoChatAgentHandler(chatAgent)

	e := echo.New()
	e.Use(middleware.CORS())

	e.POST("/v2/chat", einoHandler.HandleQuery)

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
