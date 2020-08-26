package http

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/jnikolaeva/catalogservice/internal/catalog/application"
)

const valuesSeparator = ","

type parser func(value string, filter interface{}) error

func parseFilter(values url.Values, paramName string, parser parser, filter interface{}) error {
	paramValue := values.Get(paramName)
	if paramValue == "" {
		return nil
	}
	if err := parser(paramValue, filter); err != nil {
		return err
	}
	return nil
}

func parseStringOrFilter(value string, filter interface{}) error {
	f, ok := filter.(*[]string)
	if !ok {
		return errors.New("invalid type for filter")
	}
	*f = strings.Split(value, valuesSeparator)
	return nil
}

func parseDecimalRangeFilter(value string, filter interface{}) error {
	f, ok := filter.(*application.DecimalRangeFilter)
	if !ok {
		return errors.New("wrong filter type")
	}
	values := strings.Split(value, valuesSeparator)
	if len(values) == 0 {
		return nil
	}
	var err error
	if f.Min == nil {
		f.Min = &decimal.Decimal{}
	}

	if *f.Min, err = decimal.NewFromString(values[0]); err != nil {
		return errors.New("can't parse min value for the filter")
	}

	if len(values) > 1 {
		if f.Max == nil {
			f.Max = &decimal.Decimal{}
		}
		if *f.Max, err = decimal.NewFromString(values[1]); err != nil {
			return errors.New("can't parse min value for the filter")
		}
	}
	return nil
}
