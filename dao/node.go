package dao

import (
	"tenet-server/database"
	"tenet-server/model"
)

type NodeDAO struct{}

func NewNodeDAO() *NodeDAO {
	return &NodeDAO{}
}

// GetByDocumentId 根据文档ID查询所有节点
func (d *NodeDAO) GetByDocumentId(documentId string) ([]model.Node, error) {
	var nodes []model.Node
	err := database.DB.Debug().Where("documentid = ?", documentId).Find(&nodes).Error
	return nodes, err
}

func (d *NodeDAO) Create(node *model.Node) error {
	return database.DB.Create(node).Error
}

func (d *NodeDAO) Update(node *model.Node) error {
	return database.DB.Save(node).Error
}

func (d *NodeDAO) Delete(nodeId string) error {
	return database.DB.Where("nodeId = ?", nodeId).Delete(&model.Node{}).Error
}
