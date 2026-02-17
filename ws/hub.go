package ws

import (
	"encoding/json"
	"log"
)

// Hub 管理所有 WebSocket 连接
type Hub struct {
	// 房间管理（documentId -> clients）
	rooms map[string]map[*Client]bool

	// 注册请求
	Register chan *Client

	// 注销请求
	Unregister chan *Client

	// 广播消息
	Broadcast chan *BroadcastMessage
}

// NewHub 创建 Hub
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		Register:   make(chan *Client, 256),           // 缓冲256个连接请求
		Unregister: make(chan *Client, 256),           // 缓冲256个断开请求
		Broadcast:  make(chan *BroadcastMessage, 256), // 缓冲256条消息
	}
}

// Run 启动 Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// 加入房间
			if _, ok := h.rooms[client.documentId]; !ok {
				h.rooms[client.documentId] = make(map[*Client]bool)
			}
			h.rooms[client.documentId][client] = true
			log.Printf("Client %s (user: %s) joined room %s", client.id, client.userId, client.documentId)

			// 广播用户加入消息给房间其他人
			joinMsg := &Message{
				Meta: MessageMeta{
					UserId:     client.userId,
					DocumentId: client.documentId,
					Type:       "user_join",
				},
			}
			h.Broadcast <- &BroadcastMessage{
				Message:   joinMsg,
				RoomId:    client.documentId,
				ExcludeId: client.id, // 不发给刚加入的用户自己
			}

		case client := <-h.Unregister:
			// 离开房间
			if clients, ok := h.rooms[client.documentId]; ok {
				if _, ok := clients[client]; ok {
					// 先广播用户离开消息（在删除之前）
					leaveMsg := &Message{
						Meta: MessageMeta{
							UserId:     client.userId,
							DocumentId: client.documentId,
							Type:       "user_leave",
						},
					}
					// 广播给房间其他人
					if len(clients) > 1 {
						messageBytes, _ := json.Marshal(leaveMsg)
						for c := range clients {
							if c.id != client.id {
								select {
								case c.send <- messageBytes:
								default:
								}
							}
						}
					}

					delete(clients, client)
					close(client.send)
					log.Printf("Client %s (user: %s) left room %s", client.id, client.userId, client.documentId)

					// 如果房间为空，删除房间
					if len(clients) == 0 {
						delete(h.rooms, client.documentId)
						log.Printf("Room %s is empty and removed", client.documentId)
					}
				}
			}

		case broadcast := <-h.Broadcast:
			// 广播给房间内其他客户端
			if clients, ok := h.rooms[broadcast.RoomId]; ok {
				messageBytes, _ := json.Marshal(broadcast.Message)
				for client := range clients {
					if client.id != broadcast.ExcludeId {
						select {
						case client.send <- messageBytes:
							// 非阻塞容错
						default:
							close(client.send)
							delete(clients, client)
						}
					}
				}
			}
		}
	}
}

// GetRoomClients 获取房间内的客户端数量
func (h *Hub) GetRoomClients(documentId string) int {
	if clients, ok := h.rooms[documentId]; ok {
		return len(clients)
	}
	return 0
}

// GetRoomUsers 获取房间内的用户列表
func (h *Hub) GetRoomUsers(documentId string) []map[string]string {
	var users []map[string]string
	if clients, ok := h.rooms[documentId]; ok {
		for client := range clients {
			users = append(users, map[string]string{
				"userId":   client.userId,
				"clientId": client.id,
			})
		}
	}
	return users
}
