package repository

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/malikfajr/beli-mang/internal/driver/db"
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

func (p *PurchaseRepo) TotalMerchantNearby(ctx context.Context, pool *pgxpool.Pool, lat float64, long float64, params *converter.MerchanNearbyParams) int {
	userGeohash := geohash.Encode(lat, long)
	geoPrefix := userGeohash[:3]

	query := `
		SELECT
		COUNT(m.id)
	FROM
		merchants m
	WHERE
		m.geohash LIKE @geoparam
	`

	args := pgx.NamedArgs{
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

	var total int
	err := pool.QueryRow(ctx, query, args).Scan(&total)
	if err != nil {
		return 0
	}

	return total
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

func (p *PurchaseRepo) GetHistory(ctx context.Context, pool *pgxpool.Pool, params *entity.OrderHistoryParams) []entity.OrderHistory {

	rows, err := pool.Query(ctx, p.generateQueryOrderHistory(params))
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	history := []entity.OrderHistory{}
	merchants := make(map[string]map[string]*entity.OrderDetail)

	for rows.Next() {
		var orderID, merchantID, merchantName, merchantCategory, merchantImageURL, productID, productName, productCategory, productImageURL string
		var merchantLat, merchantLong float64
		var productPrice float64
		var orderItemQuantity int
		var merchantCreatedAt, orderItemCreatedAt time.Time

		err := rows.Scan(&orderID, &merchantID, &merchantName,
			&merchantCategory, &merchantImageURL, &merchantLat,
			&merchantLong, &merchantCreatedAt, &productID,
			&productName, &productCategory, &productPrice,
			&productImageURL, &orderItemQuantity, &orderItemCreatedAt)

		if err != nil {
			panic(err)
		}

		if _, exist := merchants[orderID][merchantID]; exist == false {
			if _, order := merchants[orderID]; order == false {
				merchants[orderID] = make(map[string]*entity.OrderDetail)
			}

			merchants[orderID][merchantID] = &entity.OrderDetail{
				Merchant: entity.Merchant{
					Id:       merchantID,
					Name:     merchantName,
					Category: merchantCategory,
					ImageUrl: merchantImageURL,
					Location: &entity.Coordinate{
						Lat:  merchantLat,
						Long: merchantLong,
					},
					CreatedAt: &merchantCreatedAt,
				},
				Items: []entity.ItemHistory{},
			}
		}

		items := &merchants[orderID][merchantID].Items
		*items = append(*items, entity.ItemHistory{
			Product: entity.Product{
				Id:        productID,
				Name:      productName,
				Category:  productCategory,
				Price:     uint(productPrice),
				ImageUrl:  productImageURL,
				CreatedAt: &orderItemCreatedAt,
			},
			Quantity: orderItemQuantity,
		})
	}

	orderMap := make(map[string]*entity.OrderHistory)

	for orderId, order := range merchants {
		orderMap[orderId] = &entity.OrderHistory{
			OrderId: orderId,
			Orders:  []entity.OrderDetail{},
		}

		for _, detail := range order {
			order := orderMap[orderId]
			order.Orders = append(order.Orders, *detail)
		}
	}

	for _, orderResponse := range orderMap {
		history = append(history, *orderResponse)
	}

	return history
}

func (p *PurchaseRepo) generateQueryOrderHistory(params *entity.OrderHistoryParams) string {
	query := `
		WITH limited_orders AS (
		SELECT *
		FROM orders
		WHERE username = '` + db.Escape(params.Username) + `'
		ORDER BY created_at DESC
		LIMIT ` + strconv.Itoa(int(params.Limit)) + ` 
		OFFSET ` + strconv.Itoa(int(params.Offset)) + `
		)
	`

	query += `
		SELECT 
			lo.id as order_id,
			m.id as merchant_id,
			m.name as merchant_name,
			m.category as merchant_category,
			m.image_url as merchant_image_url,
			m.lat as merchant_location_lat,
			m.long as merchant_location_long,
			m.created_at as merchant_created_at,
			p.id as product_id,
			p.name as product_name,
			p.category as product_category,
			p.price as product_price,
			p.image_url as product_image_url,
			oi.quantity as order_item_quantity,
			oi.created_at as order_item_created_at
		FROM limited_orders lo
		JOIN order_items oi ON lo.id = oi.order_id
		JOIN products p ON oi.item_id = p.id
		JOIN merchants m ON oi.merchant_id = m.id
		WHERE TRUE
	`

	if params.MerchantId != "" {
		query += " AND m.id = '" + db.Escape(params.MerchantId) + "'"
	}

	if params.Name != "" {
		query += " AND (LOWER(m.name) LIKE '%" + db.Escape(strings.ToLower(params.Name))
		query += "%' OR LOWER(p.name) LIKE '%" + db.Escape(strings.ToLower(params.Name)) + "%')"
	}

	if params.MerchantCategory != "" {
		query += " AND m.category = '" + db.Escape(params.MerchantCategory) + "'"
	}

	return query
}
