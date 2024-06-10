package handler

import (
	"context"
	"errors"
	"log"
	"math"
	"net/http"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/malikfajr/beli-mang/internal/entity"
	"github.com/malikfajr/beli-mang/internal/entity/converter"
	"github.com/malikfajr/beli-mang/internal/exception"
	"github.com/malikfajr/beli-mang/internal/pkg/token"
	"github.com/malikfajr/beli-mang/internal/usecase"
	"github.com/oklog/ulid/v2"
)

const (
	EarthRadiusKm         = 6371        // Radius bumi dalam kilometer
	DeliverySpeedKmPerMin = 40.0 / 60.0 // Kecepatan pengiriman dalam kilometer per menit (40 km/jam)
)

type CacheEstimate struct {
	MerchantId string
	ProductId  string
	Qty        int
	Price      int
}

type purchaseHandler struct {
	pool     *pgxpool.Pool
	pcase    usecase.PurchaseCase
	estimate map[string][]CacheEstimate
	sync.Mutex
}

type PurchaseHandler interface {
	GetMerchantNearby(c echo.Context) error
	CreateEstimate(c echo.Context) error
	PostOrder(c echo.Context) error
	GetHistory(c echo.Context) error
}

func NewPurchasehandler(pool *pgxpool.Pool) PurchaseHandler {
	return &purchaseHandler{
		pool:     pool,
		pcase:    usecase.NewPurchaseCase(pool),
		estimate: make(map[string][]CacheEstimate, 0),
	}
}

// GetHistory implements PurchaseHandler.
func (p *purchaseHandler) GetHistory(c echo.Context) error {
	user := c.Get("user").(*token.JwtClaim)
	params := &entity.OrderHistoryParams{}

	c.Bind(params)

	params.Username = user.Username

	history := p.pcase.GetHistory(c.Request().Context(), params)

	return c.JSON(http.StatusOK, history)

}

func (p *purchaseHandler) GetMerchantNearby(c echo.Context) error {
	params := &converter.MerchanNearbyParams{}

	c.Bind(params)

	data, total, err := p.pcase.GetMerchantNearby(c.Request().Context(), params)
	if err != nil {
		ex, ok := err.(*exception.CustomError)
		if ok {
			return c.JSON(ex.StatusCode, ex)
		}
		panic(err)
	}

	return c.JSON(http.StatusOK, &converter.MerchanNearbyResponse{
		Data: data,
		Meta: &converter.Meta{
			Limit:  params.Limit,
			Offset: params.Offset,
			Total:  total,
		},
	})
}

func (p *purchaseHandler) CreateEstimate(c echo.Context) error {
	var payload entity.OrderPayload

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, exception.BadRequest("request doesn't pass validation"))
	}

	if err := c.Validate(payload); err != nil {
		return c.JSON(http.StatusBadRequest, exception.BadRequest("request doesn't pass validation"))
	}

	if err, statusCode := p.validatePayloadOrder(&payload); err != nil {
		return c.JSON(statusCode, exception.CustomError{
			Message:    err.Error(),
			StatusCode: statusCode,
		})
	}

	// Retrieve merchant locations
	merchants := make(map[string]entity.Coordinate)
	for _, order := range payload.Orders {
		var location entity.Coordinate
		err := p.pool.QueryRow(context.Background(), "SELECT lat, long FROM merchants WHERE id = $1", order.MerchantId).Scan(&location.Lat, &location.Long)
		if err != nil {
			return c.JSON(http.StatusNotFound, exception.NotFound("Merchant id not found"))
		}
		merchants[order.MerchantId] = location
	}

	// Calculate total price
	totalPrice, err := p.calculateTotalPrice(payload.Orders)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Identify starting point
	startingPointID := ""
	for _, order := range payload.Orders {
		if order.StartingPoint {
			startingPointID = order.MerchantId
			break
		}
	}

	// Calculate total travel time using TSP
	totalTravelTime := calculateTotalTravelTimeTSP(payload.UserLocation, merchants, startingPointID)

	// Save calculation to database
	calculationID := ulid.Make().String()

	go p.SaveEstimate(calculationID, payload)

	return c.JSON(http.StatusOK, entity.EstimateResponse{
		TotalPrice:                     int(totalPrice),
		EstimatedDeliveryTimeInMinutes: int(totalTravelTime),
		CalculatedEstimateId:           calculationID,
	})
}

