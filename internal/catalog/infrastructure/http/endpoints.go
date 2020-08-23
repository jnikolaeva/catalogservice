package http

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/jnikolaeva/eshop-common/uuid"

	"github.com/jnikolaeva/catalogservice/internal/catalog/application"
)

type Endpoints struct {
	ListCatalogItems   endpoint.Endpoint
	GetCatalogItemByID endpoint.Endpoint
	CreateCatalogItem  endpoint.Endpoint
}

func MakeEndpoints(s application.Service) Endpoints {
	return Endpoints{
		ListCatalogItems:   makeListCatalogItemsEndpoint(s),
		GetCatalogItemByID: makeGetCatalogItemByIDEndpoint(s),
		CreateCatalogItem:  makeCreateCatalogItemEndpoint(s),
	}
}

func makeListCatalogItemsEndpoint(s application.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*listCatalogItemsRequest)
		items, err := s.Find(req.Page)
		if err != nil {
			return nil, err
		}
		count := len(items)
		catalogItems := make([]*catalogItem, count)
		for i, item := range items {
			catalogItems[i] = toCatalogItem(item)
		}
		res := &listCatalogItemsResponse{
			Items: catalogItems,
			Count: count,
		}
		if count > 0 {
			res.After = items[count-1].ID.String()
		}
		return res, nil
	}
}

func makeGetCatalogItemByIDEndpoint(s application.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		id := request.(*uuid.UUID)
		item, err := s.FindByID(*id)
		if err != nil {
			return nil, err
		}
		return &getCatalogItemByIDResponse{*toCatalogItem(item)}, nil
	}
}

func makeCreateCatalogItemEndpoint(s application.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*createCatalogItemRequest)
		id, err := s.Create(req)
		if err != nil {
			return nil, err
		}
		return &createCatalogItemResponse{ID: id.String()}, nil
	}
}

func toCatalogItem(item *application.CatalogItem) *catalogItem {
	return &catalogItem{
		ID:           item.ID.String(),
		Title:        item.Title,
		SKU:          item.SKU,
		Price:        item.Price,
		AvailableQty: item.AvailableQty,
		Image: image{
			URL:    item.Image.URL,
			Width:  &item.Image.Width,
			Height: &item.Image.Height,
		},
	}
}
