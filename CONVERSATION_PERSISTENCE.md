# AI 对话上下文持久化技术方案

## 1. 方案概述

### 1.1 目标
- 用户每轮对话（user → assistant）全部持久化到 MySQL
- 用户重新登录/切换对话后能完整恢复历史上下文
- 支持多会话管理（会话列表、切换、删除）
- 支持 Tool Calling 消息持久化（tool_calls + tool results）

### 1.2 技术栈
- **语言**: Go 1.22+
- **框架**: Hertz (HTTP/SSE)
- **ORM**: GORM
- **数据库**: MySQL 8.0+
- **消息去重**: 基于 messageId（前端生成或后端 UUID）

---

## 2. 数据库表设计

### 2.1 Conversation（会话表）

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| id | INT | 自增主键 | PRIMARY KEY |
| conversation_id | VARCHAR(255) | 会话唯一ID（UUID） | UNIQUE INDEX |
| userid | VARCHAR(255) | 用户ID | INDEX |
| title | VARCHAR(500) | 会话标题 | |
| create_time | DATETIME | 创建时间 | AUTO |
| modify_time | DATETIME | 修改时间 | AUTO |

**索引设计**：
```sql
CREATE UNIQUE INDEX unique_index_conversation_id ON Conversation(conversation_id);
CREATE INDEX idx_userid_modifytime ON Conversation(userid, modify_time DESC);
```

---

### 2.2 ChatMessage（消息表）

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| id | INT | 自增主键 | PRIMARY KEY |
| message_id | VARCHAR(255) | 消息唯一ID | UNIQUE INDEX |
| conversation_id | VARCHAR(255) | 关联会话ID | INDEX |
| model_id | VARCHAR(255) | Agent/模型ID | |
| role | TINYINT | 角色（0=user, 1=assistant, 2=tool） | NOT NULL |
| content | TEXT | 消息内容 | |
| tool_calls | TEXT | 工具调用JSON（可为NULL） | |
| seq | INT | 会话内序号（从1递增） | NOT NULL |
| create_time | DATETIME | 创建时间 | AUTO |

**唯一约束**：
```sql
CREATE UNIQUE INDEX unique_index_message_id ON ChatMessage(message_id);
CREATE UNIQUE INDEX unique_index_conversation_seq ON ChatMessage(conversation_id, seq);
```

**说明**：
- `seq` 字段保证消息顺序，同一会话内唯一递增
- `tool_calls` 为 NULL 或 JSON 字符串（`*string` 类型）
- `role`: 0=user, 1=assistant, 2=tool

---

## 3. API 接口设计

### 3.1 创建会话
```
POST /api/conversation/create
```

**Request Body:**
```json
{
  "userId": "user_001",
  "title": "新对话"  // 可选
}
```

**Response:**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "conversationId": "uuid-xxx",
    "userId": "user_001",
    "title": "新对话",
    "createTime": "2026-02-22T10:00:00Z"
  }
}
```

---

### 3.2 获取会话列表
```
GET /api/conversation/list?userId=user_001&limit=50
```

**Response:**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "conversations": [
      {
        "conversationId": "uuid-xxx",
        "userId": "user_001",
        "title": "电商项目架构图",
        "createTime": "2026-02-22T10:00:00Z",
        "modifyTime": "2026-02-22T11:30:00Z"
      }
    ],
    "total": 15
  }
}
```

---

### 3.3 获取对话消息历史
```
GET /api/conversation/:conversationId/messages
```

