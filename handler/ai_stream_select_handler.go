package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"tenet-server/model"
	"tenet-server/service"
	"tenet-server/utils"
)

type AIStreamSelectHandler struct {
	agentService    *service.AgentService
	toolCallService *service.ToolCallService
	aiService       *service.AIService
}

func NewAIStreamSelectHandler(agentService *service.AgentService, toolCallService *service.ToolCallService, aiService *service.AIService) *AIStreamSelectHandler {
	return &AIStreamSelectHandler{
		agentService:    agentService,
		toolCallService: toolCallService,
		aiService:       aiService,
	}
}

// SelectTools 工具筛选接口
func (h *AIStreamSelectHandler) SelectTools(ctx context.Context, c *app.RequestContext) {
	var req model.AIStreamRequest
	if err := c.Bind(&req); err != nil {
		hlog.Errorf("参数绑定失败: %v", err)
		utils.SendSSEChunk(c, model.SSEChunk{
			Type:  "error",
			Error: "参数解析失败",
		})
		return
	}

	// 验证消息列表
	if len(req.Messages) == 0 {
		hlog.Error("消息列表为空")
		utils.SendSSEChunk(c, model.SSEChunk{
			Type:  "error",
			Error: "缺少对话消息",
		})
		return
	}

	// 根据 agentId 查询 Agent 配置
	if req.AgentID != "" {
		agent, err := h.agentService.GetAgentByID(req.AgentID)
		if err != nil {
			hlog.Errorf("查询 Agent 失败: %v", err)
			utils.SendSSEChunk(c, model.SSEChunk{
				Type:  "error",
				Error: "未找到指定的 Agent 配置",
			})
			return
		}

		// 获取所有工具的轻量级信息
		allTools, err := h.toolCallService.GetAllToolsBasicInfo()
		if err != nil {
			hlog.Errorf("查询工具列表失败: %v", err)
			utils.SendSSEChunk(c, model.SSEChunk{
				Type:  "error",
				Error: "查询工具列表失败",
			})
			return
		}

		// 构建工具筛选的系统消息
		systemMessage := buildToolSelectionSystemMessage(allTools)

		// 填充 Agent 配置
		req.Config = &model.AIModelConfig{
			Provider:      agent.Provider,
			Model:         agent.Model,
			Temperature:   agent.Temperature,
			MaxTokens:     agent.MaxTokens,
			SystemMessage: systemMessage,
		}

		hlog.Infof("工具筛选使用 %d 条对话历史消息", len(req.Messages))
	}

	// 验证配置
	if req.Config == nil {
		utils.SendSSEChunk(c, model.SSEChunk{
			Type:  "error",
			Error: "缺少 Agent ID 或模型配置",
		})
		return
	}

	hlog.Infof("用户发起工具筛选请求, agentId: %s, provider: %s, model: %s",
		req.AgentID, req.Config.Provider, req.Config.Model)

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 调用 AIService 进行流式对话（AI 会返回筛选结果）
	if err := h.aiService.StreamChat(ctx, &req, c); err != nil {
		hlog.Errorf("AI 流式调用失败: %v", err)
		utils.SendSSEChunk(c, model.SSEChunk{
			Type:  "error",
			Error: fmt.Sprintf("AI 调用失败: %v", err),
		})
	}
}

// buildToolSelectionSystemMessage 构建工具筛选的系统提示词
func buildToolSelectionSystemMessage(allTools []map[string]string) string {
	// 将工具列表转为 JSON
	toolsJSON, _ := json.Marshal(allTools)

	return fmt.Sprintf(`你是一个智能工具筛选助手。你的任务是根据多轮对话历史和可用工具列表，选择最合适的工具提供给绘图大师进行绘图，使得图案尽量美观，好看，专业。

**任务要求：**
1. 仔细分析完整的对话历史，理解用户的真实意图和上下文
2. 从可用工具列表中选择能够帮助完成当前任务的工具
3. 只选择真正需要的工具，避免选择无关的工具
4. 能使用批量工具，优先使用批量工具
5. 可以适当使用svg工具生成svg图标
6. 如果不需要任何工具，返回空数组
7. 考虑对话上下文，例如用户说"改成红色"时，需要结合之前的对话理解是要修改什么

**可用工具列表：**
%s

**返回格式（严格按照JSON格式返回）：**
{
  "selectedTools": ["tool_name_1", "tool_name_2"],
  "reasoning": "简要说明选择这些工具的理由"
}

请直接返回 JSON，不要包含任何其他文本。`, string(toolsJSON))
}