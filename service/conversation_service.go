package service

import (
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/google/uuid"

	"tenet-server/database"
	"tenet-server/model"
)

type ConversationService struct{}

func NewConversationService() *ConversationService {
	return &ConversationService{}
}

// CreateConversation 创建新对话
func (s *ConversationService) CreateConversation(conversationID, userID, title string) (*model.Conversation, error) {
	if conversationID == "" {
		conversationID = uuid.New().String()
	}
	if title == "" {
		title = "新对话"
	}

	conversation := &model.Conversation{
		ConversationID: conversationID,
		UserID:         userID,
		Title:          title,
		CreateTime:     time.Now(),
		ModifyTime:     time.Now(),
	}

	err := database.DB.Create(conversation).Error
	if err != nil {
		hlog.Errorf("创建对话失败: %v", err)
		return nil, err
	}

	hlog.Infof("创建新对话成功: conversationId=%s", conversationID)
	return conversation, nil
}

// GetByConversationID 根据ID查询对话
func (s *ConversationService) GetByConversationID(conversationID string) (*model.Conversation, error) {
	var conversation model.Conversation
	err := database.DB.Where("conversation_id = ?", conversationID).First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

// EnsureConversationExists 确保对话存在，不存在则创建
func (s *ConversationService) EnsureConversationExists(conversationID, userID, agentID string) error {
	if conversationID == "" {
		return nil
	}

	_, err := s.GetByConversationID(conversationID)
	if err != nil {
		// 不存在，创建新对话
		hlog.Infof("对话不存在，创建新对话: conversationId=%s, userId=%s", conversationID, userID)
		_, err = s.CreateConversation(conversationID, userID, "新对话")
		return err
	}

	hlog.Infof("对话已存在: conversationId=%s", conversationID)
	return nil
}

// ListByUserID 查询用户的所有对话
func (s *ConversationService) ListByUserID(userID string, limit int) ([]model.Conversation, error) {
	var conversations []model.Conversation
	query := database.DB.Where("userid = ?", userID).
		Order("modify_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&conversations).Error
	return conversations, err
}
