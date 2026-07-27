package client

import (
	"context"
	"fmt"
	einoMcp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"log"
	"time"
)

var (
	McpClient *client.Client
	ToolsInfo []*schema.ToolInfo
	EinoTools []tool.BaseTool
)

func InitMcpClient(server string) {
	createStreamableHTTPClient(server)
}

func createStreamableHTTPClient(server string) {
	// Create StreamableHTTP client
	c, err := client.NewStreamableHttpClient(server)
	if err != nil {
		log.Fatalln("create streamable http client failed, err ", err)
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancelFunc()

	// 3. 握手 (Initialize)
	// 就像 USB 插入时的握手，交换能力信息
	initializeRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "rag_finance_news_tools",
				Version: "1.0.0",
			},
		},
	}

	// Initialize
	if _, err = c.Initialize(ctx, initializeRequest); err != nil {
		log.Fatal(err)
	}

	// Use client
	tools, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Available tools: %d", len(tools.Tools))
	McpClient = c
}

func InitTools() {
	ctx := context.Background()
	tools, err := einoMcp.GetTools(ctx, &einoMcp.Config{Cli: McpClient}) // 只提供了InvokeRun方法，不支持stream调用
	if err != nil {
		log.Fatalln("get tools failed, err ", err.Error())
	}

	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			log.Println(fmt.Sprintf("get tool info failed, err %s", err.Error()))
			continue
		}
		ToolsInfo = append(ToolsInfo, info)
		EinoTools = append(EinoTools, t)
	}
}
