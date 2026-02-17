package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/hertz-contrib/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB
)

// Client WebSocket 客户端
type Client struct {
	id         string
	hub        *Hub
	conn       *websocket.Conn
	send       chan []byte
	documentId string          // 当前所在房间
	userId     string          // 用户ID
	msgHandler *MessageHandler // 消息处理器
}

// NewClient 创建新客户端
func NewClient(hub *Hub, conn *websocket.Conn, documentId string, userId string) *Client {
	return &Client{
		id:         generateClientId(),
		hub:        hub,
		conn:       conn,
		send:       make(chan []byte, 256),
		documentId: documentId,
		userId:     userId,
		msgHandler: NewMessageHandler(),
	}
}

// ReadPump 读取消息
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}

		// 解析消息
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("json unmarshal error: %v", err)
			continue
		}

		// 处理消息（保存到数据库）
		var handleErr error
		if !msg.Meta.Streaming {
			// 非流式消息需要落库
			handleErr = c.msgHandler.HandleMessage(&msg)
		}

		// 广播给房间内其他用户
		c.hub.Broadcast <- &BroadcastMessage{
			Message:   &msg,
			RoomId:    c.documentId,
			ExcludeId: c.id,
		}

		// 发送确认消息
		ack := Ack{
			OpId:    msg.Meta.OpId,
			Success: handleErr == nil,
		}
		if handleErr != nil {
			ack.Error = handleErr.Error()
		}
		ackBytes, _ := json.Marshal(ack)
		c.send <- ackBytes
	}
}

// WritePump 发送消息给浏览器
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送缓冲区中的消息
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// generateClientId 生成客户端ID
func generateClientId() string {
	return time.Now().Format("20060102150405") + "_" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