**Response:**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "conversationId": "uuid-xxx",
    "messages": [
      {
        "messageId": "msg-001",
        "role": "user",
        "content": "画一个电商项目架构图",
        "createTime": "2026-02-22T10:00:00Z"
      },
      {
        "messageId": "msg-002",
        "role": "assistant",
        "content": "我来帮你创建...",
        "toolCalls": null,
        "createTime": "2026-02-22T10:00:05Z"
      }
    ],
    "total": 4
  }
}
```

---

### 3.4 AI 流式对话接口（已集成持久化）
```
POST /api/ai-stream/chat
```

**Request Body:**
```json
{
  "conversationId": "uuid-xxx",  // 可选，不传则不持久化
  "agentId": "agent_deepseek_001",
  "messages": [
    {
      "messageId": "msg-001",  // 可选，用于去重
      "role": "user",
      "content": "画一个架构图"
    }
  ],
  "config": {
    "provider": "deepseek",
    "model": "deepseek-chat",
    "temperature": 0.7,
    "maxTokens": 2000,
    "systemMessage": "你是AI助手"
  }
}
```

**流程变更**：
1. 如果 `conversationId` 存在且不为空：
   - 调用 `EnsureConversationExists` 确保会话存在
   - 调用 `SaveUserMessages` 保存用户消息（基于 messageId 去重）
   - AI 响应完成后调用 `SaveAssistantMessage` 保存

2. 如果 `conversationId` 为空：
   - 不进行持久化操作

---

## 4. 核心实现逻辑

### 4.1 消息去重机制

**问题**：前端可能传递完整历史消息数组，导致重复保存

**解决方案**：基于 `messageId` 去重

```go
// 1. 查询会话中已存在的 messageId
existingIds := getExistingMessageIds(conversationID)

// 2. 过滤出新消息
for _, msg := range messages {
    // 如果前端没传 messageId，后端生成
    if msg.MessageID == "" {
        msg.MessageID = uuid.New().String()
    }
    
    // 去重：只保存不存在的消息
    if !existingIds[msg.MessageID] {
        newMessages = append(newMessages, msg)
    }
}
```

**重要约定**：
- 前端应为每条消息生成唯一的 `messageId`（UUID）
- 多轮对话时，历史消息保持相同的 `messageId`
- 如果前端不传 `messageId`，后端自动生成（但可能导致重复）

---

### 4.2 seq 序号管理

**目的**：保证消息顺序，避免并发冲突

**实现**：
```go
// 1. 查询当前会话的最大 seq
var maxSeq int
db.Model(&ChatMessageRecord{}).
    Where("conversation_id = ?", conversationID).
    Select("COALESCE(MAX(seq), 0)").
    Scan(&maxSeq)

// 2. 为每条新消息分配递增的 seq
for _, msg := range newMessages {
    maxSeq++
    record.Seq = maxSeq
    // 保存...
}
```

---

### 4.3 消息保存时机

#### 时机1：用户发送消息时（同步保存）
```
Client → POST /api/ai-stream/chat
↓
Handler: ai_stream_handler.go
↓
if conversationId != "" && agentId != "":
  conversationService.EnsureConversationExists()
  messageService.SaveUserMessages()  // 保存 user 消息
↓
调用 AI Provider 流式返回
```

#### 时机2：AI 响应完成时（异步保存）
```
DeepSeekProvider.handleStream()
↓
接收完所有 SSE chunks
↓
累积完整 content 和 toolCalls
↓
if conversationId != "" && agentId != "":
  messageService.SaveAssistantMessage()  // 保存 assistant 消息
```

---

## 5. 代码架构

### 5.1 Model 层
```go
// model/conversation.go
type Conversation struct {
    ID             int
    ConversationID string
    UserID         string
    Title          string
    CreateTime     time.Time
    ModifyTime     time.Time
}

type ChatMessageRecord struct {
    ID             int
    MessageID      string
    ConversationID string
    ModelID        string
    Role           int8      // 0=user, 1=assistant, 2=tool
    Content        string
    ToolCalls      *string   // 可为 NULL
    Seq            int
    CreateTime     time.Time
}
```

---

### 5.2 Service 层
```go
// service/conversation_service.go
type ConversationService struct{}

func (s *ConversationService) CreateConversation(conversationID, userID, title string) (*Conversation, error)
func (s *ConversationService) GetByConversationID(conversationID string) (*Conversation, error)
func (s *ConversationService) EnsureConversationExists(conversationID, userID, agentID string) error
func (s *ConversationService) ListByUserID(userID string, limit int) ([]Conversation, error)
```

```go
// service/message_service.go
type MessageService struct{}

