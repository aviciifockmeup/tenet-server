package handler

import (
    "context"
    "tenet-server/model"
    "tenet-server/service"

    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/protocol/consts"
)

// GetNodesByDocumentId 获取文档下所有节点
func GetNodesByDocumentId(ctx context.Context, c *app.RequestContext) {
    documentId := c.Param("documentId")
    if documentId == "" {
        c.JSON(consts.StatusBadRequest, model.ErrorWithCode(model.ErrInvalidParam))
        return
    }

    nodeService := service.NewNodeService()
    nodes, err := nodeService.GetNodesByDocumentId(documentId)
    if err != nil {
        c.JSON(consts.StatusInternalServerError, model.Error(model.CodeDatabaseError, err.Error()))
        return
    }

    c.JSON(consts.StatusOK, model.Success(nodes))
}

func CreateNode(ctx context.Context, c *app.RequestContext) {
    var node model.Node
    if err := c.BindJSON(&node); err != nil {
        c.JSON(consts.StatusBadRequest, model.ErrorWithCode(model.ErrInvalidParam))
        return
    }

    nodeService := service.NewNodeService()
    if err := nodeService.CreateNode(&node); err != nil {
        c.JSON(consts.StatusInternalServerError, model.Error(model.CodeDatabaseError, err.Error()))
        return
    }

    c.JSON(consts.StatusOK, model.Success(node))
}

func UpdateNode(ctx context.Context, c *app.RequestContext) {
    var node model.Node
    if err := c.BindJSON(&node); err != nil {
        c.JSON(consts.StatusBadRequest, model.ErrorWithCode(model.ErrInvalidParam))
        return
    }

    nodeService := service.NewNodeService()
    if err := nodeService.UpdateNode(&node); err != nil {
        c.JSON(consts.StatusInternalServerError, model.Error(model.CodeDatabaseError, err.Error()))
        return
    }

    c.JSON(consts.StatusOK, model.Success(node))
}

func DeleteNode(ctx context.Context, c *app.RequestContext) {
    nodeId := c.Param("nodeId")
    if nodeId == "" {
        c.JSON(consts.StatusBadRequest, model.ErrorWithCode(model.ErrInvalidParam))
        return
    }

    nodeService := service.NewNodeService()
    if err := nodeService.DeleteNode(nodeId); err != nil {
        c.JSON(consts.StatusInternalServerError, model.Error(model.CodeDatabaseError, err.Error()))
        return
    }

    c.JSON(consts.StatusOK, model.Success(nil))
}