package dao

import (
	"tenet-server/database"
	"tenet-server/model"
)

type DocumentDAO struct{}

func NewDocumentDAO() *DocumentDAO {
	return &DocumentDAO{}
}

// GetList 获取文档列表
func (d *DocumentDAO) GetList(limit int) ([]model.Document, error) {
	var documents []model.Document
	err := database.DB.Limit(limit).Find(&documents).Error
	return documents, err
}

// GetById 根据ID获取文档
func (d *DocumentDAO) GetById(documentId string) (*model.Document, error) {
	var document model.Document
	err := database.DB.Where("document_id = ?", documentId).First(&document).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

// Create 创建文档
func (d *DocumentDAO) Create(document *model.Document) error {
	return database.DB.Create(document).Error
}

// Update 更新文档
func (d *DocumentDAO) Update(document *model.Document) error {
	return database.DB.Model(&model.Document{}).
		Where("document_id = ?", document.DocumentId).
		Updates(map[string]interface{}{
			"document_title": document.DocumentTitle,
			"cover_url":      document.CoverUrl,
			"modify_time":    document.ModifyTime,
		}).Error
}

// Delete 删除文档
func (d *DocumentDAO) Delete(documentId string) error {
	return database.DB.Where("document_id = ?", documentId).Delete(&model.Document{}).Error
}