func (s *MessageService) SaveUserMessages(conversationID string, messages []ChatMessage, agentID string) error
func (s *MessageService) SaveAssistantMessage(conversationID, agentID, content string, toolCalls []ToolCall) error
func (s *MessageService) GetMessagesByConversationID(conversationID string) ([]ChatMessageRecord, error)
func (s *MessageService) getExistingMessageIds(conversationID string) map[string]bool
```

---

### 5.3 Handler 层
```go
// handler/conversation_handler.go
type ConversationHandler struct {
    conversationService *ConversationService
    messageService      *MessageService
}

func (h *ConversationHandler) GetMessages(ctx, c)       // GET /:conversationId/messages
func (h *ConversationHandler) CreateConversation(ctx, c) // POST /create
func (h *ConversationHandler) ListConversations(ctx, c)  // GET /list
```

```go
// handler/ai_stream_handler.go
func (h *AIStreamHandler) StreamChat(ctx, c) {
    // 1. 解析请求
    // 2. 查询 Agent 配置
    // 3. 查询 Tools
    
    // 4. 对话持久化：保存用户消息
    if req.ConversationID != "" && req.AgentID != "" {
        conversationService.EnsureConversationExists(req.ConversationID, userID, req.AgentID)
        messageService.SaveUserMessages(req.ConversationID, req.Messages, req.AgentID)
    }
    
    // 5. 调用 AI Service 流式返回
    aiService.StreamChat(ctx, &req, c)
}
```

```go
// service/deepseek_provider.go
func (p *DeepSeekProvider) handleStream(body, c, req) {
    // 1. 流式解析 SSE
    // 2. 累积 content 和 toolCalls
    // 3. 发送 complete 事件
    
    // 4. 保存 Assistant 消息
    if req.ConversationID != "" && req.AgentID != "" {
        messageService.SaveAssistantMessage(
            req.ConversationID,
            req.AgentID,
            fullContent.String(),
            toolCalls,
        )
    }
}
```

---

## 6. 前端对接指南

### 6.1 消息 messageId 管理

**推荐方案**：前端为每条消息生成唯一 UUID

```typescript
import { v4 as uuidv4 } from 'uuid';

// 发送消息时
const userMessage = {
  messageId: uuidv4(),  // 生成唯一ID
  role: 'user',
  content: userInput
};

// 多轮对话时，保持历史消息的 messageId 不变
const messages = [
  { messageId: 'msg-001', role: 'user', content: '...' },
  { messageId: 'msg-002', role: 'assistant', content: '...' },
  { messageId: 'msg-003', role: 'user', content: '新消息' }  // 只有这条是新的
];
```

---

### 6.2 会话管理流程

#### 新建会话
```typescript
// 1. 调用创建会话接口
const response = await fetch('/api/conversation/create', {
  method: 'POST',
  body: JSON.stringify({ userId: 'user_001', title: '新对话' })
});
const { conversationId } = await response.json();

// 2. 保存 conversationId 到状态
setCurrentConversationId(conversationId);
```

#### 切换会话
```typescript
// 1. 加载会话历史
const response = await fetch(`/api/conversation/${conversationId}/messages`);
const { messages } = await response.json();

// 2. 渲染历史消息
setMessages(messages);
setCurrentConversationId(conversationId);
```

#### 继续对话
```typescript
// 发送消息时带上 conversationId
const response = await fetch('/api/ai-stream/chat', {
  method: 'POST',
  body: JSON.stringify({
    conversationId: currentConversationId,  // 关键参数
    agentId: 'agent_deepseek_001',
    messages: [...historyMessages, newUserMessage]
  })
});
```

---

## 7. 测试验证

### 7.1 功能测试用例

**测试1：创建对话并保存消息**
```bash
# 1. 创建会话
POST /api/conversation/create
Body: {"userId": "test_user", "title": "测试对话"}

# 2. 发送消息
POST /api/ai-stream/chat
Body: {
  "conversationId": "返回的conversationId",
  "agentId": "default_agent",
  "messages": [{"role": "user", "content": "你好"}]
}

