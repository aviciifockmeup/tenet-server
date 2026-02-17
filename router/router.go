package router

import (
	"context"
	"tenet-server/handler"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

// Setup 注册所有路由
func Setup(h *server.Hertz, cfg interface{}) {
	// 健康检查
	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(200, utils.H{
			"status":  "ok",
			"service": "tenet-server",
		})
	})

	// API 路由组
	api := h.Group("/api")
	{
		// 文档相关
		doc := api.Group("/document")
		{
			doc.GET("/list", handler.GetDocumentList)
			doc.GET("/open/:documentId", handler.OpenDocument)
			doc.POST("/create", handler.CreateDocument)
			doc.POST("/updateDocTitle", handler.UpdateDocumentTitle)
			doc.DELETE("/delete/:documentId", handler.DeleteDocument)
		}

		// 节点相关
		node := api.Group("/node")
		{
			node.GET("/list/:documentId", handler.GetNodesByDocumentId)
			node.POST("/create", handler.CreateNode)
			node.PUT("/update", handler.UpdateNode)
			node.DELETE("/delete/:nodeId", handler.DeleteNode)
		}
	}
}
