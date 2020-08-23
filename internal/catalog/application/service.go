package application

import (
	"github.com/jnikolaeva/eshop-common/uuid"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

var (
	ErrCatalogItemNotFound  = errors.New("catalog item not found")
	ErrDuplicateCatalogItem = errors.New("catalog item with such SKU already exists")
)

type CatalogItemParams interface {
	GetTitle() string
	GetSKU() string
	GetPrice() decimal.Decimal
	GetAvailableQty() int
	GetImageURL() string
	GetImageWidth() int
	GetImageHeight() int
}

type Service interface {
	FindByID(id uuid.UUID) (*CatalogItem, error)
	Find(spec *PageSpec) ([]*CatalogItem, error)
	Create(params CatalogItemParams) (CatalogItemID, error)
}

type service struct {
	repo Repository
}

func NewService(repository Repository) Service {
	return &service{repo: repository}
}

func (s *service) FindByID(id uuid.UUID) (*CatalogItem, error) {
	return s.repo.FindByID(CatalogItemID(id))
}

func (s *service) Find(spec *PageSpec) ([]*CatalogItem, error) {
	return s.repo.Find(spec)
}

func (s *service) Create(params CatalogItemParams) (CatalogItemID, error) {
	id := s.repo.NextID()
	item := CatalogItem{
		ID:           id,
		Title:        params.GetTitle(),
		SKU:          params.GetSKU(),
		Price:        params.GetPrice(),
		AvailableQty: params.GetAvailableQty(),
		Image: &Image{
			URL:    params.GetImageURL(),
			Width:  params.GetImageWidth(),
			Height: params.GetImageHeight(),
		},
	}
	err := s.repo.Add(item)
	if err != nil {
		return CatalogItemID{}, errors.WithStack(err)
	}
	return id, nil
}
