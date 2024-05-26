package entity

import "time"

type Coordinate struct {
	Lat  float64 `json:"lat" validate:"required"`
	Long float64 `json:"long" validate:"required"`
}

type Merchant struct {
	Id        string      `json:"merchantId"`
	Username  string      `json:"-"`
	Name      string      `json:"name"`
	Category  string      `json:"merchantCategory"`
	ImageUrl  string      `json:"imageUrl"`
	Location  *Coordinate `json:"location"`
	CreatedAt *time.Time  `json:"createdAt"`
}

type AddMerchantPayload struct {
	Name     string      `json:"name" validate:"required,min=2,max=30"`
	Category string      `json:"merchantCategory" validate:"required,oneof=SmallRestaurant	MediumRestaurant LargeRestaurant MerchandiseRestaurant BoothKiosk ConvenienceStore"`
	ImageUrl string      `json:"imageUrl" validate:"required"`
	Location *Coordinate `json:"location" validate:"required"`
}

type MerchantParams struct {
	Limit      uint   `query:"limit"`
	Offset     uint   `query:"offset"`
	MerchantId string `query:"merchantId"`
	Name       string `query:"name"`
	Category   string `query:"merchantCategory"`
	CreatedAt  string `query:"createdAt"`
}

// Merchant Item
type MerchantItem struct {
	Id         string     `json:"itemId"`
	MerchantId string     `json:"-"`
	Name       string     `json:"name"`
	Category   string     `json:"productCategory"`
	Price      uint       `json:"price"`
	ImageUrl   string     `json:"imageUrl"`
	CreatedAt  *time.Time `json:"createdAt"`
}

type AddMerchantItemPayload struct {
	Id string `path:"merchantId"`
	// TODO: Fix the items payload
}
