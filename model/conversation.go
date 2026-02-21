package model

import "time"

// Conversation 对话会话
type Conversation struct {
	ID             int       `gorm:"column:id;primaryKey;autoIncrement"`
	ConversationID string    `gorm:"column:conversation_id;type:varchar(255);uniqueIndex"`
	UserID         string    `gorm:"column:userid;type:varchar(255)"`
	Title          string    `gorm:"column:title;type:varchar(500)"`
	CreateTime     time.Time `gorm:"column:create_time;autoCreateTime"`
	ModifyTime     time.Time `gorm:"column:modify_time;autoUpdateTime"`
}

func (Conversation) TableName() string {
	return "Conversation"
}

// ChatMessageRecord 聊天消息记录（数据库模型）
type ChatMessageRecord struct {
	ID             int       `gorm:"column:id;primaryKey;autoIncrement"`
	MessageID      string    `gorm:"column:message_id;type:varchar(255);uniqueIndex"`
	ConversationID string    `gorm:"column:conversation_id;type:varchar(255);index"`
	ModelID        string    `gorm:"column:model_id;type:varchar(255)"` // agentId
	Role           int8      `gorm:"column:role;type:tinyint"`          // 0=user, 1=assistant, 2=tool
	Content        string    `gorm:"column:content;type:text"`
	ToolCalls      *string   `gorm:"column:tool_calls;type:text"`                // JSON string, 可为NULL
	Seq            int       `gorm:"column:seq;type:int;uniqueIndex:idx_conv_seq"` // 会话内序号
	CreateTime     time.Time `gorm:"column:create_time;autoCreateTime"`
}

func (ChatMessageRecord) TableName() string {
	return "ChatMessage"
}
