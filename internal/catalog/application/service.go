package application

import (
	"github.com/jnikolaeva/eshop-common/uuid"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

var (
	ErrProductNotFound  = errors.New("product not found")
	ErrDuplicateProduct = errors.New("product with such SKU already exists")
)

type ProductParams interface {
	GetTitle() string
	GetSKU() string
	GetPrice() decimal.Decimal
	GetAvailableQty() int
	GetImageURL() string
	GetImageWidth() int
	GetImageHeight() int
	GetColor() string
	GetMaterial() string
}

type Service interface {
	FindByID(id uuid.UUID) (*Product, error)
	Find(spec *PageSpec, filters *Filters) ([]*Product, error)
	Create(params ProductParams) (ProductID, error)
}

type service struct {
	repo Repository
}

func NewService(repository Repository) Service {
	return &service{repo: repository}
}

func (s *service) FindByID(id uuid.UUID) (*Product, error) {
	return s.repo.FindByID(ProductID(id))
}

func (s *service) Find(spec *PageSpec, filters *Filters) ([]*Product, error) {
	return s.repo.Find(spec, filters)
}

func (s *service) Create(params ProductParams) (ProductID, error) {
	id := s.repo.NextID()
	item := Product{
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
		Color:    params.GetColor(),
		Material: params.GetMaterial(),
	}
	err := s.repo.Add(item)
	if err != nil {
		return ProductID{}, errors.WithStack(err)
	}
	return id, nil
}
