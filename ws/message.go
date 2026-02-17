package ws

import "tenet-server/model"

// Message WebSocket 消息结构
type Message struct {
	Meta MessageMeta `json:"meta"`
	Data MessageData `json:"data"`
}

// MessageMeta 消息元数据
type MessageMeta struct {
	UserId     string `json:"userId"`     // 用户ID
	DocumentId string `json:"documentId"` // 文档ID
	OpId       string `json:"opId"`       // 操作ID
	Type       string `json:"type"`       // 消息类型: node_create, node_update, node_delete
	Streaming  bool   `json:"streaming"`  // 是否流式消息（不落库）
}

// MessageData 消息数据
type MessageData struct {
	Node *model.Node `json:"node,omitempty"` // 节点数据
}

// BroadcastMessage 广播消息
type BroadcastMessage struct {
	Message   *Message
	RoomId    string // documentId
	ExcludeId string // 排除的客户端ID（发送者自己）
}

// Ack 消息确认
type Ack struct {
	OpId    string `json:"opId"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// UserJoinLeaveMessage 用户加入/离开消息
type UserJoinLeaveMessage struct {
	Type       string `json:"type"` // "user_join" 或 "user_leave"
	UserId     string `json:"userId"`
	DocumentId string `json:"documentId"`
	Timestamp  int64  `json:"timestamp"`
}