func (p *purchaseHandler) validatePayloadOrder(payload *entity.OrderPayload) (error, int) {
	// TODO: Optimize validation -> using any / in for select id
	userLat := payload.UserLocation.Lat
	userLong := payload.UserLocation.Long

	startingPoints := 0
	for _, order := range payload.Orders {
		if order.StartingPoint {
			startingPoints++
		}

		if _, err := ulid.Parse(order.MerchantId); err != nil {
			return errors.New("merchant id not found"), http.StatusNotFound
		}

		var merchantId string
		var lat, long float64
		err := p.pool.QueryRow(context.Background(), `SELECT id, lat, long FROM merchants WHERE id = $1`, order.MerchantId).Scan(&merchantId, &lat, &long)
		if err != nil {
			return errors.New("merchant id not found"), http.StatusNotFound
		}

		distance := haversine(userLat, userLong, lat, long)
		if order.StartingPoint && distance > 3 {
			return errors.New("Merchant " + merchantId + " too far"), http.StatusBadRequest
		}

		// if len(order.Items) < 1 {
		// 	return errors.New("item min 1 in merchant id: " + merchantId), http.StatusBadRequest
		// }

		for _, item := range order.Items {
			_, err := ulid.Parse(item.ItemId)
			if err != nil {
				return errors.New("item id not found"), http.StatusNotFound
			}

			var itemId string

			err = p.pool.QueryRow(context.Background(), `SELECT id FROM products WHERE id = $1 AND merchant_id = $2`, item.ItemId, merchantId).Scan(&itemId)
			if err != nil {
				return errors.New("item id " + item.ItemId + " not found"), http.StatusNotFound
			}
		}
	}
	if startingPoints != 1 {
		return errors.New("invalid payload: exactly one order must have isStartingPoint set to true"), http.StatusBadRequest
	}
	return nil, 0
}

func (p *purchaseHandler) calculateTotalPrice(orders []entity.Order) (float64, error) {
	var totalPrice float64
	for _, order := range orders {
		for _, item := range order.Items {
			var price float64
			err := p.pool.QueryRow(context.Background(), "SELECT price FROM products WHERE id = $1", item.ItemId).Scan(&price)
			if err != nil {
				return 0, errors.New("item with ID " + item.ItemId + " not found")
			}

			totalPrice += price * float64(item.Quantity)
		}
	}
	return totalPrice, nil
}

func calculateTotalTravelTime(payload entity.OrderPayload, merchants map[string]entity.Coordinate) float64 {
	var totalDistance float64
	userLocation := payload.UserLocation
	startingPointID := ""

	// Identify starting point
	for _, order := range payload.Orders {
		if order.StartingPoint {
			startingPointID = order.MerchantId
			break
		}
	}

	// Calculate distances between merchants
	lastLocation := merchants[startingPointID]
	for _, order := range payload.Orders {
		if order.MerchantId != startingPointID {
			totalDistance += haversine(lastLocation.Lat, lastLocation.Long, merchants[order.MerchantId].Lat, merchants[order.MerchantId].Long)
			lastLocation = merchants[order.MerchantId]
		}
	}

	totalDistance += haversine(lastLocation.Lat, lastLocation.Long, userLocation.Lat, userLocation.Long)

	// Convert distance to time (in minutes)
	return totalDistance / DeliverySpeedKmPerMin
}

