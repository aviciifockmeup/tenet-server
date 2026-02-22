# Tenet Server

<div align="center">

**Tenet Agent 画板 - 后端服务**

AI Agent 驱动的流程图/架构图协作后端系统

</div>

## 📖 项目简介

Tenet Server 是 Tenet Agent 画板的核心后端服务，提供 Agent 驱动的流程图/架构图自动生成、多人实时协作等功能。用户通过自然语言描述需求，Agent 理解意图并通过工具调用自动绘制图表。项目采用 Go 语言开发，基于 CloudWeGo Hertz 框架，支持 WebSocket 实时通信和 AI 流式响应。

### 核心特性

- 🤖 **Agent 绘图** - 集成 DeepSeek AI 模型，支持 Agent 理解意图并自动生成图表
- 🔧 **工具调用** - Agent 通过 Function Calls 直接操作画板（创建节点、连线等）
- 💬 **对话交互** - 通过自然语言与 Agent 对话，描述绘图需求
- 🎯 **Agent 配置** - 可自定义 Agent 能力和行为配置
- 🎨 **画板绘制** - 支持流程图、架构图、UML 图等多种图表类型
- 🔄 **多人协作** - 基于 WebSocket 的实时协作和房间管理
- 📝 **文档管理** - 画板文档和图形节点的存储管理

## 🏗️ 技术栈

### AI 集成
- **DeepSeek API** - Agent 模型服务
- **SSE (Server-Sent Events)** - 流式响应
- **Function Calls** - Agent 工具调用

### 核心框架
- **Go 1.24.0** - 编程语言
- **Hertz 0.10.4** - CloudWeGo 高性能 HTTP 框架
- **GORM 1.31.1** - ORM 框架

### 数据存储
- **MySQL 8.0+** - 关系型数据库
- **Redis 6.0+** - 缓存和会话管理

### 工具库
- **Zap** - 日志库
- **Viper** - 配置管理
- **WebSocket** - 实时通信
- **UUID** - 唯一标识生成

## 📁 项目结构

```
tenet-server/
├── main.go                           # 应用入口
├── config/                           # 配置模块
│   ├── config.yaml                   # 配置文件
│   └── config.go                     # 配置加载
├── handler/                          # 控制器层
│   ├── ai_stream_handler.go          # AI 流式对话
│   ├── ai_stream_select_handler.go   # AI 工具选择
│   ├── conversation_handler.go       # 对话管理
│   ├── room_handler.go               # 房间管理
│   ├── ws_handler.go                 # WebSocket 处理
│   ├── document.go                   # 文档管理
│   └── node.go                       # 节点管理
├── service/                          # 服务层
│   ├── ai_provider.go                # AI 提供商接口
│   ├── deepseek_provider.go          # DeepSeek 实现
│   ├── agent_service.go              # Agent 服务
│   ├── conversation_service.go       # 对话服务
│   ├── message_service.go            # 消息服务
│   ├── toolcall_service.go           # 工具调用服务
│   ├── document.go                   # 文档服务
│   └── node.go                       # 节点服务
├── model/                            # 数据模型
│   ├── ai_message.go                 # AI 消息模型
│   ├── conversation.go               # 对话模型
│   ├── agent.go                      # Agent 模型
│   └── ...
├── ws/                               # WebSocket 模块
│   ├── hub.go                        # WebSocket Hub
│   └── client.go                     # WebSocket Client
├── dao/                              # 数据访问层
├── database/                         # 数据库初始化
├── logger/                           # 日志模块
├── router/                           # 路由配置
├── utils/                            # 工具类
├── go.mod                            # Go 依赖管理
└── go.sum                            # 依赖校验文件
```

## 🚀 快速开始

### 环境要求

- **Go 1.24.0** 或兼容版本
- **MySQL 8.0+**
- **Redis 6.0+**

### 安装步骤

1. **克隆项目**
```bash
git clone <repository-url>
cd tenet-server
```

2. **配置数据库**
```bash
# 创建数据库
mysql -u root -p
CREATE DATABASE oolong CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

执行数据库迁移脚本（创建所需的表）。

3. **修改配置文件**

编辑 `config/config.yaml`，配置数据库连接：

```yaml
server:
  port: 8081
  name: tenet-server

log:
  level: debug  # debug / info / warn / error

ai:
  deepseek:
    api_key: "your_deepseek_api_key"

database:
  mysql: 
    host: localhost
    port: 3306
    username: root
    password: your_password
    database: oolong
  redis:
    host: localhost
    port: 6379
    db: 0
```

4. **安装依赖**
```bash
go mod tidy
```

5. **启动服务**
```bash
# 直接运行
go run main.go

