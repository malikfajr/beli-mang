package converter

import "github.com/malikfajr/beli-mang/internal/entity"

type MerchanNearby struct {
	Merchant entity.Merchant  `json:"merchant"`
	Items    []entity.Product `json:"items"`
}

type MerchanNearbyParams struct {
	Coordinate string `param:"coordinate"`
	MerchantId string `query:"merchantId"`
	Name       string `query:"name"`
	Category   string `query:"merchantCategory"`

	Limit  uint `query:"limit"`
	Offset uint `query:"offset"`
}

type MerchanNearbyResponse struct {
	Data *[]MerchanNearby `json:"data"`
	Meta *Meta          `json:"meta"`
}
