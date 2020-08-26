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

const defaultPageSize = 10

var (
	ErrBadRouting = errors.New("bad routing")
	ErrBadRequest = errors.New("bad request")
)

func MakeHandler(pathPrefix string, endpoints Endpoints, errorLogger log.Logger, metrics *httpkit.MetricsHolder) http.Handler {
	options := []gokithttp.ServerOption{
		gokithttp.ServerErrorEncoder(encodeErrorResponse),
		gokithttp.ServerErrorHandler(gokittransport.NewLogErrorHandler(errorLogger)),
	}

	listProductsHandler := gokithttp.NewServer(endpoints.ListProducts, decodeListProductsRequest, encodeResponse, options...)
	getProductByIDHandler := gokithttp.NewServer(endpoints.GetProductByID, decodeGetProductByIDRequest, encodeResponse, options...)
	createProductHandler := gokithttp.NewServer(endpoints.CreateProduct, decodeCreateProductRequest, encodeResponse, options...)

	r := mux.NewRouter()
	s := r.PathPrefix(pathPrefix).Subrouter()
	s.Handle("/products", httpkit.InstrumentingMiddleware(listProductsHandler, metrics, "ListProducts")).Methods(http.MethodGet)
	s.Handle("/products", httpkit.InstrumentingMiddleware(createProductHandler, metrics, "CreateProduct")).Methods(http.MethodPost)
	s.Handle("/products/{id}", httpkit.InstrumentingMiddleware(getProductByIDHandler, metrics, "GetProductByID")).Methods(http.MethodGet)
	return r
}

func decodeListProductsRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	query := r.URL.Query()
	pageSize, err := strconv.Atoi(query.Get("page_size"))
	if err != nil || pageSize <= 0 {
		pageSize = defaultPageSize
	}
	pageNum, err := strconv.Atoi(query.Get("page_num"))
	if err != nil || pageNum <= 0 {
		pageNum = 1
	}
	pageSpec := &application.PageSpec{
		Size:   pageSize,
		Number: pageNum,
	}
	result := &listProductsRequest{
		PageSpec: pageSpec,
		Filters: &application.Filters{
			Color:    &[]string{},
			Material: &[]string{},
		},
	}
	if err := parseFilter(query, "price", parseDecimalRangeFilter, &result.Filters.Price); err != nil {
		return nil, errors.Wrap(ErrBadRequest, err.Error())
	}
	if err := parseFilter(query, "color", parseStringOrFilter, result.Filters.Color); err != nil {
		return nil, errors.Wrap(ErrBadRequest, err.Error())
	}
	if err := parseFilter(query, "material", parseStringOrFilter, result.Filters.Material); err != nil {
		return nil, errors.Wrap(ErrBadRequest, err.Error())
	}
	return result, nil
}

func decodeCreateProductRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req createProductRequest
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
	if req.Color == "" {
		return nil, errors.Wrap(ErrBadRequest, "missing required parameter 'color'")
	}
	if req.Material == "" {
		return nil, errors.Wrap(ErrBadRequest, "missing required parameter 'material'")
	}

	req.Price, err = decimal.NewFromString(req.PriceStr)
	if err != nil {
		return nil, errors.Wrap(ErrBadRequest, err.Error())
	}

	return &req, nil
}

func decodeGetProductByIDRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
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
	} else if errors.Is(err, application.ErrProductNotFound) {
		return transportError{
			Status: http.StatusNotFound,
			Response: errorResponse{
				Code:    102,
				Message: err.Error(),
			},
		}
	} else if errors.Is(err, application.ErrDuplicateProduct) {
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
