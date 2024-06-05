package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/malikfajr/beli-mang/internal/entity"
	"github.com/malikfajr/beli-mang/internal/entity/converter"
	"github.com/mmcloughlin/geohash"
)

type PurchaseRepo struct{}

func (p *PurchaseRepo) GetMerchantNearby(ctx context.Context, pool *pgxpool.Pool, lat float64, long float64, params *converter.MerchanNearbyParams) []converter.MerchanNearby {
	userGeohash := geohash.Encode(lat, long)
	geoPrefix := userGeohash[:3]

	query := `
		SELECT
		m.id AS merchantId,
		m.name AS merchantName,
		m.category AS merchantCategory,
		m.image_url AS merchantImageUrl,
		m.lat,
		m.long,
		m.created_at AS merchantCreatedAt,
		(SELECT 
			json_agg(
				json_build_object(
					'itemId', p.id,
					'name', p.name,
					'productCategory', p.category,
					'price', p.price,
					'imageUrl', p.image_url,
					'createdAt', p.created_at
				)
			)
		FROM products p WHERE m.id = p.merchant_id) AS items,
		haversine(@lat, @long, lat, long) AS distance
	FROM
		merchants m
	WHERE
		m.geohash LIKE @geoparam
	`

	args := pgx.NamedArgs{
		"lat":      lat,
		"long":     long,
		"limit":    int(params.Limit),
		"offset":   int(params.Offset),
		"geoparam": geoPrefix + "%",
	}

	if params.MerchantId != "" {
		query += " AND m.id = @m_id"
		args["m_id"] = params.MerchantId
	}

	if params.Category != "" {
		query += " AND m.category = @m_category"
		args["m_category"] = params.Category
	}

	if params.Name != "" {
		query += " AND m.name = @m_name"
		args["m_name"] = params.Name
	}

	query += `
	ORDER BY distance
	LIMIT @limit OFFSET @offset;`

	rows, err := pool.Query(ctx, query, args)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var data []converter.MerchanNearby = []converter.MerchanNearby{}

	for rows.Next() {
		var products []entity.Product = []entity.Product{}
		var productJSON []byte
		merchant := &entity.Merchant{
			Location: &entity.Coordinate{},
		}

		rows.Scan(&merchant.Id, &merchant.Name, &merchant.Category, &merchant.ImageUrl, &merchant.Location.Lat, &merchant.Location.Long, &merchant.CreatedAt, &productJSON, nil)

		if productJSON != nil {
			err := json.Unmarshal(productJSON, &products)
			if err != nil {
				panic(err)
			}
		}

		data = append(data, converter.MerchanNearby{
			Merchant: *merchant,
			Items:    products,
		})
	}

	return data
}

func (p *PurchaseRepo) GetMerchantBydIds(ctx context.Context, pool *pgxpool.Pool, merchantIds []string, lat float64, long float64) {
	hash := geohash.Encode(lat, long)

	query := "SELECT id, username, name, category, image_url, lat, long, created_at, haversine($1, $2, lat, long) AS distance FROM merchants WHERE geohash LIKE $3, id = ANY($4) ORDER BY distance"

	rows, err := pool.Query(ctx, query, lat, long, hash[:3], merchantIds)
	if err != nil {
		panic(err)
	}

	for rows.Next() {

	}

}
