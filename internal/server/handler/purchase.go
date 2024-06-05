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
}

func NewPurchasehandler(pool *pgxpool.Pool) PurchaseHandler {
	return &purchaseHandler{
		pool:     pool,
		pcase:    usecase.NewPurchaseCase(pool),
		estimate: make(map[string][]CacheEstimate, 0),
	}
}

func (p *purchaseHandler) GetMerchantNearby(c echo.Context) error {
	params := &converter.MerchanNearbyParams{}

	c.Bind(params)

	data, err := p.pcase.GetMerchantNearby(c.Request().Context(), params)
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
			Total:  len(*data),
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

	if err := p.validatePayloadOrder(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, exception.BadRequest(err.Error()))
	}

	// Retrieve merchant locations
	merchants := make(map[string]entity.Coordinate)
	for _, order := range payload.Orders {
		var location entity.Coordinate
		err := p.pool.QueryRow(context.Background(), "SELECT lat, long FROM merchants WHERE id = $1", order.MerchantId).Scan(&location.Lat, &location.Long)
		if err != nil {
			return c.JSON(http.StatusBadRequest, exception.BadRequest("Merchant id not found"))
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
	// _, err = p.pool.Exec(context.Background(), "INSERT INTO calculations (id, user_lat, user_long, total_price, estimated_delivery_time_in_minutes) VALUES ($1, $2, $3, $4, $5)",
	// 	calculationID, payload.UserLocation.Lat, payload.UserLocation.Long, totalPrice, totalTravelTime)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save calculation"})
	// }

	// query := `
	// 	SELECT
	// `
	go p.SaveEstimate(calculationID, payload)

	return c.JSON(http.StatusOK, entity.EstimateResponse{
		TotalPrice:                     int(totalPrice),
		EstimatedDeliveryTimeInMinutes: int(totalTravelTime),
		CalculatedEstimateId:           calculationID,
	})
}

func (p *purchaseHandler) validatePayloadOrder(payload *entity.OrderPayload) error {
	// TODO: Optimize validation -> using any / in for select id
	startingPoints := 0
	for _, order := range payload.Orders {
		if order.StartingPoint {
			startingPoints++
		}

		if _, err := ulid.Parse(order.MerchantId); err != nil {
			return errors.New("merchant id not found")
		}

		var merchantId string
		err := p.pool.QueryRow(context.Background(), `SELECT id FROM merchants WHERE id = $1`, order.MerchantId).Scan(&merchantId)
		if err != nil {
			return errors.New("merchant id not found")
		}

		if len(order.Items) < 1 {
			return errors.New("item min 1 in merchant id: " + merchantId)
		}

		for _, item := range order.Items {
			_, err := ulid.Parse(item.ItemId)
			if err != nil {
				return errors.New("item id not found")
			}

			var itemId string

			err = p.pool.QueryRow(context.Background(), `SELECT id FROM products WHERE id = $1 AND merchant_id = $2`, item.ItemId, merchantId).Scan(&itemId)
			if err != nil {
				return errors.New("item id " + item.ItemId + " not found")
			}
		}
	}
	if startingPoints != 1 {
		return errors.New("invalid payload: exactly one order must have isStartingPoint set to true")
	}
	return nil
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

	// Calculate distance from user to starting point
	totalDistance += haversine(userLocation.Lat, userLocation.Long, merchants[startingPointID].Lat, merchants[startingPointID].Long)

	// Calculate distances between merchants
	lastLocation := merchants[startingPointID]
	for _, order := range payload.Orders {
		if order.MerchantId != startingPointID {
			totalDistance += haversine(lastLocation.Lat, lastLocation.Long, merchants[order.MerchantId].Lat, merchants[order.MerchantId].Long)
			lastLocation = merchants[order.MerchantId]
		}
	}

	// Convert distance to time (in minutes)
	return totalDistance / DeliverySpeedKmPerMin
}

// Permutations generates all permutations of a slice of strings
func permutations(arr []string) [][]string {
	var helper func([]string, int)
	res := [][]string{}

	helper = func(arr []string, n int) {
		if n == 1 {
			tmp := make([]string, len(arr))
			copy(tmp, arr)
			res = append(res, tmp)
		} else {
			for i := 0; i < n; i++ {
				helper(arr, n-1)
				if n%2 == 1 {
					arr[0], arr[n-1] = arr[n-1], arr[0]
				} else {
					arr[i], arr[n-1] = arr[n-1], arr[i]
				}
			}
		}
	}

	helper(arr, len(arr))
	return res
}

// Calculate total travel time using TSP
func calculateTotalTravelTimeTSP(userLocation entity.Coordinate, merchants map[string]entity.Coordinate, startingPointID string) float64 {
	// Create a slice of merchant IDs excluding the starting point
	merchantIDs := []string{}
	for id := range merchants {
		if id != startingPointID {
			merchantIDs = append(merchantIDs, id)
		}
	}

	// Handle the case with only one merchant
	if len(merchantIDs) == 0 {
		distance := haversine(userLocation.Lat, userLocation.Long, merchants[startingPointID].Lat, merchants[startingPointID].Long)
		distance += haversine(merchants[startingPointID].Lat, merchants[startingPointID].Long, userLocation.Lat, userLocation.Long) // Distance back to user
		log.Printf("Only one merchant, total distance: %.2f km", distance)
		return distance / DeliverySpeedKmPerMin
	}

	// Generate all permutations of the merchant IDs
	perms := permutations(merchantIDs)
	log.Println(perms, merchantIDs, merchants, startingPointID)

	// Calculate the minimum travel distance
	minDistance := math.MaxFloat64
	for _, perm := range perms {
		distance := 0.0
		currentLocation := merchants[startingPointID]
		for _, id := range perm {
			distance += haversine(currentLocation.Lat, currentLocation.Long, merchants[id].Lat, merchants[id].Long)
			currentLocation = merchants[id]
		}
		distance += haversine(userLocation.Lat, userLocation.Long, merchants[startingPointID].Lat, merchants[startingPointID].Long) // Distance back to user
		if distance < minDistance {
			minDistance = distance
		}
	}

	// Convert distance to time (in minutes)
	return minDistance / DeliverySpeedKmPerMin
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
		CalculatedEstimateId string `json:"calculatedEstimateId"`
	}

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, exception.BadRequest("request doesn't pass validation"))
	}

	_, ok := p.estimate[payload.CalculatedEstimateId]
	if !ok {
		return c.JSON(http.StatusNotFound, exception.NotFound("calculatedEstimateId is not found"))
	}

	user := c.Get("user").(*token.JwtClaim)

	orderId := ulid.Make().String()

	go p.saveOrder(user.Username, orderId, payload.CalculatedEstimateId)

	return c.JSON(http.StatusCreated, entity.OrderResponse{
		OrderId: orderId,
	})
}

func (p *purchaseHandler) saveOrder(username string, orderId string, estimateId string) {
	tx, err := p.pool.Begin(context.Background())
	if err != nil {
		panic(err)
	}
	defer func() {
		err := recover().(error)
		if err != nil {
			tx.Rollback(context.Background())
		} else {
			tx.Commit(context.Background())
		}
	}()

	query1 := "INSERT INTO orders(id, username) VALUES($1, $2)"
	_, err = tx.Exec(context.Background(), query1, orderId, username)
	if err != nil {
		panic(err)
	}

	query2 := "INSERT INTO order_items(order_id, merchant_id, item_id, quantity) VALUES($1, $2, $3, $4)"

	cacheEtimate := p.estimate[estimateId]

	for _, item := range cacheEtimate {
		_, err := tx.Exec(context.Background(), query2, orderId, item.MerchantId, item.ProductId,
			item.Qty)
		if err != nil {
			panic(err)
		}
	}
}
