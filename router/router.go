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

    // 文档路由
    documentHandler := handler.NewDocumentHandler()
    h.GET("/api/document/list", documentHandler.GetDocumentList)
}