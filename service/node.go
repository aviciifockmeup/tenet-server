package service

import (
    "tenet-server/dao"
    "tenet-server/model"
    "time"
)

type NodeService struct {
	dao *dao.NodeDAO
}

func NewNodeService() *NodeService {
	return &NodeService{
		dao: dao.NewNodeDAO(),
	}
}

// GetNodesByDocumentId 获取文档的所有节点
func (s *NodeService) GetNodesByDocumentId(documentId string) ([]model.Node, error) {
    return s.dao.GetByDocumentId(documentId)
}

func (s *NodeService) CreateNode(node *model.Node) error {
    node.CreateTime = time.Now()
    node.ModifyTime = time.Now()
    return s.dao.Create(node)
}

func (s *NodeService) UpdateNode(node *model.Node) error {
    node.ModifyTime = time.Now()
    return s.dao.Update(node)
}

func (s *NodeService) DeleteNode(nodeId string) error {
    return s.dao.Delete(nodeId)
}