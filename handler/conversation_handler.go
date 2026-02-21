package handler

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"

	"tenet-server/model"
	"tenet-server/service"
)

type ConversationHandler struct {
	conversationService *service.ConversationService
	messageService      *service.MessageService
}

func NewConversationHandler(cs *service.ConversationService, ms *service.MessageService) *ConversationHandler {
	return &ConversationHandler{
		conversationService: cs,
		messageService:      ms,
	}
}

// GetMessages 获取对话的所有消息
func (h *ConversationHandler) GetMessages(ctx context.Context, c *app.RequestContext) {
	conversationID := c.Param("conversationId")

	messages, err := h.messageService.GetMessagesByConversationID(conversationID)
	if err != nil {
		c.JSON(500, utils.H{"error": err.Error()})
		return
	}

	// 转换为前端格式
	var result []map[string]interface{}
	for _, msg := range messages {
		item := map[string]interface{}{
			"messageId":      msg.MessageID,
			"conversationId": msg.ConversationID,
			"role":           service.RoleIntToString(msg.Role),
			"content":        msg.Content,
			"createTime":     msg.CreateTime,
		}

		// 解析 toolCalls
		if msg.ToolCalls != nil && *msg.ToolCalls != "" {
			var toolCalls []model.ToolCall
			if err := json.Unmarshal([]byte(*msg.ToolCalls), &toolCalls); err == nil {
				item["toolCalls"] = toolCalls
			}
		}

		result = append(result, item)
	}

	c.JSON(200, utils.H{
		"conversationId": conversationID,
		"messages":       result,
		"total":          len(result),
	})
}

// CreateConversation 创建新对话
func (h *ConversationHandler) CreateConversation(ctx context.Context, c *app.RequestContext) {
	var req struct {
		UserID string `json:"userId"`
		Title  string `json:"title"`
	}
	if err := c.Bind(&req); err != nil {
		c.JSON(400, utils.H{"error": "参数错误"})
		return
	}

	conversation, err := h.conversationService.CreateConversation("", req.UserID, req.Title)
	if err != nil {
		c.JSON(500, utils.H{"error": err.Error()})
		return
	}

	c.JSON(200, utils.H{"conversation": conversation})
}

// ListConversations 获取用户的对话列表
func (h *ConversationHandler) ListConversations(ctx context.Context, c *app.RequestContext) {
	userID := c.Query("userId")
	if userID == "" {
		c.JSON(400, utils.H{"error": "缺少userId参数"})
		return
	}

	conversations, err := h.conversationService.ListByUserID(userID, 50)
	if err != nil {
		c.JSON(500, utils.H{"error": err.Error()})
		return
	}

	c.JSON(200, utils.H{
		"conversations": conversations,
		"total":         len(conversations),
	})
}
