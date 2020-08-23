package http

import (
	"github.com/shopspring/decimal"

	"github.com/jnikolaeva/catalogservice/internal/catalog/application"
)

type listCatalogItemsRequest struct {
	Spec *application.PageSpec
}

type listCatalogItemsResponse struct {
	Items []*catalogItem `json:"items"`
	After string         `json:"after,omitempty"`
	Count int            `json:"count"`
}

type catalogItem struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	SKU          string          `json:"sku"`
	Price        decimal.Decimal `json:"price"`
	AvailableQty int             `json:"available_qty"`
	Image        image           `json:"image"`
}

type image struct {
	URL    string `json:"url"`
	Width  *int   `json:"width"`
	Height *int   `json:"height"`
}

type createCatalogItemRequest struct {
	Title        string `json:"title"`
	SKU          string `json:"sku"`
	PriceStr     string `json:"price"`
	AvailableQty *int   `json:"available_qty"`
	Price        decimal.Decimal
	Image        *image `json:"image"`
}

func (c *createCatalogItemRequest) GetTitle() string {
	return c.Title
}

func (c *createCatalogItemRequest) GetSKU() string {
	return c.SKU
}

func (c *createCatalogItemRequest) GetPrice() decimal.Decimal {
	return c.Price
}

func (c *createCatalogItemRequest) GetAvailableQty() int {
	return *c.AvailableQty
}

func (c *createCatalogItemRequest) GetImageURL() string {
	return c.Image.URL
}

func (c *createCatalogItemRequest) GetImageWidth() int {
	return *c.Image.Width
}

func (c *createCatalogItemRequest) GetImageHeight() int {
	return *c.Image.Height
}

type createCatalogItemResponse struct {
	ID string `json:"id"`
}

type getCatalogItemByIDResponse struct {
	catalogItem
}

type errorResponse struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
}
