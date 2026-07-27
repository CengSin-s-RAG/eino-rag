# z-agent

一个使用 OpenAI Chat Completions 兼容协议的金融问答 Agent。

## 架构

- `provider/openai`：OpenAI-compatible Chat Completions Provider。
- `tools`：原生命令、Skill 与 MCP 工具 Registry。
- `service`：聊天 ReAct 循环、查询改写、文档重排序与会话服务。
- `store`：Redis 会话历史和会话索引。
- `transport`：Echo HTTP API（`:8086`）。
- `web`：嵌入二进制的静态聊天页面（`:3000`）。

## 运行

先启动 Redis 和配置中的 MCP 服务，然后：

```bash
export API_KEY='your-api-key'
make run
```

访问 `http://localhost:3000`。API 位于 `http://localhost:8086/v2`。

MCP 是启动必需依赖；服务地址在 `config/config.yaml` 的 `mcp` 中配置。

## 验证

```bash
go test ./...
go vet ./...
```
