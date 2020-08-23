package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-kit/kit/log"
	gokittransport "github.com/go-kit/kit/transport"
	gokithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/jnikolaeva/eshop-common/httpkit"
	"github.com/jnikolaeva/eshop-common/uuid"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/jnikolaeva/catalogservice/internal/catalog/application"
)

var (
	ErrBadRouting = errors.New("bad routing")
	ErrBadRequest = errors.New("bad request")
)

func MakeHandler(pathPrefix string, endpoints Endpoints, errorLogger log.Logger, metrics *httpkit.MetricsHolder) http.Handler {
	options := []gokithttp.ServerOption{
		gokithttp.ServerErrorEncoder(encodeErrorResponse),
		gokithttp.ServerErrorHandler(gokittransport.NewLogErrorHandler(errorLogger)),
	}

	listCatalogItemsHandler := gokithttp.NewServer(endpoints.ListCatalogItems, decodeListCatalogItemsRequest, encodeResponse, options...)
	getCatalogItemByIDHandler := gokithttp.NewServer(endpoints.GetCatalogItemByID, decodeGetCatalogItemByIDRequest, encodeResponse, options...)
	createCatalogItemHandler := gokithttp.NewServer(endpoints.CreateCatalogItem, decodeCreateCatalogItemRequest, encodeResponse, options...)

	r := mux.NewRouter()
	s := r.PathPrefix(pathPrefix).Subrouter()
	s.Handle("", httpkit.InstrumentingMiddleware(listCatalogItemsHandler, metrics, "ListCatalogItems")).Methods(http.MethodGet)
	s.Handle("", httpkit.InstrumentingMiddleware(createCatalogItemHandler, metrics, "CreateCatalogItem")).Methods(http.MethodPost)
	s.Handle("/{id}", httpkit.InstrumentingMiddleware(getCatalogItemByIDHandler, metrics, "GetCatalogItemByID")).Methods(http.MethodGet)
	return r
}

func decodeListCatalogItemsRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	query := r.URL.Query()
	pageCount, err := strconv.Atoi(query.Get("page_count"))
	if err != nil {
		pageCount = 10
	}
	pageAfter := query.Get("page_after")
	page := &application.PageSpec{
		Count: pageCount,
		After: pageAfter,
	}
	return &listCatalogItemsRequest{Page: page}, nil
}

func decodeCreateCatalogItemRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req createCatalogItemRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil && e != io.EOF {
		return nil, e
	}
	if req.Title == "" {
		return nil, errors.Wrap(ErrBadRequest, "missing required parameter 'title'")
	}
	if req.SKU == "" {
		return nil, errors.Wrap(ErrBadRequest, "missing required parameter 'sku'")
	}
	if req.PriceStr == "" {
		return nil, errors.Wrap(ErrBadRequest, "missing required parameter 'price'")
	}
	if req.AvailableQty == nil {
		return nil, errors.Wrap(ErrBadRequest, "missing required parameter 'availableQty'")
	}
	if req.Image == nil {
		return nil, errors.Wrap(ErrBadRequest, "missing required parameter 'image'")
	}
	if req.Image.URL == "" {
		return nil, errors.Wrap(ErrBadRequest, "missing required parameter 'image.url'")
	}
	if req.Image.Width == nil {
		return nil, errors.Wrap(ErrBadRequest, "missing required parameter 'image.width'")
	}
	if req.Image.Height == nil {
		return nil, errors.Wrap(ErrBadRequest, "missing required parameter 'image.height'")
	}

	req.Price, err = decimal.NewFromString(req.PriceStr)
	if err != nil {
		return nil, errors.Wrap(ErrBadRequest, err.Error())
	}

	return &req, nil
}

func decodeGetCatalogItemByIDRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	sID, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	id, err := uuid.FromString(sID)
	if err != nil {
		return nil, ErrBadRequest
	}
	return &id, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if response == nil {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeErrorResponse(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var errorResponse = translateError(err)
	w.WriteHeader(errorResponse.Status)
	_ = json.NewEncoder(w).Encode(errorResponse.Response)
}

type transportError struct {
	Status   int
	Response errorResponse
}

func translateError(err error) transportError {
	if errors.Is(err, ErrBadRequest) {
		return transportError{
			Status: http.StatusBadRequest,
			Response: errorResponse{
				Code:    101,
				Message: err.Error(),
			},
		}
	} else if errors.Is(err, application.ErrCatalogItemNotFound) {
		return transportError{
			Status: http.StatusNotFound,
			Response: errorResponse{
				Code:    102,
				Message: err.Error(),
			},
		}
	} else if errors.Is(err, application.ErrDuplicateCatalogItem) {
		return transportError{
			Status: http.StatusConflict,
			Response: errorResponse{
				Code:    103,
				Message: err.Error(),
			},
		}
	} else {
		return transportError{
			Status: http.StatusInternalServerError,
			Response: errorResponse{
				Code:    100,
				Message: "unexpected error",
			},
		}
	}
}
