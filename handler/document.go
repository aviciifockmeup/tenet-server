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

// OpenDocument 打开文档
func OpenDocument(ctx context.Context, c *app.RequestContext) {
    documentId := c.Param("documentId")
    if documentId == "" {
        c.JSON(consts.StatusBadRequest, model.ErrorWithCode(model.ErrInvalidParam))
        return
    }

    documentService := service.NewDocumentService()
    docInfo, err := documentService.OpenDocument(documentId)
    if err != nil {
        c.JSON(consts.StatusInternalServerError, model.Error(model.CodeDatabaseError, err.Error()))
        return
    }

    c.JSON(consts.StatusOK, model.Success(docInfo))
}

// CreateDocument 创建文档
func CreateDocument(ctx context.Context, c *app.RequestContext) {
    var req struct {
        Title  string `json:"title"`
        UserId string `json:"userId"`
    }

    if err := c.BindJSON(&req); err != nil {
        c.JSON(consts.StatusBadRequest, model.ErrorWithCode(model.ErrInvalidParam))
        return
    }

    documentService := service.NewDocumentService()
    document, err := documentService.CreateDocument(req.Title, req.UserId)
    if err != nil {
        c.JSON(consts.StatusInternalServerError, model.Error(model.CodeDatabaseError, err.Error()))
        return
    }

    c.JSON(consts.StatusOK, model.Success(document))
}

// UpdateDocumentTitle 更新文档标题
func UpdateDocumentTitle(ctx context.Context, c *app.RequestContext) {
    var req struct {
        DocumentId string `json:"documentId"`
        Title      string `json:"title"`
    }

    if err := c.BindJSON(&req); err != nil {
        c.JSON(consts.StatusBadRequest, model.ErrorWithCode(model.ErrInvalidParam))
        return
    }

    documentService := service.NewDocumentService()
    if err := documentService.UpdateDocumentTitle(req.DocumentId, req.Title); err != nil {
        c.JSON(consts.StatusInternalServerError, model.Error(model.CodeDatabaseError, err.Error()))
        return
    }

    c.JSON(consts.StatusOK, model.Success(nil))
}

// DeleteDocument 删除文档
func DeleteDocument(ctx context.Context, c *app.RequestContext) {
    documentId := c.Param("documentId")
    if documentId == "" {
        c.JSON(consts.StatusBadRequest, model.ErrorWithCode(model.ErrInvalidParam))
        return
    }

    documentService := service.NewDocumentService()
    if err := documentService.DeleteDocument(documentId); err != nil {
        c.JSON(consts.StatusInternalServerError, model.Error(model.CodeDatabaseError, err.Error()))
        return
    }

    c.JSON(consts.StatusOK, model.Success(nil))
}