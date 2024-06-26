package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/malikfajr/beli-mang/internal/entity"
	"github.com/mmcloughlin/geohash"
)

type MerchantRepo struct{}

func (m *MerchantRepo) GetById(ctx context.Context, pool *pgxpool.Pool, merchantId string) (*entity.Merchant, error) {
	merchant := &entity.Merchant{}
	coordinate := &entity.Coordinate{}

	query := "SELECT id, username_admin, name, category, image_url, lat, long, created_at  FROM merchants WHERE id = $1 LIMIT 1;"

	err := pool.QueryRow(ctx, query, merchantId).Scan(&merchant.Id, &merchant.Username, &merchant.Name, &merchant.Category, &merchant.ImageUrl, &coordinate.Lat, &coordinate.Long, &merchant.CreatedAt)
	merchant.Location = coordinate
	if err != nil {
		return nil, errors.New("merchant not found")
	}

	return merchant, nil
}

func (m *MerchantRepo) Insert(ctx context.Context, pool *pgxpool.Pool, merchant *entity.Merchant) error {
	query := "INSERT INTO merchants(id, username_admin, name, category, image_url, lat, long, geohash) VALUES(@id, @username,  @name, @category, @image, @lat, @long, @geohash) ON CONFLICT DO NOTHING"
	args := pgx.NamedArgs{
		"id":       merchant.Id,
		"username": merchant.Username,
		"name":     merchant.Name,
		"category": merchant.Category,
		"image":    merchant.ImageUrl,
		"lat":      merchant.Location.Lat,
		"long":     merchant.Location.Long,
		"geohash":  geohash.Encode(merchant.Location.Lat, merchant.Location.Long),
	}

	tag, err := pool.Exec(ctx, query, args)
	if err != nil {
		panic(err)
	}

	if tag.RowsAffected() == 0 {
		return errors.New("Failed to create merchants")
	}

	return nil
}

func (m *MerchantRepo) GetAll(ctx context.Context, pool *pgxpool.Pool, username string, params *entity.MerchantParams) []entity.Merchant {

	query := "SELECT id, username_admin, name, category, image_url, lat, long, geohash, created_at  FROM merchants WHERE TRUE "
	args := pgx.NamedArgs{
		// "username": username,
	}

	if params.Name != "" {
		query += " AND LOWER(name) like @name"
		args["name"] = "%" + strings.ToLower(params.Name) + "%"
	}

	if params.Category != "" {
		query += " AND category = @category"
		args["category"] = params.Category
	}

	if params.CreatedAt != "" {
		query += " ORDER BY created_at " + params.CreatedAt
	} else {
		query += " ORDER BY created_at desc"
	}

	query += " LIMIT @limit OFFSET @offset"
	args["limit"] = params.Limit
	args["offset"] = params.Offset

	rows, err := pool.Query(ctx, query, args)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	merchants := make([]entity.Merchant, 0)

	for rows.Next() {
		merchant := &entity.Merchant{}
		coordinate := &entity.Coordinate{}

		rows.Scan(&merchant.Id, &merchant.Username, &merchant.Name, &merchant.Category, &merchant.ImageUrl, &coordinate.Lat, &coordinate.Long, &merchant.Geohash, &merchant.CreatedAt)
		merchant.Location = coordinate

		merchants = append(merchants, *merchant)
	}

	return merchants
}

func (m *MerchantRepo) GetTotalMerchant(ctx context.Context, pool *pgxpool.Pool, username string, params *entity.MerchantParams) int {
	query := "SELECT COUNT(id) FROM merchants WHERE TRUE "
	args := pgx.NamedArgs{
		// "username": username,
	}

	if params.Name != "" {
		query += " AND LOWER(name) like @name"
		args["name"] = "%" + strings.ToLower(params.Name) + "%"
	}

	if params.Category != "" {
		query += " AND category = @category"
		args["category"] = params.Category
	}

	var total int
	err := pool.QueryRow(ctx, query, args).Scan(&total)
	if err != nil {
		return 0
	}

	return total
}

func (m *MerchantRepo) AddProduct(ctx context.Context, pool *pgxpool.Pool, product *entity.Product) error {
	query := "INSERT INTO products(id, merchant_id, name, category, price, image_url) VALUES($1, $2, $3, $4, $5, $6)"

	tag, err := pool.Exec(ctx, query, product.Id, product.MerchantId, product.Name, product.Category, product.Price, product.ImageUrl)
	if err != nil {
		panic(err)
	}

	if tag.RowsAffected() == 0 {
		return errors.New("Failed to insert items")
	}

	return nil
}

func (m *MerchantRepo) GetProducts(ctx context.Context, pool *pgxpool.Pool, params *entity.ProductParams) []entity.Product {
	query := "SELECT id, name, category, price, image_url, created_at FROM products WHERE merchant_id = @merchant_id "
	args := pgx.NamedArgs{
		"merchant_id": params.MerchantId,
		"limit":       params.Limit,
		"offset":      params.Offset,
	}

	if params.Id != "" {
		query += " AND id = @id"
		args["id"] = params.Id
	}

	if params.Name != "" {
		query += " AND LOWER(name) LIKE @name"
		args["name"] = "%" + params.Name + "%"
	}

	if params.Category != "" {
		query += " AND category = @category"
		args["category"] = params.Category
	}

	if params.CreatedAt != "" {
		query += " ORDER BY created_at " + params.CreatedAt
	} else {
		query += " ORDER BY created_at desc"
	}

	query += " LIMIT @limit OFFSET @offset"

	rows, err := pool.Query(ctx, query, args)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	products := make([]entity.Product, 0)
	for rows.Next() {
		product := &entity.Product{}
		rows.Scan(&product.Id, &product.Name, &product.Category, &product.Price, &product.ImageUrl, &product.CreatedAt)
		products = append(products, *product)
	}

	return products
}

func (m *MerchantRepo) GetTotalProduct(ctx context.Context, pool *pgxpool.Pool, params *entity.ProductParams) int {
	var total int
	query := "SELECT COUNT(id) products WHERE merchant_id = @merchant_id "
	args := pgx.NamedArgs{
		"merchant_id": params.MerchantId,
	}

	if params.Id != "" {
		query += " AND id = @id"
		args["id"] = params.Id
	}

	if params.Name != "" {
		query += " AND LOWER(name) LIKE @name"
		args["name"] = "%" + params.Name + "%"
	}

	if params.Category != "" {
		query += " AND category = @category"
		args["category"] = params.Category
	}

	query += " LIMIT 1"

	err := pool.QueryRow(ctx, query, args).Scan(&total)
	if err != nil {
		return 0
	}

	return total
}
