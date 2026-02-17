package model

import "time"

type Node struct {
	ID         int       `gorm:"column:id;primaryKey" json:"id"`
	DocumentId string    `gorm:"column:documentid" json:"documentId"`
	NodeId     string    `gorm:"column:nodeId" json:"nodeId"`
	ParentId   string    `gorm:"column:parentId" json:"parentId"`
	Type       int       `gorm:"column:type" json:"type"`
	ZIndex     float32   `gorm:"column:zIndex" json:"zIndex"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
	ModifyTime time.Time `gorm:"column:modify_time" json:"modifyTime"`
	CapInfo    string    `gorm:"column:capInfo" json:"capInfo"`
}

func (Node) TableName() string {
	return "NodeInfo"
}
