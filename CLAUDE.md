# z-agent

Go 服务，使用 OpenAI Chat Completions 兼容协议、Redis 会话存储和必需的 MCP 工具服务。

## Commands

```bash
make run
go test ./...
go vet ./...
```

## Boundaries

- `runtime` 定义 Provider 无关的消息与工具调用模型。
- `provider/openai` 负责 OpenAI wire format；工具必须使用 `type: function` 和嵌套 `function`。
- `service` 不依赖 Echo；`transport` 只处理 HTTP。
- `bootstrap` 负责配置装配、MCP 握手和生命周期。
- MCP 配置为空或任何 MCP 连接/工具发现失败时，服务必须启动失败。
