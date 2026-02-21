# AI Tool Selection 工具筛选实现文档

## 概述
实现了两阶段工具调用优化机制，通过智能筛选工具显著降低 Token 消耗，同时保持 AI 工具调用的灵活性和准确性。

**核心特性**：
- 两阶段工具调用：筛选 → 执行
- Token 消耗优化 76%（2,050 vs 8,500 tokens）
- AI 智能工具选择，支持多轮对话
- 轻量级工具描述与完整工具定义分离

**技术栈**：
- Go 1.22+, Hertz v0.9.0
- MySQL (GORM)
- DeepSeek API

---

## 架构设计

### 两阶段调用流程

```
┌──────────────────────────────────────────────────────────────┐
│  阶段 1: 工具筛选 (Tool Selection)                            │
│  POST /api/ai-stream/select-tools                            │
├──────────────────────────────────────────────────────────────┤
│  • 输入：用户消息 + 17个轻量级工具（name + description）      │
│  • 处理：AI 分析需求，返回所需工具名称列表                    │
│  • 输出：{"selectedTools": ["tool1", "tool2"], "reasoning"}  │
│  • Token：~850 tokens (17工具 × 50 tokens)                  │
└──────────────────────────────────────────────────────────────┘
                            ↓
┌──────────────────────────────────────────────────────────────┐
│  阶段 2: 工具执行 (Tool Execution)                            │
│  POST /api/ai-stream/chat                                    │
├──────────────────────────────────────────────────────────────┤
│  • 输入：toolNames + 对话历史                                 │
│  • 处理：加载完整工具定义（含参数 schema）                    │
│  • 输出：AI 生成 toolCalls 及具体参数                         │
│  • Token：~1,200 tokens (2工具 × 500 tokens + 对话)         │
└──────────────────────────────────────────────────────────────┘

总消耗：2,050 tokens  vs  不优化：8,500 tokens
节省：76% 🎉
```

### 系统架构

```
Frontend                    Backend
   │                           │
   │  1. select-tools          │
   ├─────────────────────────► │ GetAllToolsBasicInfo()
   │                           │ buildToolSelectionPrompt()
   │                           │ AI 筛选工具
   │  ["tool1", "tool2"]       │
   │◄───────────────────────── │
   │                           │
   │  2. chat (toolNames)      │
   ├─────────────────────────► │ GetToolsByNames()
   │                           │ 完整工具定义
   │                           │ AI 生成 toolCalls
   │  toolCalls + args         │
   │◄───────────────────────── │
   │                           │
   │  3. 执行工具（浏览器）     │
   │  生成 tool result          │
   │                           │
   │  4. chat (tool result)    │
   ├─────────────────────────► │ 继续用同一批工具
   │                           │ AI 判断是否完成
   │  done / more toolCalls    │
   │◄───────────────────────── │
```

---

## 核心实现

### 1. 轻量级工具查询

**service/toolcall_service.go**

```go
// GetAllToolsBasicInfo 获取所有工具的轻量级信息
func (s *ToolCallService) GetAllToolsBasicInfo() ([]map[string]string, error) {
    // 只返回 name + description，不包含参数定义
    // 每个工具 ~50 tokens
}
```

### 2. 工具筛选 Handler

**handler/ai_stream_select_handler.go**

核心逻辑：
1. 查询 Agent 配置
2. 获取所有工具的轻量级信息
3. 构建工具筛选系统提示词
4. 调用 AI 返回选中的工具名称

**系统提示词模板**：
```
你是一个智能工具筛选助手。根据对话历史选择最合适的工具。

可用工具列表：
[{"name":"add_node","description":"添加节点"}, ...]

返回格式：
{
  "selectedTools": ["tool_name_1", "tool_name_2"],
  "reasoning": "选择理由"
}
```

### 3. 路由注册

**router/router.go**

```go
// 创建服务
agentService := service.NewAgentService()
toolCallService := service.NewToolCallService()

// 创建 Handler
aiSelectHandler := handler.NewAIStreamSelectHandler(
    agentService, toolCallService, aiService
)

// 注册路由
aiStream.POST("/select-tools", aiSelectHandler.SelectTools)
```

---

## 完整交互流程

### 单次用户输入的多轮工具调用

```
用户输入："给我画一个蛋炒饭流程图"
    │
    ├─► 1. select-tools (只调用 1 次) ⭐
    │      返回：["batch_add_nodes", "batch_add_lines"]
    │      保存到：currentToolIds
    │
    ├─► 2. chat (toolNames: currentToolIds)
    │      AI 返回：toolCall(batch_add_nodes, args)
    │
    ├─► 3. 前端执行工具，生成 tool result
    │
    ├─► 4. chat (toolNames: currentToolIds) ⭐ 使用同一批工具
    │      AI 返回：toolCall(batch_add_lines, args)
    │
    ├─► 5. 前端执行工具，生成 tool result
    │
    ├─► 6. chat (toolNames: currentToolIds) ⭐ 还是同一批
    │      AI 返回：done (不再需要工具)
    │
    └─► 结束
```

**关键规则**：
- **工具只筛选一次**：在用户输入时调用 select-tools
- **后续循环使用相同工具**：AI 多次调用工具都用同一批 toolNames
- **直到 AI 不再需要工具**：返回纯文本响应，结束本轮对话

