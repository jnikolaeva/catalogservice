package http

import (
	"github.com/shopspring/decimal"

	"github.com/jnikolaeva/catalogservice/internal/catalog/application"
)

type listProductsRequest struct {
	PageSpec *application.PageSpec
	Filters  *application.Filters
}

type listProductsResponse struct {
	Items []*product `json:"items"`
	After string     `json:"after,omitempty"`
	Count int        `json:"count"`
}

type product struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	SKU          string          `json:"sku"`
	Price        decimal.Decimal `json:"price"`
	AvailableQty int             `json:"available_qty"`
	Image        image           `json:"image"`
	Color        string          `json:"color"`
	Material     string          `json:"material"`
}

type image struct {
	URL    string `json:"url"`
	Width  *int   `json:"width"`
	Height *int   `json:"height"`
}

type createProductRequest struct {
	Title        string `json:"title"`
	SKU          string `json:"sku"`
	PriceStr     string `json:"price"`
	AvailableQty *int   `json:"available_qty"`
	Price        decimal.Decimal
	Image        *image `json:"image"`
	Color        string `json:"color"`
	Material     string `json:"material"`
}

func (c *createProductRequest) GetTitle() string {
	return c.Title
}

func (c *createProductRequest) GetSKU() string {
	return c.SKU
}

func (c *createProductRequest) GetPrice() decimal.Decimal {
	return c.Price
}

func (c *createProductRequest) GetAvailableQty() int {
	return *c.AvailableQty
}

func (c *createProductRequest) GetImageURL() string {
	return c.Image.URL
}

func (c *createProductRequest) GetImageWidth() int {
	return *c.Image.Width
}

func (c *createProductRequest) GetImageHeight() int {
	return *c.Image.Height
}

func (c *createProductRequest) GetColor() string {
	return c.Color
}

func (c *createProductRequest) GetMaterial() string {
	return c.Material
}

type createProductResponse struct {
	ID string `json:"id"`
}

type getProductByIDResponse struct {
	product
}

type errorResponse struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
}
