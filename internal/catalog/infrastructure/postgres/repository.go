package postgres

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jnikolaeva/eshop-common/uuid"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/jnikolaeva/catalogservice/internal/catalog/application"
)

const errUniqueConstraint = "23505"

type rawProduct struct {
	ID           string          `db:"id"`
	Title        string          `db:"title"`
	SKU          string          `db:"sku"`
	Price        decimal.Decimal `db:"price"`
	AvailableQty int             `db:"available_qty"`
	ImageURL     string          `db:"image_url"`
	ImageWidth   int             `db:"image_width"`
	ImageHeight  int             `db:"image_height"`
	Color        string          `db:"color"`
	Material     string          `db:"material"`
}

type repository struct {
	connPool *pgx.ConnPool
}

func New(connPool *pgx.ConnPool) application.Repository {
	return &repository{
		connPool: connPool,
	}
}

func (r *repository) NextID() application.ProductID {
	return application.ProductID(uuid.Generate())
}

func (r *repository) FindByID(id application.ProductID) (*application.Product, error) {
	var raw rawProduct
	query := `SELECT id, Title, sku, price, available_qty, image_url, image_width, image_height, color, material 
              FROM products WHERE id = $1`
	err := r.connPool.QueryRow(query, id.String()).Scan(
		&raw.ID,
		&raw.Title,
		&raw.SKU,
		&raw.Price,
		&raw.AvailableQty,
		&raw.ImageURL,
		&raw.ImageWidth,
		&raw.ImageHeight,
		&raw.Color,
		&raw.Material)
	if err != nil {
		if err == pgx.ErrNoRows {
			err = application.ErrProductNotFound
		}
		return nil, errors.WithStack(err)
	}
	return mapToProduct(raw)
}

func (r *repository) Find(pageSpec *application.PageSpec, filters *application.Filters) ([]*application.Product, error) {
	var items []*application.Product
	query := "SELECT id, title, sku, price, available_qty, image_url, image_width, image_height, color, material FROM products"

	args := applyFilters(&query, filters)
	applyPageSpec(&query, pageSpec)

	rows, err := r.connPool.Query(query, args...)
	if err != nil {
		return nil, errors.WithMessage(err, "Database error")
	}
	defer rows.Close()

	var item *application.Product
	var raw rawProduct
	for rows.Next() {
		err = rows.Scan(
			&raw.ID,
			&raw.Title,
			&raw.SKU,
			&raw.Price,
			&raw.AvailableQty,
			&raw.ImageURL,
			&raw.ImageWidth,
			&raw.ImageHeight,
			&raw.Color,
			&raw.Material)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		item, err = mapToProduct(raw)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *repository) Add(item application.Product) error {
	_, err := r.connPool.Exec(
		`INSERT INTO products (id, title, sku, price, available_qty, image_url, image_width, image_height, color, material) 
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		item.ID.String(),
		item.Title,
		item.SKU,
		item.Price,
		item.AvailableQty,
		item.Image.URL,
		item.Image.Width,
		item.Image.Height,
		item.Color,
		item.Material)
	if err != nil {
		pgErr, ok := err.(pgx.PgError)
		if ok && pgErr.Code == errUniqueConstraint {
			return application.ErrDuplicateProduct
		}
		return errors.WithStack(err)
	}
	return nil
}

func mapToProduct(raw rawProduct) (*application.Product, error) {
	itemID, _ := uuid.FromString(raw.ID)
	item := &application.Product{
		ID:           application.ProductID(itemID),
		Title:        raw.Title,
		SKU:          raw.SKU,
		Price:        raw.Price,
		AvailableQty: raw.AvailableQty,
		Image: &application.Image{
			URL:    raw.ImageURL,
			Width:  raw.ImageWidth,
			Height: raw.ImageHeight,
		},
		Color:    raw.Color,
		Material: raw.Material,
	}
	return item, nil
}

func applyFilters(query *string, filters *application.Filters) (args []interface{}) {
	if filters == nil {
		return args
	}
	var conditions []string
	argInd := 1
	if filters.Price.Min != nil {
		conditions = append(conditions, fmt.Sprintf("price >= $%d", argInd))
		args = append(args, filters.Price.Min)
	}
	if filters.Price.Max != nil {
		argInd++
		conditions = append(conditions, fmt.Sprintf("price <= $%d", argInd))
		args = append(args, filters.Price.Max)
	}

	conditions, args = applyStringOrFilter("color", filters.Color, conditions, args)
	conditions, args = applyStringOrFilter("material", filters.Material, conditions, args)

	where := strings.Join(conditions, " AND ")
	if where != "" {
		*query += " WHERE " + where
	}
	return args
}

func applyStringOrFilter(field string, filter *[]string, conditions []string, args []interface{}) ([]string, []interface{}) {
	if filter == nil || len(*filter) == 0 {
		return conditions, args
	}
	cnt := len(args)
	var clauses []string
	for i, v := range *filter {
		clauses = append(clauses, fmt.Sprintf("%v = $%d", field, cnt+i+1))
		args = append(args, v)
	}
	conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(clauses, " OR ")))
	return conditions, args
}

func applyPageSpec(query *string, pageSpec *application.PageSpec) {
	if pageSpec == nil {
		return
	}
	if pageSpec.Number == 1 {
		*query += fmt.Sprintf(" LIMIT %d", pageSpec.Size)
	} else {
		*query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSpec.Size, (pageSpec.Number-1)*pageSpec.Size)
	}
}
