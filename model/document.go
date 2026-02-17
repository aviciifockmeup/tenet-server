package model

import "time"

// Document 文档模型
type Document struct {
	ID            int       `gorm:"column:id;primaryKey" json:"id"`
	DocumentId    string    `gorm:"column:document_id" json:"documentId"`
	DocumentTitle string    `gorm:"column:document_title" json:"documentTitle"`
	CoverUrl      string    `gorm:"column:cover_url" json:"coverUrl"`
	UserId        string    `gorm:"column:userid" json:"userId"`
	CategoryId    string    `gorm:"column:category_id" json:"categoryId"`
	Status        int       `gorm:"column:status" json:"status"`
	CreateTime    time.Time `gorm:"column:create_time" json:"createTime"`
	ModifyTime    time.Time `gorm:"column:modify_time" json:"modifyTime"`
}

// TableName 指定表名
func (Document) TableName() string {
	return "DocumentInfo"
}

// DocumentInfo 打开文档时返回的结构（文档信息+所有节点）
type DocumentInfo struct {
	DocumentId string `json:"documentId"`
	Title      string `json:"title"`
	Nodes      []Node `json:"nodes"`
}
