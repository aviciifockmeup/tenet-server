package handler

import (
	"context"
	"tenet-server/model"
	"tenet-server/ws"

	"github.com/cloudwego/hertz/pkg/app"
)

// GetRoomUsers 获取房间内的用户列表
func GetRoomUsers(hub *ws.Hub) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		documentId := c.Param("documentId")
		if documentId == "" {
			c.JSON(400, model.Error(400, "缺少 documentId 参数"))
			return
		}

		users := hub.GetRoomUsers(documentId)
		c.JSON(200, model.Success(map[string]interface{}{
			"documentId": documentId,
			"userCount":  len(users),
			"users":      users,
		}))
	}
}
