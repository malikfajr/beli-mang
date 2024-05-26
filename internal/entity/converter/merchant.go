package converter

import "github.com/malikfajr/beli-mang/internal/entity"

type MerchantResponse struct {
	Data *[]entity.Merchant `json:"data"`
	Meta *Meta              `json:"meta"`
}
