package converter

import "github.com/malikfajr/beli-mang/internal/entity"

type ProductResponse struct {
	Data *[]entity.Product `json:"data"`
	Meta *Meta `json:"meta"`
}