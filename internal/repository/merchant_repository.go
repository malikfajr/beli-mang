package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/malikfajr/beli-mang/internal/entity"
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
	query := "INSERT INTO merchants(id, username_admin, name, category, image_url, lat, long) VALUES(@id, @username,  @name, @category, @image, @lat, @long) ON CONFLICT DO NOTHING"
	args := pgx.NamedArgs{
		"id":       merchant.Id,
		"username": merchant.Username,
		"name":     merchant.Name,
		"category": merchant.Category,
		"image":    merchant.ImageUrl,
		"lat":      merchant.Location.Lat,
		"long":     merchant.Location.Long,
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

	query := "SELECT id, username_admin, name, category, image_url, lat, long, created_at  FROM merchants WHERE username_admin = @username"
	args := pgx.NamedArgs{
		"username": username,
	}

	if params.Name != "" {
		query += " AND LOWER(name) like @name"
		args["name"] = "%" + params.Name + "%"
	}

	if params.Category != "" {
		query += " AND category = @category"
		args["category"] = params.Category
	}

	if params.CreatedAt != "" {
		query += " ORDER BY created_at " + params.CreatedAt
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

		rows.Scan(&merchant.Id, &merchant.Username, &merchant.Name, &merchant.Category, &merchant.ImageUrl, &coordinate.Lat, &coordinate.Long, &merchant.CreatedAt)
		merchant.Location = coordinate

		merchants = append(merchants, *merchant)
	}

	return merchants
}
