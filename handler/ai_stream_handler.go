package handler

import (
	"context"
	"tenet-server/model"
	"tenet-server/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"go.uber.org/zap"
)

// AIStreamHandler AI 流式对话处理器
type AIStreamHandler struct {
	aiService *service.AIService
}

// NewAIStreamHandler 创建 AI Stream Handler（依赖注入 AIService）
func NewAIStreamHandler(aiService *service.AIService) *AIStreamHandler {
	return &AIStreamHandler{
		aiService: aiService,
	}
}

// StreamChat 处理流式对话请求
func (h *AIStreamHandler) StreamChat(ctx context.Context, c *app.RequestContext) {
	var req model.AIStreamRequest

	// 解析请求体
	if err := c.BindJSON(&req); err != nil {
		c.JSON(consts.StatusBadRequest, model.Error(model.CodeInvalidParam, "Invalid request body"))
		return
	}

	// 如果提供了 agentId，从数据库查询配置
	if req.AgentID != "" {
		agentService := service.NewAgentService()
		agent, err := agentService.GetAgentByID(req.AgentID)
		if err != nil {
			c.JSON(consts.StatusBadRequest, model.Error(model.CodeInvalidParam, "Agent not found"))
			return
		}

		// 从数据库填充配置
		req.Config = &model.AIModelConfig{
			Provider:      agent.Provider,
			Model:         agent.Model,
			Temperature:   agent.Temperature,
			MaxTokens:     agent.MaxTokens,
			SystemMessage: agent.SystemMessage,
		}
	}

	// 如果提供了 toolNames，从数据库查询工具定义
	if len(req.ToolNames) > 0 {
		toolCallService := service.NewToolCallService()
		tools, err := toolCallService.GetToolsByNames(req.ToolNames)
		if err != nil {
			zap.L().Error("Failed to load tools", zap.Error(err))
			c.JSON(consts.StatusInternalServerError, model.Error(model.CodeDatabaseError, "Failed to load tools"))
			return
		}
		req.Tools = tools
		zap.L().Info("Loaded tools from database",
			zap.Int("count", len(tools)),
			zap.Strings("toolNames", req.ToolNames))
	}

	// ========== 对话持久化：保存用户消息 ==========
	if req.ConversationID != "" && req.AgentID != "" {
		conversationService := service.NewConversationService()
		messageService := service.NewMessageService()

		// 确保 Conversation 记录存在
		err := conversationService.EnsureConversationExists(
			req.ConversationID,
			"user_id", // TODO: 从上下文或JWT中获取真实的userId
			req.AgentID,
		)
		if err != nil {
			zap.L().Error("Failed to ensure conversation exists", zap.Error(err))
		}

		// 保存用户消息和工具消息
		err = messageService.SaveUserMessages(req.ConversationID, req.Messages, req.AgentID)
		if err != nil {
			zap.L().Error("Failed to save user messages", zap.Error(err))
		}
	}

	// 检查必要参数
	if req.Config == nil {
		c.JSON(consts.StatusBadRequest, model.Error(model.CodeInvalidParam, "Missing agent ID or model config"))
		return
	}
	if len(req.Messages) == 0 {
		c.JSON(consts.StatusBadRequest, model.Error(model.CodeInvalidParam, "Messages cannot be empty"))
		return
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 立即刷新响应头
	c.Flush()

	// 调用 AI Service 进行流式对话
	if err := h.aiService.StreamChat(ctx, &req, c); err != nil {
		zap.L().Error("StreamChat failed",
			zap.String("provider", req.Config.Provider),
			zap.Error(err))
	}
}
