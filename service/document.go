package service

import (
    "tenet-server/dao"
    "tenet-server/model"
)

type DocumentService struct {
    dao *dao.DocumentDAO
}

func NewDocumentService() *DocumentService {
    return &DocumentService{
        dao: dao.NewDocumentDAO(),
    }
}

func (s *DocumentService) GetDocumentList() ([]model.Document, error) {
	return s.dao.GetList(10)
}