# CLAUDE.md

## 项目概述

金融文章智能问答 Agent，基于 CloudWeGo Eino 框架的 **Agentic RAG** 系统。用户提问后，ReAct Agent 自主决定是否调用工具检索华尔街见闻金融文章，经过混合检索（向量 + 全文）→ 去重 → MySQL 取原文 → LLM 重排序后，生成带引用的回答。

**模块名**: `agent.article.fp`

## 技术栈

| 层 | 技术 |
|---|---|
| Agent 框架 | CloudWeGo Eino (`eino/flow/agent/react`) |
| Web 框架 | Echo v4，监听端口 `:8086` |
| 向量数据库 | Qdrant（端口 6334），混合检索：向量 + 全文 |
| 缓存/会话 | Redis（DB 1），按 session 存储聊天历史（List 结构） |
| 关系数据库 | MySQL `ivanka_content` 库，`article_entries` 表存储文章原文 |
| LLM | OpenRouter API（环境变量配置：`API_KEY`, `BASE_URL`, `MODEL`） |
| Embedding | `qwen/qwen3-embedding-8b`，1536 维 |
| Tool 协议 | MCP（`mcp-go`），连接外部 MCP Server（端口 8085） |
| Skill 规范 | [Agent Skills](https://agentskills.io)，`skills/` 目录下 SKILL.md |

## 项目结构

```
.
├── main.go                          # 程序入口，路由注册，模型初始化
├── agent/
│   ├── graph_agent.go               # ReAct Agent 构建，工具绑定，Skill 集成
│   ├── state.go                     # ChatState 定义
│   └── component/
│       ├── my_content_retriever.go  # 混合检索器（向量 + 全文 → 去重 → MySQL → Rerank）
│       ├── vector_retriever.go      # Qdrant 向量检索器（qwen3-embedding-8b）
│       ├── full_text_retriever.go   # Qdrant 全文检索器
│       ├── embedding.go             # Embedding 组件
│       ├── rerank.go                # LLM 文档重排序（HTTP 调用 /v2/rerank）
│       └── rewrite.go               # 查询改写（HTTP 调用 /v2/rewrite）
├── api/
│   ├── eino_handler.go              # /v2/chat, /v2/rerank, /v2/rewrite 处理器
│   ├── session.go                   # /v2/session/new, /list, /history 处理器
│   └── struct.go                    # ChatReq, ChatResp 结构体
├── client/
│   ├── client.go                    # Init()/Close()，初始化 Qdrant/Redis/MySQL/MCP
│   ├── mcp.go                       # MCP 客户端，ToolsInfo/EinoTools 全局注册表
│   ├── tools.go                     # NewRetrieverTool — 将 Retriever 包装为 Tool
│   ├── skill_tool.go                # NewActivateSkillTool — Skill 激活工具
│   ├── exec_tool.go                 # NewExecCommandTool — Shell 命令执行工具（白名单）
│   └── model.go                     # 全局模型实例（RerankModel, QueryRewriteModel）
├── skill/
│   └── skill.go                     # Skill Manager — 发现、解析、Catalog 生成
├── skills/
│   ├── article-summary/SKILL.md     # 示例 Skill：文章结构化摘要
│   └── tickflow-realtime/           # 示例 Skill：TickFlow 实时数据
├── config/
│   ├── config.go                    # Config 结构体定义
│   └── config.yaml                  # 配置文件
├── memory/
│   ├── store.go                     # MemoryStore 接口 + Gob 序列化
│   ├── redis.go                     # Redis 实现（RPUSH/LRANGE）
│   └── inmem.go                     # 内存实现（测试用，基于 miniredis）
├── model/
│   └── model.go                     # ArticleEntries 数据模型（~60 字段）
├── dao/
│   └── article_entries.go           # MySQL 数据访问层（GORM）
├── util/
│   ├── constant.go                  # 常量（TopK, ScoreTs, 提示词变量）
│   ├── prompt.go                    # GetSystemPrompt(skillCatalog) 系统提示词
│   ├── log.go                       # SimpleLogger（Eino 回调日志）
│   └── interface.go                 # 通用接口定义
├── systemPrompt.md                  # 系统提示词文件
├── rerankSystemPrompt.txt           # Rerank 提示词
└── rewriteSystemPrompt.txt          # Query Rewrite 提示词
```

## API 路由

所有路由在 `/v2` 组下，服务监听 `:8086`：

| 方法 | 路径 | 处理器 | 说明 |
|------|------|--------|------|
| POST | `/v2/chat` | `einoHandler.HandleQuery` | 主对话接口（RAG Chat） |
| POST | `/v2/rerank` | `einoHandler.HandleRerank` | 文档重排序 |
| POST | `/v2/rewrite` | `einoHandler.HandleRewrite` | 查询改写 |
| POST | `/v2/session/new` | `session.NewSession` | 创建新会话（返回 UUID） |
| GET | `/v2/session/list` | `session.List` | 会话列表（Redis SCAN 分页） |
| GET | `/v2/session/history` | `session.History` | 会话历史消息 |

## 工具注册机制

Agent 有三个工具来源，全部注册到全局切片 `client.ToolsInfo`（schema）和 `client.EinoTools`（executable）：

### 1. MCP 工具（`client/mcp.go`）
- `InitMcpClient()` 连接 MCP Server（`http://localhost:8085/mcp`）
- `InitTools()` 将 MCP 工具转为 Eino 工具并注册

### 2. 检索工具（`client/tools.go` + `agent/graph_agent.go`）
- `NewRetrieverTool()` 将混合检索器包装为 `search_financial_knowledge` 工具
- 检索流程：向量搜索（TopK=10, ScoreThreshold=0.5）∥ 全文搜索（TopK=10）→ 去重 → MySQL 取原文 → LLM Rerank（ScoreTs=0.5 过滤）

### 3. Skill 工具（`client/skill_tool.go`）
- `NewActivateSkillTool()` 创建 `activate_skill` 工具
- LLM 判断任务匹配某个 Skill 时调用，返回完整 SKILL.md 指令

### 4. 命令执行工具（`client/exec_tool.go`）
- `NewExecCommandTool()` 创建 `execute_command` 工具
- 白名单校验（`config.yaml` 的 `exec.allowedCommands`）
- 60 秒超时，输出截断 8000 字符
- 命令失败时将错误信息作为工具输出返回（非 error），让 LLM 决定如何处理

绑定顺序：在 `agent/graph_agent.go` 的 `NewEinoChatAgent()` 中依次注册，最后通过 `chatModel.BindTools(client.ToolsInfo)` 绑定到模型。

## Skill 系统（Agent Skills 规范）

遵循 [agentskills.io](https://agentskills.io) 规范，实现渐进式加载：

1. **启动时**：扫描 `skills/` 目录，解析 `SKILL.md` 的 YAML frontmatter（`name` + `description`）
2. **注入 Catalog**：将所有 Skill 的 name+description 以 XML 格式追加到系统提示词
3. **LLM 激活**：LLM 判断需要时调用 `activate_skill` 工具加载完整指令
4. **按需加载资源**：Skill 目录下的 `scripts/`、`references/`、`assets/` 由 LLM 按需读取

添加新 Skill：在 `skills/` 下创建目录 + `SKILL.md`，重启服务即可。

## 系统提示词

文件 `util/systemPrompt.md`，通过 `util.GetSystemPrompt(skillCatalog)` 加载。当 Skill 可用时，追加 `<available_skills>` 目录。在 `MessageModifier` 中，对话超过 19 条时会触发查询改写并重新注入系统提示词。

## 配置

`config/config.yaml` 关键配置项：

```yaml
ragPromptPath: "systemPrompt.md"    # 系统提示词路径
skillsPath: "skills"                 # Skill 目录路径
exec:
  allowedCommands:                   # 命令执行白名单
    - "ls" - "cat" - "curl" - "python3" ...
  timeout: 60                        # 命令超时（秒）
qdrant: { host: "localhost", port: 6334 }
redis: { addr: "localhost:6379", DB: 1 }
mcpServer: "http://localhost:8085/mcp"
ivankaContent: { host, port, userName, password, DB }
rerank: { prompt, url }              # Rerank 提示词和自调用 URL
rewrite: { prompt, url }             # Rewrite 提示词和自调用 URL
```

环境变量：`API_KEY`、`BASE_URL`、`MODEL`

## 开发环境

```bash
# Docker 依赖
docker run -p 6333:6333 -p 6334:6334 -v "$(pwd)/qdrant_storage:/qdrant/storage:z" qdrant/qdrant
docker run -d --name mysql-work -p 3306:3306 -e MYSQL_ROOT_PASSWORD=rootpassword mysql:5.7
docker run -d --name redis -p 6379:6379 redis:latest redis-server --appendonly yes

# 启动
export API_KEY="your-key" BASE_URL="https://openrouter.ai/api/v1" MODEL="openai/gpt-5"
go run main.go
```

## 请求流程

```
POST /v2/chat {session_id, question}
  → Redis 加载最近 20 条历史
  → 追加 user message
  → ReAct Agent（MaxStep=15）:
      LLM 推理 → 决定调用工具 → 执行 → 返回结果 → LLM 综合回答
      可用工具：
        - search_financial_knowledge（混合检索）
        - activate_skill（加载 Skill 指令）
        - execute_command（执行 Shell 命令）
        - MCP 外部工具
  → Redis 异步存储 user + assistant 消息
  → 返回 {session_id, reply_message}
```

## 注意事项

- `rerank` 和 `rewrite` 端点是自调用（URL 指向 `localhost:8086`），Agent 通过 HTTP 调用自身
- `ChatReq.UnmarshalJSON` 中 session_id 为空时自动生成 UUID
- 并发检索使用 goroutine，闭包变量需通过参数传递（已修复 `my_content_retriever.go`）
- `execute_command` 的错误信息作为工具输出返回（非 error），避免 ReAct Agent 中断
- `temperature` 和 `top_p` 不要同时设置，OpenRouter 部分模型会返回 400