# 3. 验证数据库
SELECT * FROM ChatMessage WHERE conversation_id = '...';
-- 应该有 2 条记录（user + assistant）
```

---

**测试2：多轮对话去重**
```bash
# 第二轮发送完整历史
POST /api/ai-stream/chat
Body: {
  "conversationId": "同一个id",
  "messages": [
    {"messageId": "msg-001", "role": "user", "content": "你好"},
    {"messageId": "msg-002", "role": "assistant", "content": "你好"},
    {"messageId": "msg-003", "role": "user", "content": "第二条"}  # 新消息
  ]
}

# 验证：数据库应该只新增 2 条（msg-003 的 user + assistant）
-- 不会重复保存 msg-001 和 msg-002
```

---

**测试3：查询历史消息**
```bash
GET /api/conversation/{conversationId}/messages

# 验证返回顺序
-- messages 应该按 seq 升序排列
-- role 已转换为字符串（"user", "assistant"）
-- toolCalls 正确解析为 JSON 对象
```

---

### 7.2 数据库验证

```sql
-- 查看会话列表
SELECT conversation_id, userid, title, create_time 
FROM Conversation 
WHERE userid = 'test_user' 
ORDER BY create_time DESC;

-- 查看消息记录
SELECT message_id, role, seq, 
       CASE role WHEN 0 THEN 'user' WHEN 1 THEN 'assistant' WHEN 2 THEN 'tool' END as role_name,
       LEFT(content, 50) as content_preview,
       create_time
FROM ChatMessage 
WHERE conversation_id = 'xxx' 
ORDER BY seq ASC;

-- 验证 seq 唯一性
SELECT conversation_id, seq, COUNT(*) 
FROM ChatMessage 
GROUP BY conversation_id, seq 
HAVING COUNT(*) > 1;
-- 应该返回空结果
```

---

## 8. 技术难点与解决方案

### 难点1：消息重复保存
**问题**：前端传完整历史，后端每次都保存导致重复

**解决方案**：
- 前端为每条消息生成唯一 `messageId`
- 后端查询已存在的 `messageId` 进行去重
- 只保存新消息

---

### 难点2：seq 并发冲突
**问题**：并发请求可能生成相同 seq

**解决方案**：
- 利用数据库唯一约束 `UNIQUE(conversation_id, seq)`
- 如果插入失败，重新查询最大 seq 并重试
- 或使用数据库锁：`SELECT MAX(seq) FOR UPDATE`

---

### 难点3：tool_calls 字段 JSON 格式
**问题**：空数组会导致 MySQL JSON 验证错误

**解决方案**：
- 将 `ToolCalls` 定义为 `*string`（指针类型）
- 当 `len(toolCalls) == 0` 时，保持 NULL
- 只有真正有工具调用时才赋值 JSON 字符串

---

## 9. 总结

**已完成功能**：
- ✅ 数据库表设计（Conversation + ChatMessage）
- ✅ Service 层实现（会话管理 + 消息管理）
- ✅ Handler 层实现（REST API）
- ✅ AI Stream 集成（自动保存 user + assistant 消息）
- ✅ 消息去重机制（基于 messageId）
- ✅ seq 序号管理（保证顺序）
- ✅ tool_calls 支持（NULL 安全）

**关键特性**：
- 📝 完整记录每轮对话（user → assistant）
- 🔄 支持多轮对话历史恢复
- 🎯 自动去重（避免重复保存）
- 📊 seq 排序（保证消息顺序）
- 🛠️ Tool Calling 支持（tool_calls 字段）

**前端对接要点**：
- 每条消息必须有唯一 `messageId`（UUID）
- 发送时带上 `conversationId`（不传则不持久化）
- 多轮对话传完整历史，后端自动去重

**下一步优化方向**：
- Redis 缓存（减少数据库查询）
- 分页加载历史消息（长对话优化）
- 会话标题自动生成（AI 总结）
- 会话归档/搜索功能
