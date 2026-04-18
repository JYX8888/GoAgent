# GoAgent 框架文档

## 目录

1. [框架简介](#框架简介)
2. [快速开始](#快速开始)
3. [核心架构](#核心架构)
4. [使用方法](#使用方法)
5. [技术亮点与处理方法](#技术亮点与处理方法)
6. [配置说明](#配置说明)
7. [框架优点](#框架优点)

---

## 框架简介

GoAgent 是一个用 Go 语言实现的 AI Agent 框架。采用 Go 语言以获得更好的性能和并发安全性。

### 特性

- 多 Provider 支持：OpenAI、DeepSeek、Qwen、Moonshot、Kimi、Zhipu、Ollama、VLLM 等
- 多种 Agent 类型：SimpleAgent、ReActAgent、PlanAndSolveAgent、ReflectionAgent
- 混合记忆系统：工作记忆、情景记忆、语义记忆、感知记忆
- 工具系统：函数调用、异步执行、工具链
- 通信协议：A2A、ANP、MCP

---

## 快速开始

### 1. 安装依赖

```bash
cd GoAgent
go mod tidy
```

### 2. 配置环境变量

创建 `.env` 文件：

```bash
cp config/.env.example .env
```

编辑 `.env`，填写你的 API Key：

```bash
# OpenAI 配置
OPENAI_API_KEY=your_key_here

# 或者使用其他 Provider
DASHSCOPE_API_KEY=your_aliyun_key
```

### 3. 运行示例

```go
package main

import (
	"context"
	"fmt"
	"log"

	"GoAgent/pkg/agents"
	"GoAgent/pkg/core"
)

func main() {
	// 创建 LLM 客户端
	llm := core.NewLLM(
		core.WithProvider("openai"),
		core.WithModel("gpt-3.5-turbo"),
	)

	// 创建 Agent
	agent := agents.NewSimpleAgent(
		"assistant",
		llm,
		"You are a helpful assistant.",
	)

	// 运行
	ctx := context.Background()
	resp, err := agent.Run(ctx, "你好，请介绍一下你自己")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp)
}
```

---

## 核心架构

```
GoAgent/
├── pkg/
│   ├── core/          # 核心接口与类型
│   │   ├── agent.go   # Agent 接口定义
│   │   ├── llm.go    # LLM 客户端
│   │   ├── message.go # 消息类型
│   │   └── config.go  # 配置
│   ├── agents/       # Agent 实现
│   │   ├── simple_agent.go
│   │   ├── react_agent.go
│   │   ├── plan_solve_agent.go
│   │   └── reflection_agent.go
│   ├── memory/      # 记忆系统
│   │   ├── working.go    # 工作记忆
│   │   ├── episodic.go   # 情景记忆
│   │   ├── semantic.go # 语义记忆
│   │   ├── perceptual.go
│   │   └── manager.go
│   ├── tools/       # 工具系统
│   │   ├── tool.go
│   │   ├── registry.go
│   │   ├── builtin.go
│   │   └── async.go
│   ├── protocols/   # 通信协议
│   │   ├── a2a.go
│   │   ├── anp.go
│   │   └── mcp.go
│   └── storage/    # 存储（已集成 MySQL/Neo4j/Qdrant）
└── config/        # 配置管理
```

---

## 使用方法

### 1. 创建不同类型的 Agent

#### SimpleAgent（简单 Agent）

```go
agent := agents.NewSimpleAgent(
	"my-agent",
	llm,
	"You are a helpful assistant.",
	agents.WithToolRegistry(registry),
)
```

#### ReActAgent（推理+行动 Agent）

```go
agent := agents.NewReActAgent(
	"react-agent",
	llm,
	systemPrompt,
	agents.WithToolRegistry(registry),
)
```

#### PlanAndSolveAgent（计划+执行 Agent）

```go
agent := agents.NewPlanAndSolveAgent(
	"planner",
	llm,
	systemPrompt,
)
```

#### ReflectionAgent（反思 Agent）

```go
agent := agents.NewReflectionAgent(
	"reflector",
	llm,
	systemPrompt,
)
```

### 2. 使用工具系统

#### 注册自定义工具

```go
type MyTool struct {
	BaseTool
}

func NewMyTool() *MyTool {
	return &MyTool{
		BaseTool: BaseTool{
			Name_:        "my_tool",
			Description_: "My custom tool",
			Parameters_: []ToolParameter{
				NewToolParameter("input", "string", "Input description", true, nil),
			},
		},
	}
}

func (t *MyTool) Run(params map[string]interface{}) (string, error) {
	input := params["input"].(string)
	// 处理逻辑
	return "result", nil
}

// 注册
registry.Register(NewMyTool())
```

#### 使用内置工具

```go
// 计算器
calc := tools.NewCalculatorTool()

// 搜索
search := tools.NewSearchTool()

// 记忆工具
memTool := tools.NewMemoryTool(manager)
```

### 3. 使用记忆系统

```go
// 创建记忆管理器
manager := memory.NewMemoryManager(
	memory.WithUserIDOpt("user123"),
)

// 添加记忆
manager.AddMemory(
	"重要的事情",
	memory.MemoryTypeEpisodic,
	0.8,
	map[string]interface{}{"topic": "work"},
)

// 搜索记忆
results := manager.Search("重要的事情", nil)

// 获取所有记忆
all := manager.List()
```

### 4. 使用存储系统

#### MySQL 存储

```go
import "GoAgent/pkg/memory/storage"

// 配置（从环境变量加载）
cfg := storage.LoadMySQLConfigFromEnv()

// 获取数据库连接
db, err := storage.GetDB()
if err != nil {
	log.Fatal(err)
}

// 初始化表结构
err = storage.InitMySQLSchema()
```

### 5. 使用 LLM Provider

框架支持多种 LLM Provider：

```go
// OpenAI
llm := core.NewLLM(
	core.WithProvider(core.ProviderOpenAI),
	core.WithModel("gpt-4o"),
)

// DeepSeek
llm := core.NewLLM(
	core.WithProvider(core.ProviderDeepSeek),
	core.WithModel("deepseek-chat"),
)

// 阿里 Qwen (通义千问)
llm := core.NewLLM(
	core.WithProvider(core.ProviderQwen),
	core.WithModel("qwen-plus"),
)

// Kimi (月之暗面)
llm := core.NewLLM(
	core.WithProvider(core.ProviderKimi),
	core.WithModel("moonshot-v1-8k"),
)

// Zhipu (智谱)
llm := core.NewLLM(
	core.WithProvider(core.ProviderZhipu),
	core.WithModel("glm-4"),
)

// Ollama (本地)
llm := core.NewLLM(
	core.WithProvider(core.ProviderOllama),
	core.WithModel("llama3.2"),
)

// VLLM (本地)
llm := core.NewLLM(
	core.WithProvider(core.ProviderVLLM),
)
```

---

## 技术亮点与处理方法

### 1. TF-IDF 文本向量化

Python sklearn 的 `TfidfVectorizer` 在 Go 中使用 `github.com/rioloc/tfidf-go` 实现：

```go
import (
	tfidf "github.com/rioloc/tfidf-go"
	"github.com/rioloc/tfidf-go/token"
)

// 创建 TF-IDF Embedder
embedder := &tfidfEmbedder{
	maxFeatures: 1000,
}

// Fit 模型
embedder.Fit(documents)

// Encode 文本
vectors := embedder.Encode(texts)
```

### 2. 数据库存储方案

#### MySQL 替代 SQLite

使用 `github.com/go-sql-driver/mysql`：

```go
import _ "github.com/go-sql-driver/mysql"

// 配置加载
cfg := &MySQLConfig{
	Host:     "localhost",
	Port:     3306,
	User:     "root",
	Password: "",
	Database: "goagent",
}

// DSN 连接字符串
dsn := "root:Yl300822!@tcp(localhost:3306)/goagent?charset=utf8mb4"
```

#### Neo4j 知识图谱

使用 `github.com/neo4j/neo4j-go-driver/v6`：

```go
client, err := storage.NewNeo4jClient(cfg)
defer client.Close(ctx)

// 创建节点
client.CreateNode(ctx, "Person", props)

// ��建关系
client.CreateRelationship(ctx, "person1", "person2", "KNOWS", props)

// 查询邻居
neighbors, _ := client.FindNeighbors(ctx, "node_id", 3)
```

#### Qdrant 向量数据库

使用 `github.com/qdrant/go-client`：

```go
client, _ := storage.NewQdrantClient(cfg)

// 创建集合
client.CreateCollection(ctx, "my_collection", 384, "cosine")

// 插入向量
client.UpsertVectors(ctx, "my_collection", vectors)

// 搜索
results, _ := client.Search(ctx, "my_collection", queryVector, 10, nil)
```

### 3. 并发安全处理

使用 `sync.RWMutex` 实现读写锁：

```go
type BaseAgent struct {
	mu      sync.RWMutex
	history []*Message
}

func (a *BaseAgent) GetHistory() []*Message {
	a.mu.RLock()
	defer a.mu.RUnlock()
	// 复制数据防止竞态
	result := make([]*Message, len(a.history))
	copy(result, a.history)
	return result
}
```

### 4. 选项模式（Functional Options）

使用函数式选项替代构造函数参数：

```go
type AgentOption func(*Agent)

func WithSystemPrompt(prompt string) AgentOption {
	return func(a *Agent) { a.SystemPrompt = prompt }
}

func WithConfig(cfg *Config) AgentOption {
	return func(a *Agent) { a.Config = cfg }
}

// 使用
agent := NewAgent(
	WithSystemPrompt("You are helpful."),
	WithConfig(myConfig),
)
```

### 5. 工具调用协议

采用 `[TOOL_CALL:tool_name:params]` 格式：

```
[TOOL_CALL:calculator:a=12,b=8]
[TOOL_CALL:search:query=Python]
```

解析正则表达式：`\[TOOL_CALL:([^:]+):([^\]]+)\]`

### 6. 流式输出支持

```go
ch, errCh := llm.Stream(ctx, messages)

for {
	select {
	case chunk, ok := <-ch:
		if !ok {
			break
		}
		fmt.Print(chunk)
	case err := <-errCh:
		fmt.Println("Error:", err)
		break
	}
}
```

### 7. 配置管理

支持多种配置加载方式：

```go
// 从环境变量
cfg := config.Load()

// 从文件
cfg, err := config.LoadFromFile(".env")

// 自动查找 .env 文件
cfg, err := config.LoadFromEnvFile()
```

---

## 配置说明

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `OPENAI_API_KEY` | OpenAI API Key | - |
| `DASHSCOPE_API_KEY` | 阿里云 API Key | - |
| `DEEPSEEK_API_KEY` | DeepSeek API Key | - |
| `LLM_MODEL_ID` | 模型名称 | - |
| `LLM_BASE_URL` | API 端点 | - |
| `LLM_PROVIDER` | Provider | openai |
| `MYSQL_HOST` | MySQL 主机 | - |
| `MYSQL_USER` | MySQL 用户 | - |
| `MYSQL_PASSWORD` | MySQL 密码 | - |
| `MYSQL_DATABASE` | MySQL 数据库 | - |
| `NEO4J_URI` | Neo4j 地址 | - |
| `NEO4J_USERNAME` | Neo4j 用户 | - |
| `NEO4J_PASSWORD` | Neo4j 密码 | - |
| `QDRANT_URL` | Qdrant 地址 | - |
| `QDRANT_API_KEY` | Qdrant Key | - |
| `LOG_LEVEL` | 日志级别 | INFO |
| `DEBUG` | 调试模式 | false |

---

## 框架优点

### 1. 零外部依赖倾向

- 核心库仅依赖 `go-openai` SDK
- 其他数据库驱动按需引入
- 保持二进制体积可控

### 2. 高并发安全

- 所有 Agent 使用 `sync.RWMutex` 保护状态
- 工具执行支持异步并发
- 记忆系统线程安全

### 3. 模块化设计

- 清晰的包结构
- 接口驱动开发
- 易于扩展新功能

### 4. 多 Provider 支持

- 一行代码切换 AI Provider
- 自动检测 API Key
- 统一的 API 接口

### 5. 工具系统

- 函数式工具注册
- 工具链支持
- 异步执行能力

### 6. 记忆系统

- 四种记忆类型
- 混合存储架构
- 向量检索支持

### 7. Go 语言优势

- 编译型语言，高性能
- 原生协程goroutine并发
- 静态类型，编译时检查

### 8. 协议支持

- A2A (Agent to Agent)
- ANP (Agent Network Protocol)
- MCP (Model Context Protocol)

---

## 常见问题

### Q: 如何添加新的 LLM Provider？

在 `pkg/core/llm.go` 中添加：

```go
const (
	ProviderNew Provider = "new_provider"
)

var providerBaseURLs = map[Provider]string{
	ProviderNew: "https://api.newprovider.com/v1",
}
```

### Q: 如何创建自定义 Agent？

实现 `core.Agent` 接口：

```go
type MyAgent struct {
	*core.BaseAgent
}

func (a *MyAgent) Run(ctx context.Context, input string) (string, error) {
	// 实现逻辑
	return "", nil
}
```

### Q: 如何连接已有的 Neo4j/Qdrant？

设置环境变量或使用配置：

```bash
export NEO4J_URI="bolt://your-neo4j:7687"
export NEO4J_USERNAME="neo4j"
export NEO4J_PASSWORD="your_password"
export QDRANT_URL="http://your-qdrant:6333"
```

---

## 更新日志

### v1.0.0 (2025-04)

- 初始版本
- 支持多种 LLM Provider
- 实现 SimpleAgent、ReActAgent、PlanAndSolveAgent、ReflectionAgent
- 实现混合记忆系统
- 集成 MySQL/Neo4j/Qdrant 存储
- 实现 TF-IDF 向量化

---

## 许可证

MIT License