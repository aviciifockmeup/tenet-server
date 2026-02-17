package service

import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "tenet-server/dao"
    "tenet-server/model"
    "time"
)

type DocumentService struct {
    documentDAO *dao.DocumentDAO
    nodeDAO     *dao.NodeDAO
}

func NewDocumentService() *DocumentService {
    return &DocumentService{
        documentDAO: dao.NewDocumentDAO(),
        nodeDAO:     dao.NewNodeDAO(),
    }
}

// GetDocumentList 获取文档列表
func (s *DocumentService) GetDocumentList() ([]model.Document, error) {
    return s.documentDAO.GetList(10)
}

// OpenDocument 打开文档，返回文档信息+所有节点
func (s *DocumentService) OpenDocument(documentId string) (*model.DocumentInfo, error) {
    // 获取文档信息
    document, err := s.documentDAO.GetById(documentId)
    if err != nil {
        return nil, err
    }

    // 获取所有节点
    nodes, err := s.nodeDAO.GetByDocumentId(documentId)
    if err != nil {
        return nil, err
    }

    // 组装返回结果
    docInfo := &model.DocumentInfo{
        DocumentId: document.DocumentId,
        Title:      document.DocumentTitle,
        Nodes:      nodes,
    }

    return docInfo, nil
}

// CreateDocument 创建新文档
func (s *DocumentService) CreateDocument(title string, userId string) (*model.Document, error) {
    now := time.Now()
    document := &model.Document{
        DocumentId:    s.generateDocumentId(),
        DocumentTitle: title,
        UserId:        userId,
        CategoryId:    "default",
        Status:        0,
        CreateTime:    now,
        ModifyTime:    now,
    }

    err := s.documentDAO.Create(document)
    if err != nil {
        return nil, err
    }

    return document, nil
}

// UpdateDocumentTitle 更新文档标题
func (s *DocumentService) UpdateDocumentTitle(documentId string, title string) error {
    document := &model.Document{
        DocumentId:    documentId,
        DocumentTitle: title,
        ModifyTime:    time.Now(),
    }
    return s.documentDAO.Update(document)
}

// DeleteDocument 删除文档
func (s *DocumentService) DeleteDocument(documentId string) error {
    return s.documentDAO.Delete(documentId)
}

// generateDocumentId 生成文档ID（简化版，实际应该用雪花算法或UUID）
func (s *DocumentService) generateDocumentId() string {
    b := make([]byte, 12)
    rand.Read(b)
    return fmt.Sprintf("doc_%s", base64.URLEncoding.EncodeToString(b))
}