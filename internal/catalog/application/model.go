package application

import (
	"github.com/jnikolaeva/eshop-common/uuid"
	"github.com/shopspring/decimal"
)

type ProductID uuid.UUID

func (u ProductID) String() string {
	return uuid.UUID(u).String()
}

type PageSpec struct {
	Size   int
	Number int
}

type DecimalRangeFilter struct {
	Min *decimal.Decimal
	Max *decimal.Decimal
}

type StringOrFilter []string

type Filters struct {
	Price    DecimalRangeFilter
	Color    *[]string
	Material *[]string
}

type Product struct {
	ID           ProductID
	Title        string
	SKU          string
	Price        decimal.Decimal
	AvailableQty int
	Image        *Image
	Color        string
	Material     string
}

type Image struct {
	URL    string
	Width  int
	Height int
}

type Repository interface {
	NextID() ProductID
	FindByID(id ProductID) (*Product, error)
	Find(spec *PageSpec, filters *Filters) ([]*Product, error)
	Add(item Product) error
}