# 或者先编译后运行
go build -o tenet-server
./tenet-server
```

6. **验证服务**
```bash
# 测试 AI 流式对话接口
curl -X POST http://localhost:8081/api/ai-stream/chat \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {"role": "user", "content": "你好"}
    ],
    "conversationId": "test-conv-001",
    "agentId": "test-agent-001"
  }'
```

服务默认运行在 `http://localhost:8081`

## 📚 核心功能模块

### 1. Agent 绘图

Tenet 的核心功能，Agent 理解用户意图并自动生成图表。

- **对话交互** - 通过自然语言与 Agent 交流，描述绘图需求
- **意图理解** - Agent 理解用户描述的流程、架构等需求
- **自动生成** - Agent 根据理解自动创建图表节点和连线
- **流式响应** - 基于 SSE 的实时流式输出，观察 Agent 思考过程
- **工具调用** - Agent 通过 Function Calls 直接操作画板
- **Agent 配置** - 可自定义 Agent 的能力、提示词和行为模式

**核心接口：**
- `POST /api/ai-stream/chat` - Agent 对话
- `POST /api/ai-stream/select` - Agent 工具选择

### 2. 对话上下文管理

完整的 Agent 对话历史存储，支持多轮对话和上下文延续。

- **对话持久化** - Agent 对话自动保存到数据库
- **历史查询** - 查看完整的对话历史
- **上下文恢复** - 继续之前的对话，Agent 记住之前的绘图内容
- **消息去重** - 基于 messageId 的去重机制
- **序列管理** - 严格的消息顺序保证

**核心接口：**
- `POST /api/conversation/create` - 创建对话会话
- `GET /api/conversation/list` - 获取对话列表
- `GET /api/conversation/:id/messages` - 获取对话历史

### 3. 画板文档与节点管理

画板的核心数据结构，存储 Agent 生成或用户手动创建的图表内容。

- **文档 CRUD** - 创建、读取、更新、删除画板文档
- **节点管理** - 支持矩形、圆形、连线等图形节点的增删改查
- **布局存储** - 保存节点位置、样式、连接关系等数据
- **版本追踪** - 画板内容的版本管理

**核心接口：**
- `POST /api/document` - 创建画板文档
- `GET /api/document/:id` - 获取画板文档
- `PUT /api/document/:id` - 更新画板文档
- `DELETE /api/document/:id` - 删除画板文档
- `POST /api/node` - 创建图形节点
- `PUT /api/node/:id` - 更新节点
- `DELETE /api/node/:id` - 删除节点

### 4. 多人实时协作

支持团队成员和 Agent 同时在画板上协作。

- **房间管理** - 每个画板对应一个协作房间，支持多房间隔离
- **实时同步** - 用户和 Agent 的操作实时广播给其他成员
- **在线状态** - 显示当前在线成员和编辑状态
- **操作广播** - 节点创建、移动、删除等操作的实时推送
- **冲突处理** - 协作冲突的检测和处理

**核心接口：**
- `GET /ws` - 建立 WebSocket 连接
- `POST /api/room/join` - 加入画板协作房间
- `POST /api/room/leave` - 离开房间
- `POST /api/room/broadcast` - 广播操作消息

## 🔧 开发指南

### 添加新的 AI 提供商

1. 在 `service/` 目录下创建新的提供商实现
2. 实现 `AIProvider` 接口：
```go
type AIProvider interface {
    Chat(ctx context.Context, req *model.AIStreamRequest, resp *app.RequestContext) error
}
```
3. 在 `handler/ai_stream_handler.go` 中注册新提供商

### 扩展 WebSocket 功能

1. 在 `ws/hub.go` 中添加新的消息类型
2. 在 `ws/client.go` 中处理新的消息
3. 更新 `handler/ws_handler.go` 处理逻辑

### 数据模型迁移

使用 GORM 的 AutoMigrate 功能：
```go
database.DB.AutoMigrate(&model.YourModel{})
```

## 📄 技术文档

项目包含详细的技术文档：

- [AI_STREAM_IMPLEMENTATION.md](AI_STREAM_IMPLEMENTATION.md) - AI 流式对话实现方案
- [CONVERSATION_PERSISTENCE.md](CONVERSATION_PERSISTENCE.md) - 对话持久化技术方案
- [WEBSOCKET_ROOM_IMPLEMENTATION.md](WEBSOCKET_ROOM_IMPLEMENTATION.md) - WebSocket 房间管理方案
- [AI_TOOL_SELECTION_IMPLEMENTATION.md](AI_TOOL_SELECTION_IMPLEMENTATION.md) - AI 工具选择实现方案


## 🤝 开发规范

- 遵循 Go 代码规范和最佳实践
- 使用有意义的命名和注释
- 保持代码简洁和模块化
- 错误处理要完整和清晰

## 📝 License

MIT License
