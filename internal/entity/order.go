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
	OrderId string        `json:"orderId"`
	Orders  []OrderDetail `json:"orders"`
}

type ItemHistory struct {
	Product
	Quantity int `json:"quantity"`
}

type OrderDetail struct {
	Merchant Merchant      `json:"merchant"`
	Items    []ItemHistory `json:"items"`
}

type OrderHistoryParams struct {
	Limit            uint   `json:"-" query:"limit"`
	Offset           uint   `json:"-" query:"offset"`
	MerchantId       string `json:"-" query:"merchantId"`
	MerchantCategory string `json:"-" query:"merchantCategory"`
	Name             string `json:"-" query:"name"`
	Username         string
}
