package handler

import (
	"context"
	"tenet-server/model"
	"tenet-server/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// GetDocumentList 获取文档列表
func GetDocumentList(ctx context.Context, c *app.RequestContext) {
	documentService := service.NewDocumentService()
	documents, err := documentService.GetDocumentList()
	if err != nil {
		c.JSON(consts.StatusInternalServerError, model.Error(model.CodeDatabaseError, err.Error()))
		return
	}

	c.JSON(consts.StatusOK, model.Success(documents))
}
