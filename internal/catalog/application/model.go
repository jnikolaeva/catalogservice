package application

import (
	"github.com/jnikolaeva/eshop-common/uuid"
	"github.com/shopspring/decimal"
)

type CatalogItemID uuid.UUID

func (u CatalogItemID) String() string {
	return uuid.UUID(u).String()
}

type PageSpec struct {
	Count int
	After string
}

type CatalogPage struct {
	Items []*CatalogItem
	Count int
	After string
}

type CatalogItem struct {
	ID           CatalogItemID
	Title        string
	SKU          string
	Price        decimal.Decimal
	AvailableQty int
	Image        *Image
}

type Image struct {
	URL    string
	Width  int
	Height int
}

type Repository interface {
	NextID() CatalogItemID
	FindByID(id CatalogItemID) (*CatalogItem, error)
	Find(spec *PageSpec) ([]*CatalogItem, error)
	Add(item CatalogItem) error
}