### 新用户输入触发重新筛选

```
用户再次输入："把第一个节点改成红色"
    │
    ├─► 1. select-tools (重新筛选) ⭐
    │      传入完整对话历史（包含之前的工具调用）
    │      返回：["update_node_color"]
    │
    ├─► 2. chat (toolNames: ["update_node_color"])
    │      AI 返回：toolCall(update_node_color, {color: "red"})
    │
    └─► 结束
```

---

## Token 优化分析

### 对比数据

| 场景 | 工具筛选 | 工具执行 | 总计 | 节省 |
|------|---------|---------|------|------|
| **两阶段优化** | 850 | 1,200 | **2,050** | **76%** |
| 不优化（发送所有工具） | 0 | 8,500 | **8,500** | - |

### 为什么节省这么多？

**轻量级工具** (select-tools):
```json
{"name": "add_node", "description": "在画板上添加节点"}
// ~50 tokens
```

**完整工具定义** (chat):
```json
{
  "name": "add_node",
  "description": "在画板上添加节点",
  "parameters": {
    "type": "object",
    "properties": {
      "x": {"type": "number", "description": "X坐标"},
      "y": {"type": "number", "description": "Y坐标"},
      "width": {"type": "number", "default": 100},
      "height": {"type": "number", "default": 100},
      // ... 更多参数
    }
  }
}
// ~500 tokens
```

17 个工具 × (500 - 50) = **节省 7,650 tokens**

---

## 技术决策

### 1. 为什么不每次都重新筛选？

**当前设计**：用户输入时筛选 1 次，后续循环使用同一批工具

**原因**：
- 减少 API 调用次数
- AI 在第一次筛选时会"预判"所需的所有工具
- 工具连贯性更好，AI 可以在一批工具中灵活组合

**例子**：
```
用户："画流程图"
AI 筛选：["batch_add_nodes", "batch_add_lines", "update_node_position"]
         （预判可能需要调整位置）

执行过程：
1. 调用 batch_add_nodes
2. 调用 batch_add_lines
3. 如需调整，可直接用 update_node_position
```

### 2. 如何保证 AI 选对工具？

**系统提示词设计**：
- 明确要求分析完整对话历史
- 强调理解上下文（如"改成红色"需要结合前文）
- 优先使用批量工具
- 可以选择多个工具组合使用

**测试验证**：
```
输入："画流程图" → ["batch_add_nodes", "batch_add_lines"] ✅
输入："把第一个改红色" → ["update_node_color"] ✅
输入："删除最后一个" → ["delete_node"] ✅
```

### 3. 前端缓存策略

**AgentManager 设计**：
```typescript
this.currentToolIds = toolIds;  // 保存筛选结果

// 后续调用都用这个
this.sendStreamRequest(this.currentToolIds);
```

**好处**：
- 前端逻辑清晰
- 避免重复筛选
- 性能更好

---

## 测试验证

### 测试用例 1: 工具筛选

**请求**:
```json
POST /api/ai-stream/select-tools
{
  "agentId": "agent_xxx",
  "messages": [
    {"role": "user", "content": "给我画一个蛋炒饭流程图"}
  ]
}
```

**响应**:
```json
{
  "type": "complete",
  "content": "{\"selectedTools\": [\"batch_add_nodes\", \"batch_add_lines\"], \"reasoning\": \"流程图需要节点和连线\"}"
}
```

### 测试用例 2: 完整流程

**第 1 步**：select-tools → `["batch_add_nodes", "batch_add_lines"]`

**第 2 步**：chat (toolNames 传入) → 返回 toolCalls:
```json
{
  "type": "tool_calls",
  "toolCalls": [{
    "function": {
      "name": "batch_add_nodes",
      "arguments": "{\"nodes\": [{\"x\":100,\"y\":100,\"text\":\"准备食材\"}, ...]}"
    }
  }]
}
```

**验证点**：
- ✅ AI 正确筛选工具
- ✅ 工具参数生成完整
- ✅ Token 消耗符合预期（~2,050 tokens）
- ✅ 多轮工具调用流程正常

---

## 后续优化方向

### 1. 工具依赖关系
如果某些工具必须配合使用，可以在筛选时自动包含依赖工具。

### 2. 工具使用频率统计
记录工具选择频率，优化 AI 筛选准确性。

### 3. 缓存优化
对于相同的用户消息，可以缓存筛选结果（需考虑对话上下文）。

### 4. 支持工具分组
将 17 个工具按功能分组（节点操作、连线操作、样式修改等），进一步优化筛选逻辑。

---

## 技术总结

**核心价值**：
- **Token 成本优化**：单次对话节省 76% Token 消耗
- **智能工具选择**：AI 根据上下文精准筛选所需工具
- **架构可扩展**：轻松支持更多工具，不影响性能

**技术亮点**：
- 两阶段设计：轻量级筛选 + 完整定义执行
- 工具缓存机制：一次筛选，多次使用
- 系统提示词优化：引导 AI 准确选择工具
- 前后端协同设计：流程清晰，职责分明

**适用场景**：
- 工具数量较多（10+ 个）
- 需要频繁调用不同工具
- 对 Token 消耗敏感
- 需要支持复杂的多轮对话

---

**文档版本**: v1.0  
**更新日期**: 2026-02-22  
**作者**: Tenet Server Team
