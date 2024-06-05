package entity

import "time"

type Product struct {
	Id         string     `json:"itemId"`
	MerchantId string     `json:"-"`
	Name       string     `json:"name"`
	Category   string     `json:"productCategory"`
	Price      uint       `json:"price"`
	ImageUrl   string     `json:"imageUrl"`
	CreatedAt  *time.Time `json:"createdAt"`
}

type AddProductPayload struct {
	Name     string `json:"name" validate:"required,min=2,max=30"`
	Category string `json:"productCategory" validate:"required,oneof=Beverage Food Snack Condiments Additions"`
	Price    uint   `json:"price" validate:"required"`
	ImageUrl string `json:"imageUrl" validate:"required,imageUrl"`
}

type ProductResponse struct {
	Id string `json:"itemId"`
}

type ProductParams struct {
	Limit      uint   `query:"limit"`
	Offset     uint   `query:"offset"`
	Id         string `query:"itemId"`
	MerchantId string `param:"merchantId"`
	Name       string `query:"name"`
	Category   string `query:"productCategory"`
	CreatedAt  string `query:"createdAt"`
}
