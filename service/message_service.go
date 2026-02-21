package service

import (
	"encoding/json"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/google/uuid"

	"tenet-server/database"
	"tenet-server/model"
)

type MessageService struct{}

func NewMessageService() *MessageService {
	return &MessageService{}
}

// SaveUserMessages 保存用户消息和工具消息（assistant消息由AI返回时保存）
// 使用messageId去重，只保存新消息
func (s *MessageService) SaveUserMessages(conversationID string, messages []model.ChatMessage, agentID string) error {
	if conversationID == "" || len(messages) == 0 {
		hlog.Warn("conversationId或messages为空，跳过保存")
		return nil
	}

	// 1. 查询已存在的 messageId（用于去重）
	existingMessageIds := s.getExistingMessageIds(conversationID)

	// 2. 过滤出新消息（user和tool角色）
	var newMessages []model.ChatMessage
	for _, msg := range messages {
		if msg.Role != "user" && msg.Role != "tool" {
			continue
		}
		// 如果前端没传messageId，生成一个（基于内容hash或UUID）
		if msg.MessageID == "" {
			msg.MessageID = uuid.New().String()
		}
		// 去重：只保存不存在的消息
		if !existingMessageIds[msg.MessageID] {
			newMessages = append(newMessages, msg)
		}
	}

	if len(newMessages) == 0 {
		hlog.Info("没有需要保存的新消息（已全部存在）")
		return nil
	}

	// 3. 获取当前会话的最大 seq
	var maxSeq int
	database.DB.Model(&model.ChatMessageRecord{}).
		Where("conversation_id = ?", conversationID).
		Select("COALESCE(MAX(seq), 0)").
		Scan(&maxSeq)

	// 4. 为新消息分配seq并保存
	var toSave []*model.ChatMessageRecord
	for _, msg := range newMessages {

		maxSeq++
		chatMsg := &model.ChatMessageRecord{
			MessageID:      msg.MessageID, // 使用前端传的或生成的messageId
			ConversationID: conversationID,
			ModelID:        agentID,
			Role:           roleStringToInt(msg.Role),
			Content:        msg.Content,
			Seq:            maxSeq,
			CreateTime:     time.Now(),
		}

		// 如果有 toolCalls，序列化为 JSON
		if len(msg.ToolCalls) > 0 {
			toolCallsJSON, _ := json.Marshal(msg.ToolCalls)
			toolCallsStr := string(toolCallsJSON)
			chatMsg.ToolCalls = &toolCallsStr
		}

		// 如果是 tool 消息，记录 toolCallId
		if msg.Role == "tool" && msg.ToolCallID != "" {
			chatMsg.Content = msg.Content + " (toolCallId: " + msg.ToolCallID + ")"
		}

		toSave = append(toSave, chatMsg)
	}

	// 5. 批量插入
	err := database.DB.Create(&toSave).Error
	if err != nil {
		hlog.Errorf("保存消息失败: %v", err)
		return err
	}
	hlog.Infof("成功保存 %d 条新消息到会话 %s", len(toSave), conversationID)
	return nil
}

// SaveAssistantMessage 保存 AI 返回的消息
func (s *MessageService) SaveAssistantMessage(conversationID, agentID, content string, toolCalls []model.ToolCall) error {
	if conversationID == "" {
		return nil
	}

	// 获取当前会话的最大 seq
	var maxSeq int
	database.DB.Model(&model.ChatMessageRecord{}).
		Where("conversation_id = ?", conversationID).
		Select("COALESCE(MAX(seq), 0)").
		Scan(&maxSeq)

	chatMsg := &model.ChatMessageRecord{
		MessageID:      uuid.New().String(),
		ConversationID: conversationID,
		ModelID:        agentID,
		Role:           1, // assistant
		Content:        content,
		Seq:            maxSeq + 1,
		CreateTime:     time.Now(),
	}

	if len(toolCalls) > 0 {
		toolCallsJSON, _ := json.Marshal(toolCalls)
		toolCallsStr := string(toolCallsJSON)
		chatMsg.ToolCalls = &toolCallsStr
	}

	err := database.DB.Create(chatMsg).Error
	if err != nil {
		hlog.Errorf("保存assistant消息失败: %v", err)
		return err
	}

	hlog.Infof("成功保存assistant消息到会话 %s", conversationID)
	return nil
}

// GetMessagesByConversationID 查询对话的所有消息
func (s *MessageService) GetMessagesByConversationID(conversationID string) ([]model.ChatMessageRecord, error) {
	var messages []model.ChatMessageRecord
	err := database.DB.Where("conversation_id = ?", conversationID).
		Order("create_time ASC").
		Find(&messages).Error

	return messages, err
}

// roleStringToInt 将角色字符串转换为数字
func roleStringToInt(role string) int8 {
	switch role {
	case "user":
		return 0
	case "assistant":
		return 1
	case "tool":
		return 2
	default:
		return 0
	}
}

// getExistingMessageIds 查询会话中已存在的消息ID集合（用于去重）
func (s *MessageService) getExistingMessageIds(conversationID string) map[string]bool {
	var messageIds []string
	database.DB.Model(&model.ChatMessageRecord{}).
		Where("conversation_id = ?", conversationID).
		Pluck("message_id", &messageIds)

	// 转为map以便快速查找
	result := make(map[string]bool)
	for _, id := range messageIds {
		result[id] = true
	}
	return result
}

// RoleIntToString 将角色数字转换为字符串
func RoleIntToString(role int8) string {
	switch role {
	case 0:
		return "user"
	case 1:
		return "assistant"
	case 2:
		return "tool"
	default:
		return "user"
	}
}
