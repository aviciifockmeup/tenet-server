package ws

import (
	"log"
	"tenet-server/service"
)

// MessageHandler 消息处理器
type MessageHandler struct {
	nodeService *service.NodeService
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		nodeService: service.NewNodeService(),
	}
}

// HandleMessage 处理消息
func (h *MessageHandler) HandleMessage(msg *Message) error {
	switch msg.Meta.Type {
	case "node_create":
		return h.handleNodeCreate(msg)
	case "node_update":
		return h.handleNodeUpdate(msg)
	case "node_delete":
		return h.handleNodeDelete(msg)
	default:
		log.Printf("Unknown message type: %s", msg.Meta.Type)
		return nil
	}
}

// handleNodeCreate 处理节点创建
func (h *MessageHandler) handleNodeCreate(msg *Message) error {
	if msg.Data.Node == nil {
		log.Printf("node_create: node is nil")
		return nil
	}

	node := msg.Data.Node
	// 设置文档ID
	node.DocumentId = msg.Meta.DocumentId

	// 调用 service 层保存节点
	if err := h.nodeService.CreateNode(node); err != nil {
		log.Printf("Failed to create node: %v", err)
		return err
	}

	log.Printf("Node created: %s in document: %s", node.NodeId, node.DocumentId)
	return nil
}

// handleNodeUpdate 处理节点更新
func (h *MessageHandler) handleNodeUpdate(msg *Message) error {
	if msg.Data.Node == nil {
		log.Printf("node_update: node is nil")
		return nil
	}

	node := msg.Data.Node
	// 调用 service 层更新节点
	if err := h.nodeService.UpdateNode(node); err != nil {
		log.Printf("Failed to update node: %v", err)
		return err
	}

	log.Printf("Node updated: %s", node.NodeId)
	return nil
}

// handleNodeDelete 处理节点删除
func (h *MessageHandler) handleNodeDelete(msg *Message) error {
	if msg.Data.Node == nil || msg.Data.Node.NodeId == "" {
		log.Printf("node_delete: nodeId is empty")
		return nil
	}

	nodeId := msg.Data.Node.NodeId
	// 调用 service 层删除节点
	if err := h.nodeService.DeleteNode(nodeId); err != nil {
		log.Printf("Failed to delete node: %v", err)
		return err
	}

	log.Printf("Node deleted: %s", nodeId)
	return nil
}
