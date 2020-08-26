package http

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/jnikolaeva/eshop-common/uuid"

	"github.com/jnikolaeva/catalogservice/internal/catalog/application"
)

type Endpoints struct {
	ListProducts   endpoint.Endpoint
	GetProductByID endpoint.Endpoint
	CreateProduct  endpoint.Endpoint
}

func MakeEndpoints(s application.Service) Endpoints {
	return Endpoints{
		ListProducts:   makeListProductsEndpoint(s),
		GetProductByID: makeGetProductByIDEndpoint(s),
		CreateProduct:  makeCreateProductEndpoint(s),
	}
}

func makeListProductsEndpoint(s application.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*listProductsRequest)
		items, err := s.Find(req.PageSpec, req.Filters)
		if err != nil {
			return nil, err
		}
		count := len(items)
		products := make([]*product, count)
		for i, item := range items {
			products[i] = toProduct(item)
		}
		res := &listProductsResponse{
			Items: products,
			Count: count,
		}
		return res, nil
	}
}

func makeGetProductByIDEndpoint(s application.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		id := request.(*uuid.UUID)
		item, err := s.FindByID(*id)
		if err != nil {
			return nil, err
		}
		return &getProductByIDResponse{*toProduct(item)}, nil
	}
}

func makeCreateProductEndpoint(s application.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*createProductRequest)
		id, err := s.Create(req)
		if err != nil {
			return nil, err
		}
		return &createProductResponse{ID: id.String()}, nil
	}
}

func toProduct(item *application.Product) *product {
	return &product{
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
		Color:    item.Color,
		Material: item.Material,
	}
}
