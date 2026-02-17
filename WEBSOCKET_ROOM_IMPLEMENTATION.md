# WebSocket Room 实时协作实现文档

## 概述
在 tenet-server 中实现了基于 WebSocket 的实时协作功能，采用 Hub-Client 模式管理房间和连接，支持多用户实时编辑、消息广播、用户加入/离开通知等核心能力。

**核心特性**：
- 基于 Hertz WebSocket 的长连接管理
- Hub 模式集中管理所有房间和连接
- 按 documentId 隔离房间（Room）
- 消息广播支持排除发送者
- 非 streaming 消息自动持久化到数据库
- 用户加入/离开实时通知
- Ping/Pong 心跳保活机制

**技术栈**：
- HTTP 框架：Hertz v0.9.0
- WebSocket 库：hertz-contrib/websocket
- 并发模型：Go Channels + Goroutines
- 数据库：MySQL (GORM)

---

## 架构设计

### 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      Tenet Server                            │
│                                                              │
│  ┌────────────┐         ┌──────────────────────────┐        │
│  │   main.go  │         │         Hub              │        │
│  │            │         │  (单 goroutine 运行)      │        │
│  │ go hub.Run()├────────►  - Register chan         │        │
│  └────────────┘         │  - Unregister chan       │        │
│                         │  - Broadcast chan        │        │
│  ┌────────────┐         │  - rooms map             │        │
│  │ ws_handler │         └───────────┬──────────────┘        │
│  │            │                     │                       │
│  │ Upgrade()  │◄────────────────────┤                       │
│  └─────┬──────┘                     │                       │
│        │                            │                       │
│        │ create Client              │                       │
│        ▼                            ▼                       │
│  ┌──────────────────────────────────────────┐              │
│  │           Client (per connection)        │              │
│  │  - ReadPump() goroutine (读取消息)       │              │
│  │  - WritePump() goroutine (发送消息)      │              │
│  │  - send chan (消息队列)                  │              │
│  │  - msgHandler (消息处理器)               │              │
│  └──────────────────────────────────────────┘              │
└─────────────────────────────────────────────────────────────┘
         ▲                                    │
         │ WebSocket                          │ WebSocket
         │ Connect                            │ Messages
         │                                    ▼
    ┌────────────┐                      ┌────────────┐
    │  Browser 1 │                      │  Browser 2 │
    └────────────┘                      └────────────┘
```

### 三层架构

```
┌──────────────────────────────────────────────────────────┐
│  Handler 层 (handler/ws_handler.go)                       │
│  - 处理 WebSocket 升级                                     │
│  - 解析 query 参数 (documentId, userId)                   │
│  - 创建 Client 并注册到 Hub                                │
└──────────────────────────────────────────────────────────┘
                          ▼
┌──────────────────────────────────────────────────────────┐
│  Hub 层 (ws/hub.go)                                       │
│  - 房间管理 (map[documentId]map[*Client]bool)            │
│  - 连接注册/注销                                          │
│  - 消息广播路由                                           │
│  - 用户加入/离开通知                                       │
└──────────────────────────────────────────────────────────┘
                          ▼
┌──────────────────────────────────────────────────────────┐
│  Client 层 (ws/client.go)                                 │
│  - ReadPump: 接收前端消息 → 处理 → 广播                   │
│  - WritePump: 从 send 通道读取 → 发送到前端               │
│  - 消息处理器: 解析消息类型 → 调用 Service → 持久化        │
└──────────────────────────────────────────────────────────┘
                          ▼
