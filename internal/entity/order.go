package entity

type OrderPayload struct {
	UserLocation Coordinate `json:"userLocation" validate:"required"`
	Orders       []Order    `json:"orders" validate:"required,dive"`
}

type Order struct {
	MerchantId    string `json:"merchantId" validate:"required"`
	StartingPoint bool   `json:"isStartingPoint"`
	Items         []Item `json:"items" validate:"required,dive"`
}

type Item struct {
	ItemId   string `json:"itemId" validate:"required"`
	Quantity uint   `json:"quantity" validate:"required,min=1"`
}

type EstimateResponse struct {
	TotalPrice                     int    `json:"totalPrice"`
	EstimatedDeliveryTimeInMinutes int    `json:"estimatedDeliveryTimeInMinutes"`
	CalculatedEstimateId           string `json:"calculatedEstimateId"`
}

type OrderResponse struct {
	OrderId string `json:"orderId"`
}

type OrderHistory struct {
	OrderId string `json:"orderId"`
}
