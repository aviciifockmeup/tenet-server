package model

import "time"

// ToolCallDef Tool Call工具定义
type ToolCallDef struct {
	ID          int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"column:name;uniqueIndex;not null" json:"name"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	Parameters  string    `gorm:"column:parameters;type:json" json:"parameters"` // JSON Schema string
	Category    string    `gorm:"column:category" json:"category"`
	CheckStatus int       `gorm:"column:checkStatus;not null;default:0" json:"checkStatus"`
	CreateTime  time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"createTime"`
	ModifyTime  time.Time `gorm:"column:modify_time;not null;autoUpdateTime" json:"modifyTime"`
}

// TableName 指定表名
func (ToolCallDef) TableName() string {
	return "ToolCallTable"
}