┌──────────────────────────────────────────────────────────┐
│  Service/DAO 层                                           │
│  - NodeService: 节点 CRUD 业务逻辑                        │
│  - NodeDAO: 数据库操作                                     │
└──────────────────────────────────────────────────────────┘
```

---

## 核心实现

### 1. Hub：中心调度器

#### 文件位置
`ws/hub.go`

#### 核心职责
- **房间管理**：按 documentId 隔离不同文档的连接
- **连接管理**：处理客户端注册/注销
- **消息路由**：根据 RoomId 广播消息给房间内所有客户端
- **用户通知**：广播用户加入/离开事件

#### 数据结构

```go
type Hub struct {
    // 房间管理（documentId -> clients）
    rooms map[string]map[*Client]bool
    
    // 注册请求（缓冲 256）
    Register chan *Client
    
    // 注销请求（缓冲 256）
    Unregister chan *Client
    
    // 广播消息（缓冲 256）
    Broadcast chan *BroadcastMessage
}
```

**关键设计点**：
- **有缓冲通道**：避免 Hub 自己向自己发送消息时死锁
- **嵌套 Map**：`rooms[documentId][*Client]` 实现房间隔离
- **使用 map 作 set**：`map[*Client]bool` 快速判断 client 存在性

#### 核心方法：Run()

**执行流程**：

**1. Register 处理（用户加入）**
- 检查房间是否存在，不存在则创建
- 将 Client 添加到房间的 clients map
- 构造 `user_join` 消息
- 发送到 Broadcast 通道（排除新用户自己）

**2. Unregister 处理（用户离开）**
- 先广播 `user_leave` 消息给房间内其他人
- 从房间 map 中删除该 Client
- 关闭 Client 的 send 通道
- 如果房间为空，删除整个房间

**3. Broadcast 处理（消息广播）**
- 根据 RoomId 找到目标房间
- 遍历房间内所有 Client
- 排除 ExcludeId（发送者自己）
- 使用 select-default 非阻塞发送到每个 client.send 通道
- 如果发送失败（通道满），清理该连接

**关键实现点**：
1. **非阻塞发送**：使用 `select-default` 避免慢客户端阻塞整个房间
2. **先广播后删除**：用户离开时先通知其他人，再清理资源
3. **房间自动清理**：最后一个用户离开时自动删除房间
4. **排除发送者**：广播时通过 `ExcludeId` 避免回显

#### 查询方法

```go
// 获取房间用户数
func (h *Hub) GetRoomClients(documentId string) int {
    if clients, ok := h.rooms[documentId]; ok {
        return len(clients)
    }
    return 0
}

