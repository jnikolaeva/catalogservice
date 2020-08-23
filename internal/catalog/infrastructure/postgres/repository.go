package postgres

import (
	"github.com/jackc/pgx"
	"github.com/jnikolaeva/eshop-common/uuid"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/jnikolaeva/catalogservice/internal/catalog/application"
)

const errUniqueConstraint = "23505"

type rawCatalogItem struct {
	ID           string          `db:"id"`
	Title        string          `db:"title"`
	SKU          string          `db:"sku"`
	Price        decimal.Decimal `db:"price"`
	AvailableQty int             `db:"available_qty"`
	ImageURL     string          `db:"image_url"`
	ImageWidth   int             `db:"image_width"`
	ImageHeight  int             `db:"image_height"`
}

type repository struct {
	connPool *pgx.ConnPool
}

func New(connPool *pgx.ConnPool) application.Repository {
	return &repository{
		connPool: connPool,
	}
}

func (r *repository) NextID() application.CatalogItemID {
	return application.CatalogItemID(uuid.Generate())
}

func (r *repository) FindByID(id application.CatalogItemID) (*application.CatalogItem, error) {
	var raw rawCatalogItem
	query := "SELECT id, Title, sku, price, available_qty, image_url, image_width, image_height FROM products WHERE id = $1"
	err := r.connPool.QueryRow(query, id.String()).Scan(
		&raw.ID,
		&raw.Title,
		&raw.SKU,
		&raw.Price,
		&raw.AvailableQty,
		&raw.ImageURL,
		&raw.ImageWidth,
		&raw.ImageHeight)
	if err != nil {
		if err == pgx.ErrNoRows {
			err = application.ErrCatalogItemNotFound
		}
		return nil, errors.WithStack(err)
	}

	return mapToCatalogItem(raw)
}

func (r *repository) Find(spec *application.PageSpec) ([]*application.CatalogItem, error) {
	var items []*application.CatalogItem
	// TODO: add pagination
	query := `SELECT id, title, sku, price, available_qty, image_url, image_width, image_height
		FROM products`
	rows, err := r.connPool.Query(query)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	var item *application.CatalogItem
	var raw rawCatalogItem
	for rows.Next() {
		err = rows.Scan(
			&raw.ID,
			&raw.Title,
			&raw.SKU,
			&raw.Price,
			&raw.AvailableQty,
			&raw.ImageURL,
			&raw.ImageWidth,
			&raw.ImageHeight)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		item, err = mapToCatalogItem(raw)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *repository) Add(item application.CatalogItem) error {
	_, err := r.connPool.Exec(
		"INSERT INTO products (id, title, sku, price, available_qty, image_url, image_width, image_height) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		item.ID.String(), item.Title, item.SKU, item.Price, item.AvailableQty, item.Image.URL, item.Image.Width, item.Image.Height)
	if err != nil {
		pgErr, ok := err.(pgx.PgError)
		if ok && pgErr.Code == errUniqueConstraint {
			return application.ErrDuplicateCatalogItem
		}
		return errors.WithStack(err)
	}
	return nil
}

func mapToCatalogItem(raw rawCatalogItem) (*application.CatalogItem, error) {
	itemID, _ := uuid.FromString(raw.ID)
	item := &application.CatalogItem{
		ID:           application.CatalogItemID(itemID),
		Title:        raw.Title,
		SKU:          raw.SKU,
		Price:        raw.Price,
		AvailableQty: raw.AvailableQty,
		Image: &application.Image{
			URL:    raw.ImageURL,
			Width:  raw.ImageWidth,
			Height: raw.ImageHeight,
		},
	}
	return item, nil
}
