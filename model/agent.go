package model

import "time"

// Agent AI Agent配置
type Agent struct {
	ID            int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AgentID       string    `gorm:"column:agentId;uniqueIndex;not null" json:"agentId"`
	Name          string    `gorm:"column:name;not null" json:"name"`
	Provider      string    `gorm:"column:provider;not null" json:"provider"`
	APIKey        string    `gorm:"column:apiKey;not null" json:"apiKey"`
	Model         string    `gorm:"column:model;not null" json:"model"`
	Temperature   float64   `gorm:"column:temperature;not null;default:0.70" json:"temperature"`
	MaxTokens     int       `gorm:"column:maxTokens;not null;default:2000" json:"maxTokens"`
	SystemMessage string    `gorm:"column:systemMessage" json:"systemMessage"`
	FunctionCalls string    `gorm:"column:functionCalls;type:json" json:"functionCalls"` // JSON string
	CreateUserID  string    `gorm:"column:createUserId;not null" json:"createUserId"`
	CheckStatus   int       `gorm:"column:checkStatus;not null;default:0" json:"checkStatus"`
	CreateTime    time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"createTime"`
	ModifyTime    time.Time `gorm:"column:modify_time;not null;autoUpdateTime" json:"modifyTime"`
}

// TableName 指定表名
func (Agent) TableName() string {
	return "AgentTable"
}
