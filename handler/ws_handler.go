package handler

import (
	"context"
	"log"
	"tenet-server/model"
	"tenet-server/ws"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"
)

var upgrader = websocket.HertzUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(c *app.RequestContext) bool {
		return true // 生产环境需要验证 origin
	},
}

// HandleWebSocket WebSocket 连接处理
func HandleWebSocket(hub *ws.Hub) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 获取 query 参数
		documentId := c.Query("documentId")
		userId := c.Query("userId")

		if documentId == "" || userId == "" {
			c.JSON(400, model.Error(400, "缺少 documentId 或 userId 参数"))
			return
		}

		// 升级为 WebSocket
		err := upgrader.Upgrade(c, func(conn *websocket.Conn) {
			// 创建客户端
			client := ws.NewClient(hub, conn, documentId, userId)

			// 注册到 Hub
			hub.Register <- client

			// 启动读写协程
			go client.WritePump()
			client.ReadPump() // 阻塞在这里，直到连接断开
		})

		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}
	}
}
