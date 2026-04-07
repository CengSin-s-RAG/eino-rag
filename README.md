# 文章智能问答 Agent

基于 Cloudwego Eino 框架构建的 RAG (Retrieval Augmented Generation) 智能问答系统，专注于金融文章的知识检索与问答。

## 技术框架

- **核心框架**: [Cloudwego Eino](https://github.com/cloudwego/eino) - Go AI Agent 开发框架
- **Web 框架**: Echo v4 - 高性能 Go HTTP 框架
- **向量数据库**: Qdrant - 高性能向量检索引擎
- **缓存/会话存储**: Redis
- **关系数据库**: MySQL (默认) / PostgreSQL (可选，支持 pgvector)
- **AI 模型**: OpenRouter API (支持多种 LLM)

## 分支说明

| 分支 | 功能说明 |
|------|----------|
| `main` | 主分支，稳定版本 |
| `query_rewriting` | **当前工作分支** - 查询改写功能，支持长对话上下文压缩 |
| `react_agent` | React Agent 实现，基于 Eino 的 ReAct 模式 |
| `rerank` | 文档重排序功能，支持 LLM 驱动的相关性排序 |
| `simple_chat` | 简化版聊天实现 |
| `chat_history` | 聊天历史管理与持久化功能 |
| `feature/migrate-to-postgresql` | PostgreSQL + pgvector 迁移支持 |
| `feature/openclaw_code_basic` | OpenClaw 代码基础集成 |

## 项目结构

```
.
├── agent/               # Agent 核心实现
│   ├── graph_agent.go  # Eino Chat Agent 封装
│   ├── state.go        # Agent 状态定义
│   └── component/      # Agent 组件
│       ├── embedding.go          # 向量化组件
│       ├── vector_retriever.go   # 向量检索器
│       ├── full_text_retriever.go# 全文检索器
│       ├── my_content_retriever.go# 混合检索器
│       ├── rerank.go             # 重排序组件
│       └── rewrite.go            # 查询改写组件
├── api/                # HTTP API 处理
│   ├── eino_handler.go # Agent 对话处理器
│   ├── session.go      # 会话管理 API
│   └── struct.go       # 请求/响应结构
├── client/             # 外部服务客户端
│   ├── mcp.go          # MCP (Model Context Protocol) 客户端
│   ├── tools.go        # Agent 工具封装
│   └── client.go       # 客户端初始化
├── config/             # 配置管理
│   └── config.yaml     # 配置文件
├── memory/             # 会话记忆存储
│   └── redis.go        # Redis 会话存储实现
├── model/              # 数据模型
├── dao/                # 数据访问层
├── demo/               # 示例代码
├── document/           # 文档处理
├── visualize/          # 可视化组件
├── util/               # 工具函数
│   └── prompt.go       # 提示词管理
└── main.go             # 程序入口
```

## Web API 接口

服务启动后监听 `:8086`

### 对话接口

#### POST `/v2/chat`
主对话接口，基于 RAG 的智能问答。

**请求体**:
```json
{
  "session_id": "string",  // 会话 ID
  "question": "string"      // 用户问题
}
```

**响应**:
```json
{
  "session_id": "string",
  "reply_message": "string"  // AI 回复内容
}
```

**功能说明**:
- 支持多轮对话，自动维护上下文
- 混合检索: 向量检索 + 全文检索
- 长对话自动压缩 (超过 19 轮对话时触发查询改写)
- 会话历史自动存入 Redis

---

#### GET `/v2/session/list`
获取会话列表。

**响应**:
```json
{
  "sessions": ["session_id_1", "session_id_2"]
}
```

---

#### GET `/v2/session/history`
获取指定会话的历史消息。

**查询参数**:
- `session_id`: 会话 ID

**响应**: 会话消息列表

---

### 检索增强接口

#### POST `/v2/rerank`
文档重排序接口，使用 LLM 对候选文档进行相关性排序。

**请求体**:
```json
{
  "question": "string",           // 用户问题
  "documents": ["string"]         // 待排序的文档列表
}
```

**响应**:
```json
{
  "ordered_documents": ["doc1", "doc2"],  // 排序后的文档
  "reason": "排序理由"
}
```

---

#### POST `/v2/rewrite`
查询改写接口，对长对话上下文进行压缩优化。

**请求体**:
```json
{
  "history": "string",   // 历史对话内容
  "query": "string"      // 当前用户问题
}
```

**响应**: 改写后的优化查询

---

## 配置说明

配置文件: `config/config.yaml`

主要配置项:
- `qdrant` - 向量数据库连接
- `redis` - 会话存储
- `mcpServer` - MCP 服务器地址
- `ivankaContent` - MySQL 数据库连接
- `rerank` / `rewrite` - 各模块提示词路径

环境变量:
- `OPENROUTER_API_KEY` - OpenRouter API 密钥
- `OPENROUTER_API_BASE_URL` - OpenRouter API 地址
- `OPENROUTER_MODEL` - 默认模型
- `AMP_PLUS_API_KEY` - APM Plus 监控密钥

## 运行

```bash
# 设置环境变量
export OPENROUTER_API_KEY="your-api-key"
export OPENROUTER_API_BASE_URL="https://openrouter.ai/api/v1"
export OPENROUTER_MODEL="your-model"
export AMP_PLUS_API_KEY="your-apm-key"

# 运行
go run main.go
```

## 核心特性

1. **混合检索**: 结合向量相似度与全文关键词匹配
2. **查询改写**: 长对话自动压缩，保留关键信息
3. **文档重排序**: 基于 LLM 的相关性智能排序
4. **会话管理**: Redis 持久化，支持多会话并发
5. **可观测性**: 集成 APM Plus 全链路追踪
