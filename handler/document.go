package handler

import (
    "context"
    "tenet-server/service"

    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/common/utils"
)

type DocumentHandler struct {
    service *service.DocumentService
}

func NewDocumentHandler() *DocumentHandler {
    return &DocumentHandler{
        service: service.NewDocumentService(),
    }
}

// GetDocumentList 获取文档列表
func (h *DocumentHandler) GetDocumentList(ctx context.Context, c *app.RequestContext) {
    documents, err := h.service.GetDocumentList()
    if err != nil {
        c.JSON(500, utils.H{
            "code":    500,
            "message": "查询失败",
            "error":   err.Error(),
        })
        return
    }

    c.JSON(200, utils.H{
        "code":    200,
        "message": "success",
        "data":    documents,
    })
}