// 获取房间用户列表
func (h *Hub) GetRoomUsers(documentId string) []map[string]string {
    var users []map[string]string
    if clients, ok := h.rooms[documentId]; ok {
        for client := range clients {
            users = append(users, map[string]string{
                "userId": client.userId,
                "clientId": client.id,
            })
        }
    }
    return users
}
```

---

### 2. Client：连接管理器

#### 文件位置
`ws/client.go`

#### 核心职责
- **双向通信**：ReadPump 读取前端消息，WritePump 发送消息到前端
- **消息处理**：解析消息类型，调用业务逻辑，持久化数据
- **心跳保活**：通过 Ping/Pong 机制保持连接活跃
- **资源清理**：连接断开时自动注销

#### 数据结构

```go
type Client struct {
    id         string              // 客户端唯一ID
    hub        *Hub                // Hub 引用
    conn       *websocket.Conn     // WebSocket 连接
    send       chan []byte         // 发送队列（缓冲 256）
    documentId string              // 所在房间
    userId     string              // 用户ID
    msgHandler *MessageHandler     // 消息处理器
}
```

**关键设计点**：
- **send 通道**：Hub 向 Client 发送消息的缓冲队列
- **msgHandler**：每个 Client 独立的消息处理器实例
- **双向引用**：Client 持有 Hub 引用，用于注销和广播

#### 核心方法：ReadPump()

**执行流程**：

**初始化阶段**
- 设置 defer 清理：连接断开时自动发送 Unregister 请求
- 设置读超时（60秒）
- 注册 Pong 处理器：收到 Pong 时重置超时计时

**消息循环**
1. 从 WebSocket 读取消息
2. JSON 解析为 Message 结构
3. 检查 `streaming` 标志：
   - false：调用 msgHandler 持久化到数据库
   - true：跳过持久化（用于 AI 流式输出）
4. 发送到 Hub.Broadcast 通道（广播给房间内其他人）
5. 构造 ACK 响应并发送到 client.send 通道

**异常处理**
- 读取错误：退出循环，触发 defer 清理
- 解析错误：跳过当前消息，继续下一条

**关键实现点**：
1. **defer 清理**：确保连接断开时自动注销
2. **Pong 处理器**：收到 Pong 时重置读超时
3. **streaming 标志**：控制是否持久化到数据库
4. **ACK 机制**：同步反馈操作结果

#### 核心方法：WritePump()

```go
func (c *Client) WritePump() {
    ticker := time.NewTicker(pingPeriod)  // 54秒
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()
    
    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                // 通道关闭，发送关闭帧
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            
            // 写入消息
            w, err := c.conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            w.Write(message)
            
            // 批量发送缓冲区中的消息（优化）
            n := len(c.send)
            for i := 0; i < n; i++ {
                w.Write([]byte{'\n'})
                w.Write(<-c.send)
            }
            
            if err := w.Close(); err != nil {
                return
**执行流程**：

**初始化阶段**
- 创建 Ping 定时器（54秒间隔）
- 设置 defer 清理：关闭定时器和连接

**消息循环（select 多路复用）**

**Case 1: 从 send 通道接收消息**
- 设置写超时（10秒）
- 检查通道是否关闭：关闭则发送 CloseMessage 并退出
- 写入当前消息
- **批量优化**：一次性取出 send 通道中积压的所有消息并发送
- 减少系统调用次数，提升性能

**Case 2: Ping 定时器触发**
- 设置写超时（10秒）
- 发送 PingMessage 保持连接活跃
- 前端收到后会自动回复 PongMessage switch msg.Meta.Type {
    case "node_create":
        return h.handleNodeCreate(msg)
    case "node_update":
        return h.handleNodeUpdate(msg)
    case "node_delete":
        return h.handleNodeDelete(msg)
    default:
        log.Printf("Unknown message type: %s", msg.Meta.Type)
        return nil
    }
}

func (h *MessageHandler) handleNodeCreate(msg *Message) error {
    if msg.Data.Node == nil {
        return nil
    }
    
    node := msg.Data.Node
    node.DocumentId = msg.Meta.DocumentId  // 设置文档ID
    
    // 调用 Service 层保存节点
    if err := h.nodeService.CreateNode(node); err != nil {
       处理流程

**HandleMessage 方法**
- 根据 `msg.Meta.Type` 路由到不同的处理方法
- 支持类型：`node_create`, `node_update`, `node_delete`
- 系统消息（`user_join`, `user_leave`）不需要处理，仅广播

**处理步骤**（以 node_create 为例）
1. 校验 `msg.Data.Node` 是否存在
2. 从 `msg.Meta` 中提取 `DocumentId` 并注入到 Node
3. 调用 `NodeService.CreateNode()` 保存到数据库
4. 返回错误（如有），最终反馈到 ACK 响应     
        if documentId == "" || userId == "" {
            c.JSON(400, model.Error(400, "缺少 documentId 或 userId 参数"))
            return
        }
        
        // 升级为 WebSocket
        err := upgrader.Upgrade(c, func(conn *websocket.Conn) {
            // 创建客户端
            client := ws.NewClient(hub, conn, documentId, userId)
            
            // 注册到 Hub
            hub.Register <- client
            
            // 启动读写协程
       处理流程

**HandleWebSocket 执行步骤**

1. **参数提取与校验**
   - 从 query 参数获取 `documentId` 和 `userId`
   - 校验必填参数，缺失则返回 400 错误

2. **协议升级**
   - 调用 `upgrader.Upgrade()` 将 HTTP 升级为 WebSocket
   - 传入升级成功后的回调函数

3. **Client 创建与注册**
   - 创建 Client 实例（包含 hub、conn、documentId、userId）
   - 发送到 `hub.Register` 通道（Hub 会将其加入房间）

4. **Goroutine 启动**
   - `go client.WritePump()`：后台 goroutine 处理发送
   - `client.ReadPump()`：当前 goroutine 阻塞处理接收
   - ReadPump 结束时自动触发清理

## 消息协议

### 基础消息结构

```go
type Message struct {
    Meta MessageMeta `json:"meta"`
    Data MessageData `json:"data"`
}

type MessageMeta struct {
    UserId     string `json:"userId"`      // 用户ID
    DocumentId string `json:"documentId"`  // 文档ID
    OpId       string `json:"opId"`        // 操作ID（用于ACK）
    Type       string `json:"type"`        // 消息类型
    Streaming  bool   `json:"streaming"`   // 是否流式（不落库）
}

type MessageData struct {
    Node *model.Node `json:"node,omitempty"`  // 节点数据
}
**开发环境**：允许所有来源（`return true`）

**生产环境**：需校验 Origin 头，只允许可信域名
- 从请求头提取 `Origin`
- 判断是否为白名单域名（如 `https://oolong.com`）
- 防止 CSRF 攻击 ACK 确认

```go
type Ack struct {
    OpId    string `json:"opId"`               // 对应的操作ID
    Success bool   `json:"success"`            // 是否成功
    Error   string `json:"error,omitempty"`    // 错误信息
}
```

### 消息示例

#### 创建节点

**客户端发送：**
```json
{
  "meta": {
    "userId": "u123",
    "documentId": "doc_7H5OuU_XlKKsEpWw",
    "opId": "op001",
    "type": "node_create",
    "streaming": false
  },
  "data": {
    "node": {
      "nodeId": "node_001",
      "type": 0,
      "parentId": "",
      "zIndex": 1.0,
      "capInfo": "{\"x\":100,\"y\":200,\"text\":\"Hello\"}"
    }
  }
}
```

**服务器返回 ACK：**
```json
{
  "opId": "op001",
  "success": true
}
```

**房间内其他用户收到（广播）：**
```json
{
  "meta": {
    "userId": "u123",
    "documentId": "doc_7H5OuU_XlKKsEpWw",
    "opId": "op001",
    "type": "node_create",
    "streaming": false
  },
  "data": {
    "node": {
      "nodeId": "node_001",
      "type": 0,
      ...
    }
  }
}
```

#### 用户加入/离开

**用户加入广播：**
```json
{
  "meta": {
    "userId": "u124",
    "documentId": "doc_7H5OuU_XlKKsEpWw",
    "type": "user_join"
  },
  "data": {}
}
```

**用户离开广播：**
```json
{
  "meta": {
    "userId": "u124",
    "documentId": "doc_7H5OuU_XlKKsEpWw",
    "type": "user_leave"
  },
  "data": {}
}
```

---

## HTTP API

### 查询房间用户

**端点**：`GET /api/room/:documentId/users`

**请求示例**：
```
GET /api/room/doc_7H5OuU_XlKKsEpWw/users
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "documentId": "doc_7H5OuU_XlKKsEpWw",
    "userCount": 2,
    "users": [
      {
        "userId": "u123",
        "clientId": "20260217230722_22222222"
      },
      {
        "userId": "u124",
        "clientId": "20260217230725_e666yyyy"
      }
    ]
  }
}
```

**实现位置**：
- Handler: `handler/room_handler.go`
- 路由: `router/router.go`

---

## 并发模型

### Goroutine 架构

```
main goroutine
  │
  ├─► Hub.Run() goroutine (1个)
  │     └─ 处理所有 Register/Unregister/Broadcast 事件
  │
  ├─► Client1.ReadPump() goroutine
  ├─► Client1.WritePump() goroutine
  │
  ├─► Client2.ReadPump() goroutine
  ├─► Client2.WritePump() goroutine
  │
  └─► ... (每个连接 2 个 goroutines)
```

**总 goroutine 数 = 1 (Hub) + 2N (N 个客户端)**

### Channel 通信

```
WebSocket Goroutine          Hub Goroutine           Client Goroutine
       │                          │                         │
       │  hub.Register <- client  │                         │
       ├─────────────────────────►│                         │
       │                          │                         │
       │                     处理注册                        │
       │                          │                         │
       │                          │  hub.Broadcast <- msg   │
       │                          │◄────────────────────────┤
       │                          │                         │
       │                      广播到房间                     │
       │                          │                         │
       │                          │  client.send <- data    │
       │                          ├────────────────────────►│
       │                          │                         │
       │                          │                 WritePump 发送
```

### 关键并发问题与解决

#### 问题1：无缓冲通道死锁

**现象**：
- Hub.Run() 在 case Register 中执行 `h.Broadcast <- msg`
- 但此时 Hub 无法同时处理 case Broadcast
- 导致发送阻塞，永远等待接收者

**解决**：
```go
// 修改前（无缓冲）
Register:   make(chan *Client)
Broadcast:  make(chan *BroadcastMessage)

// 修改后（有缓冲）
Register:   make(chan *Client, 256)
Broadcast:  make(chan *BroadcastMessage, 256)
```

**原理**：
- 无缓冲通道：发送和接收必须同时准备好（握手）
- 有缓冲通道：发送到缓冲区后立即返回，不等待接收者
 ⭐️

**问题现象**
- 第二个用户连接时卡住，无法加入房间
- Hub.Run() 在处理 Register 事件时阻塞

**根本原因**
- Hub.Run() 是单 goroutine 的 select 循环
- 在 Register case 中执行 `h.Broadcast <- msg`
- 由于通道无缓冲，发送操作等待接收者
- 但接收者是 Hub 自己的 Broadcast case
- Hub 正在执行 Register case，无法同时处理 Broadcast case
- **自己等待自己 → 死锁**

**解决方案**
- 将所有 Hub 通道改为有缓冲（容量 256）
- Register、Unregister、Broadcast 通道都添加缓冲
- 发送到缓冲通道立即返回，无需等待接收者
- Hub 在后续 select 循环中处理缓冲区中的消息

**核心原理**
- **无缓冲通道** = 握手模式（发送和接收同时就绪）
- **有缓冲通道** = 队列模式（发送到队列即可，异步处理）
- 避免单 goroutine 内自己向自己发送时的循环等待

#### 问题2：慢客户端阻塞整个房间

**问题现象**
- 房间内某个客户端网络慢，send 通道满（256 条消息积压）
- Hub 广播时向该客户端发送阻塞
- 导致整个广播循环卡住，其他用户也收不到消息

**解决方案**
- 使用 `select-default` 非阻塞发送
- 如果 send 通道满，直接进入 default 分支
- 清理该慢客户端连接（关闭通道、删除 client）
- 继续向其他客户端广播

**核心原理**
- **阻塞发送**：`ch <- data` 会一直等待直到通道有空间
- **非阻塞发送**：`select { case ch <- data: ... default: ... }` 立即返回
- 牺牲个别慢客户端，保证整体房间可用性 []byte, 256)

// Hub 通道
Register:   make(chan *Client, 256)
Unregister: make(chan *Client, 256)
Broadcast:  make(chan *BroadcastMessage, 256)
```

---

## 部署架构

### 生产环境拓扑

```
┌────────────┐
│   Nginx    │  (WebSocket 代理 + 负载均衡)
│  :443/wss  │
└──────┬─────┘
       │
       ├────────────┬────────────┐
       ▼            ▼            ▼
┌─────────┐  ┌─────────┐  ┌─────────┐
│ Server1 │  │ Server2 │  │ Server3 │
│  :8081  │  │  :8081  │  │  :8081  │
└────┬────┘  └────┬────┘  └────┬────┘
     │            │            │
     └────────────┴────────────┘
                  │
            ┌─────▼─────┐
            │   MySQL   │
            │   :3306   │
            └───────────┘
```

**注意事项**：
1. **会话亲和性**：需要配置 Sticky Session（基于 documentId）
2. **跨服务器协同**：当前实现为单机版，跨服务器需要 Redis Pub/Sub
3. **连接数限制**：监控并发连接数，设置合理的限流策略

### Nginx 配置示例

```nginx
upstream tenet_ws {
    ip_hash;  # 会话亲和性
    server 127.0.0.1:8081;
    server 127.0.0.1:8082;
    server 127.0.0.1:8083;
}

server {
    listen 443 ssl;
    server_name oolong.com;
    
    location /ws {
        proxy_pass http://tenet_ws;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        
        # WebSocket 超时配置
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
}
```

---

## 测试指南

### 单元测试

**测试 Hub 注册/注销：**
```go
func TestHubRegisterUnregister(t *testing.T) {
    hub := ws.NewHub()
    go hub.Run()
    
    // 模拟客户端
    client := &ws.Client{...}
    
    // 注册
    hub.Register <- client
    time.Sleep(100 * time.Millisecond)
    assert.Equal(t, 1, hub.GetRoomClients("doc1"))
    
    // 注销
    hub.Unregister <- client
    time.Sleep(100 * time.Millisecond)
    assert.Equal(t, 0, hub.GetRoomClients("doc1"))
}
```

### 集成测试流程**
- 启动 Hub.Run() goroutine
- 创建模拟 Client 并发送到 Register 通道
- 验证房间用户数从 0 → 1
- 发送到 Unregister 通道
- 验证房间用户数从 1 → 0  "data": {
       "node": {
         "nodeId": "node_001",
         "type": 0,
         "capInfo": "{\"x\":100,\"y\":200}"
       }
     }
   }
   ```

3. **验证响应**
   - 收到 ACK: `{"opId":"op001","success":true}`

4. **查询房间用户**
   ```
   GET http://localhost:8081/api/room/doc123/users
   ```

### 压力测试

**使用 wscat 模拟多客户端：**
```bash
# 连接 100 个客户端
for i in {1..100}; do
  wscat -c "ws://localhost:8081/ws?documentId=doc123&userId=user$i" &
done
```

**监控指标：**
- 并发连接数
- 消息延迟
- CPU/内存使用率
- Goroutine 数量

---

## 对比：Java vs Go 实现

| 特性 | Java (oolong-server) | Go (tenet-server) |
|-----|---------------------|------------------|
| **WebSocket 框架** | Spring STOMP | Hertz WebSocket |
| **并发模型** | 线程池 | Goroutines + Channels |
| **连接管理** | UserDocService (Map) | Hub (Map + Channels) |
| **消息路由** | @MessageMapping | Hub.Broadcast |
| **房间隔离** | sessionId → userId → docId | rooms[docId][client] |
| **广播机制** | SimpMessagingTemplate | select + send chan |
| **持久化** | Executor 模式 | MessageHandler |
| **用户通知** | WebSocketEventListener | Hub Register/Unregister |
| **心跳机制** | STOMP 自带 | 手动 Ping/Pong |

---

## 待优化项

### 1. 跨服务器协同（优先级：高）
**问题**：当前单机版，用户连接到不同服务器无法协作

**方案**：
- 使用 Redis Pub/Sub 实现跨服务器消息广播
- 每个服务器订阅 `room:{documentId}` 频道
- 发送消息时 Publish 到 Redis，所有服务器接收并分发

### 2. 连接认证（优先级：高）
**问题**：当前只验证参数存在性，未验证用户身份

**方案**：
- 接收 token query 参数
- 调用 UserService.CheckUserToken() 验证
- 不合法则拒绝升级

### 3. 消息持久化优化（优先级：中）
**问题**：每个操作都实时写数据库，高并发下压力大

**方案**：
- 批量写入：缓存一批操作后统一提交
- 异步写入：使用消息队列（Kafka/RabbitMQ）

### 4. 房间权限校验（优先级：中）
**问题**：任何用户都能加入任何房间

**方案**：
- 查询 DocumentPermission 表验证用户权限
- 只允许有 CanEdit 或 CanManage 权限的用户加入

### 5. 监控告警（优先级：中）
**指标**：
- 在线用户数
- 房间数
- 消息吞吐量
- Goroutine 数量
- Channel 缓冲区使用率

**工具**：
- Prometheus + Grafana
- 自定义 metrics 端点

---

## 总结

本文档详细梳理了 tenet-server WebSocket Room 功能的核心实现，包括：

1. **Hub-Client 模式**：中心化管理 + 分布式执行
2. **Channel 并发**：利用 Go 原生并发原语实现高效通信
3. **房间隔离**：按 documentId 实现多租户隔离
4. **非阻塞广播**：避免慢客户端影响整体性能
5. **自动清理**：defer + close 确保资源正确释放

**核心优势**：
- 代码简洁：约 500 行实现完整功能
- 高并发：Goroutine 轻量级，支持大量连接
- 低延迟：Channel 零拷贝，消息传递高效
- 易扩展：消息协议可扩展，支持多种业务类型

**未来方向**：
- 跨服务器协同（Redis Pub/Sub）
- 更细粒度的权限控制
- 消息持久化优化（批量/异步）
- 完善的监控告警体系