// Calculate total travel time using Greedy Nearest Neighbor
func calculateTotalTravelTimeTSP(userLocation entity.Coordinate, merchants map[string]entity.Coordinate, startingPointID string) float64 {
	merchantIDs := []string{}
	for id := range merchants {
		if id != startingPointID {
			merchantIDs = append(merchantIDs, id)
		}
	}

	if len(merchantIDs) == 0 {
		distance := haversine(userLocation.Lat, userLocation.Long, merchants[startingPointID].Lat, merchants[startingPointID].Long)
		distance += haversine(merchants[startingPointID].Lat, merchants[startingPointID].Long, userLocation.Lat, userLocation.Long) // Distance back to user
		log.Printf("Only one merchant, total distance: %.2f km", distance)
		return distance / DeliverySpeedKmPerMin
	}

	visited := make(map[string]bool)
	currentLocation := merchants[startingPointID]
	totalDistance := 0.0

	for i := 0; i < len(merchantIDs); i++ {
		nearestMerchantID := ""
		minDistance := math.MaxFloat64

		for _, id := range merchantIDs {
			if visited[id] {
				continue
			}

			distance := haversine(currentLocation.Lat, currentLocation.Long, merchants[id].Lat, merchants[id].Long)
			if distance < minDistance {
				minDistance = distance
				nearestMerchantID = id
			}
		}

		if nearestMerchantID != "" {
			visited[nearestMerchantID] = true
			totalDistance += minDistance
			currentLocation = merchants[nearestMerchantID]
		}
	}

	// Travel to the user's location after visiting all merchants
	totalDistance += haversine(currentLocation.Lat, currentLocation.Long, userLocation.Lat, userLocation.Long)

	return totalDistance / DeliverySpeedKmPerMin
}

// Calculate Haversine distance between two points
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return EarthRadiusKm * c
}

func toRadians(degree float64) float64 {
	return degree * math.Pi / 180
}

func (p *purchaseHandler) SaveEstimate(estimateId string, payload entity.OrderPayload) {
	p.Lock()
	defer p.Unlock()

	cacheEntimate := []CacheEstimate{}

	for _, order := range payload.Orders {
		for _, item := range order.Items {
			temp := CacheEstimate{
				MerchantId: order.MerchantId,
				ProductId:  item.ItemId,
				Qty:        int(item.Quantity),
			}

			cacheEntimate = append(cacheEntimate, temp)
		}
	}

	p.estimate[estimateId] = cacheEntimate
}

// PostOrder implements PurchaseHandler.
func (p *purchaseHandler) PostOrder(c echo.Context) error {
	var payload struct {
		CalculatedEstimateId string `json:"calculatedEstimateId" validate:"required"`
	}

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, exception.BadRequest("request doesn't pass validation"))
	}

	if err := c.Validate(payload); err != nil {
		return c.JSON(http.StatusBadRequest, exception.BadRequest("request doesnt't pass validation"))
	}

	p.Lock()
	_, ok := p.estimate[payload.CalculatedEstimateId]
	p.Unlock()

	if !ok {
		return c.JSON(http.StatusNotFound, exception.NotFound("calculatedEstimateId is not found"))
	}

	user := c.Get("user").(*token.JwtClaim)

	orderId := ulid.Make().String()

	// go p.saveOrder(user.Username, orderId, payload.CalculatedEstimateId)
	p.saveOrder(user.Username, orderId, payload.CalculatedEstimateId)

	return c.JSON(http.StatusCreated, entity.OrderResponse{
		OrderId: orderId,
	})
}

func (p *purchaseHandler) saveOrder(username string, orderId string, estimateId string) {
	p.Lock()
	defer p.Unlock()

	query1 := "INSERT INTO orders(id, username) VALUES($1, $2)"
	_, err := p.pool.Exec(context.Background(), query1, orderId, username)
	if err != nil {
		panic(err)
	}

	query2 := "INSERT INTO order_items(order_id, merchant_id, item_id, quantity) VALUES($1, $2, $3, $4)"

	cacheEtimate := p.estimate[estimateId]

	for _, item := range cacheEtimate {
		_, err := p.pool.Exec(context.Background(), query2, orderId, item.MerchantId, item.ProductId,
			item.Qty)
		if err != nil {
			panic(err)
		}
	}

	delete(p.estimate, estimateId)
}
