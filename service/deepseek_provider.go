package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"tenet-server/model"
	"tenet-server/utils"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// DeepSeekProvider DeepSeek实现
type DeepSeekProvider struct {
	apiKey string
	client *http.Client
}

// NewDeepSeekProvider 创建DeepSeek provider
func NewDeepSeekProvider(apiKey string) *DeepSeekProvider {
	return &DeepSeekProvider{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// StreamChat 实现AIProvider接口
func (p *DeepSeekProvider) StreamChat(ctx context.Context, req *model.AIStreamRequest, c *app.RequestContext) error {
	// 1. 构建DeepSeek格式的请求体
	body := p.buildRequestBody(req)

	// 2. 发送HTTP请求
	httpReq, _ := http.NewRequestWithContext(ctx, "POST",
		"https://api.deepseek.com/v1/chat/completions",
		bytes.NewReader(body))

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DeepSeek API错误 (%d): %s", resp.StatusCode, bodyBytes)
	}

	// 3. 处理流式响应
	return p.handleStream(resp.Body, c, req)
}

// buildRequestBody 构建请求体
func (p *DeepSeekProvider) buildRequestBody(req *model.AIStreamRequest) []byte {
	// 构建messages（system + 历史）
	messages := []map[string]interface{}{
		{"role": "system", "content": req.Config.SystemMessage},
	}

	for _, msg := range req.Messages {
		m := map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}

		// 处理tool_calls
		if len(msg.ToolCalls) > 0 {
			m["tool_calls"] = msg.ToolCalls
		}

		// 处理tool_call_id
		if msg.ToolCallID != "" {
			m["tool_call_id"] = msg.ToolCallID
		}

		messages = append(messages, m)
	}

	// 设置默认model
	model := req.Config.Model
	if model == "" {
		model = "deepseek-chat" // DeepSeek默认模型
	}

	// 设置默认temperature
	temperature := req.Config.Temperature
	if temperature == 0 {
		temperature = 0.7
	}

	// 设置默认max_tokens
	maxTokens := req.Config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	// 构建完整请求
	requestBody := map[string]interface{}{
		"model":       model,
		"messages":    messages,
		"temperature": temperature,
		"max_tokens":  maxTokens,
		"stream":      true,
	}

	// 添加tools（如果有）
	if len(req.Tools) > 0 {
		requestBody["tools"] = req.Tools
	}

	body, _ := json.Marshal(requestBody)
	hlog.Infof("DeepSeek请求: %s", string(body))
	return body
}

// handleStream 处理流式响应
func (p *DeepSeekProvider) handleStream(body io.Reader, c *app.RequestContext, req *model.AIStreamRequest) error {
	// 设置SSE响应头
	c.SetContentType("text/event-stream")
	c.Response.Header.Set("Cache-Control", "no-cache")
	c.Response.Header.Set("Connection", "keep-alive")

	reader := bufio.NewReader(body)
	var fullContent strings.Builder
	toolCallsMap := make(map[int]*model.ToolCall)

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		// 解析chunk
		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		// 提取choices
		choices, ok := chunk["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			continue
		}

		choice := choices[0].(map[string]interface{})
		delta, ok := choice["delta"].(map[string]interface{})
		if !ok {
			continue
		}

		// 处理文本内容
		if content, ok := delta["content"].(string); ok && content != "" {
			fullContent.WriteString(content)
			utils.SendSSEChunk(c, model.SSEChunk{Type: "chunk", Content: content})
		}

		// 处理tool_calls（累积）
		if toolCalls, ok := delta["tool_calls"].([]interface{}); ok {
			for _, tc := range toolCalls {
				tcMap := tc.(map[string]interface{})
				index := int(tcMap["index"].(float64))

				if toolCallsMap[index] == nil {
					toolCallsMap[index] = &model.ToolCall{
						ID:   getStr(tcMap, "id"),
						Type: getStr(tcMap, "type"),
					}
				}

				if fn, ok := tcMap["function"].(map[string]interface{}); ok {
					if name := getStr(fn, "name"); name != "" {
						toolCallsMap[index].Function.Name = name
					}
					if args := getStr(fn, "arguments"); args != "" {
						toolCallsMap[index].Function.Arguments += args
					}
				}
			}
		}
	}

	// 发送complete（包含完整的toolCalls）
	var toolCalls []model.ToolCall
	for i := 0; i < len(toolCallsMap); i++ {
		if tc := toolCallsMap[i]; tc != nil {
			toolCalls = append(toolCalls, *tc)
		}
	}

	utils.SendSSEChunk(c, model.SSEChunk{
		Type:      "complete",
		Content:   fullContent.String(),
		ToolCalls: toolCalls,
	})

	// ========== 保存 Assistant 消息到数据库 ==========
	if req.ConversationID != "" && req.AgentID != "" {
		messageService := NewMessageService()
		err := messageService.SaveAssistantMessage(
			req.ConversationID,
			req.AgentID,
			fullContent.String(),
			toolCalls,
		)
		if err != nil {
			hlog.Errorf("保存Assistant消息失败: %v", err)
		}
	}

	return nil
}

// 工具函数
func getStr(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func sendSSEError(c *app.RequestContext, msg string) error {
	c.SetContentType("text/event-stream")
	utils.SendSSEChunk(c, model.SSEChunk{Type: "error", Error: msg})
	return nil
}
