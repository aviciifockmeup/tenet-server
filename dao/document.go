package dao

import (
    "tenet-server/database"
    "tenet-server/model"
)

type DocumentDAO struct {}

func NewDocumentDAO() *DocumentDAO {
	return &DocumentDAO{}
}

func (d *DocumentDAO) GetList(limit int) ([]model.Document, error) {
    var documents []model.Document
    err := database.DB.Limit(limit).Find(&documents).Error
    return documents, err
}