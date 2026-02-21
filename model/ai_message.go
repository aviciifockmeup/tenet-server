package model

// AIStreamRequest AI流式请求
type AIStreamRequest struct {
    AgentID   string          `json:"agentId,omitempty"`   // Agent ID（可选）
    Config    *AIModelConfig  `json:"config"`              // 模型配置
    Messages  []ChatMessage   `json:"messages"`            // 对话历史
    ToolNames []string        `json:"toolNames,omitempty"` // 工具名称列表（前端传入）
    Tools     []Tool          `json:"tools,omitempty"`     // 工具定义（后端填充）
}

// AIModelConfig 模型配置
type AIModelConfig struct {
    Provider      string  `json:"provider"`      // deepseek/openai/claude
    Model         string  `json:"model"`         
    Temperature   float64 `json:"temperature"`   
    MaxTokens     int     `json:"maxTokens"`     
    SystemMessage string  `json:"systemMessage"` 
}

// ChatMessage 对话消息
type ChatMessage struct {
    Role       string      `json:"role"`                 // user/assistant/system/tool
    Content    string      `json:"content"`              
    ToolCalls  []ToolCall  `json:"toolCalls,omitempty"`  // assistant发出的工具调用
    ToolCallID string      `json:"toolCallId,omitempty"` // tool返回时对应的ID
}

// ToolCall 工具调用
type ToolCall struct {
    ID       string `json:"id"`       // call_xxx
    Type     string `json:"type"`     // function
    Function struct {
        Name      string `json:"name"`      
        Arguments string `json:"arguments"` // JSON字符串
    } `json:"function"`
}

// Tool 工具定义
type Tool struct {
    Type     string `json:"type"` // function
    Function struct {
        Name        string      `json:"name"`        
        Description string      `json:"description"` 
        Parameters  interface{} `json:"parameters"`  // JSON Schema
    } `json:"function"`
}


// SSEChunk SSE响应
type SSEChunk struct {
    Type      string     `json:"type"`                // chunk/complete/error
    Content   string     `json:"content,omitempty"`   
    ToolCalls []ToolCall `json:"toolCalls,omitempty"` 
    Error     string     `json:"error,omitempty"`     
}