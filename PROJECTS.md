# 项目索引

## fupeng-article-agent

路径：`/Users/cengsin/my_projects/Eino-Projects/fupeng-article-agent`

### 概述
这是一个使用 Go 语言开发的 RAG（检索增强生成）项目，主要用于文章内容的智能检索和问答。

### 技术栈
- **语言**: Go
- **数据库**: PostgreSQL + pgvector（原使用 MySQL + Qdrant）
- **框架**: Eino 组件
- **向量存储**: pgvector 扩展
- **全文搜索**: PostgreSQL tsvector/tsquery
- **RAG**: 混合检索器（向量 + 全文）

### 项目结构
```
├── agent/           # 代理组件
│   ├── component/  # 检索器、嵌入器等
│   └── ...
├── client/         # 客户端和数据库连接
├── config/         # 配置管理
├── dao/           # 数据访问层
├── model/          # 数据模型
├── util/          # 工具函数
└── go.mod/go.sum  # 依赖管理
```

### 分支说明

#### 主要分支
- **`main`**: 稳定版本，生产部署
- **`feature/migrate-to-postgresql`**: 从 MySQL+Qdrant 迁移到 PostgreSQL+pgvector（当前分支）
  - 已完成迁移工作，待测试验证

#### 功能分支
- **`query_rewriting`**: 查询重写功能
- **`rerank`**: 重新排序功能优化
- **`react_agent`**: React 代理实现
- **`simple_chat`**: 简化聊天功能
- **`chat_history`**: 聊天历史管理

#### 远程分支
- `origin/main`: 主分支远程镜像
- `origin/query_rewriting`: 查询重写功能远程分支
- `origin/react_agent`: React 代理远程分支
- `origin/rerank`: 重新排序远程分支
- `origin/simple_chat`: 简化聊天远程分支
- `origin/chat_history`: 聊天历史远程分支

### 最近变更
- **2026-01-28**: 完成从 MySQL+Qdrant 到 PostgreSQL+pgvector 的迁移
  - 更新配置文件使用 PostgreSQL 格式
  - 重写向量检索器使用 pgvector
  - 重写全文检索器使用 PostgreSQL tsvector
  - 添加向量存储模型

### 依赖关系
- **外部服务**: 使用 OpenAI API（通过 OpenRouter）进行嵌入和生成
- **外部工具**: 从(mcp-server)[/Users/cengsin/my_projects/mcp-server]获取可用的工具,需要先启动mcp服务才能正常启动当前服务。
- **前端页面**: (rag4financenew_frontend)[/Users/cengsin/my_projects/rag4financenew_frontend]

### 配置说明
- 数据库连接：PostgreSQL + pgvector 扩展
- 需要创建 `vector_store` 表存储向量数据
- API 密钥通过环境变量配置

### 启动说明
1. 确保 PostgreSQL 已安装 pgvector 扩展
2. 配置 `config/config.yaml` 中的数据库连接
3. 运行 `go run main.go` 启动服务