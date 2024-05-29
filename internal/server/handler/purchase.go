package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/malikfajr/beli-mang/internal/entity/converter"
	"github.com/malikfajr/beli-mang/internal/exception"
	"github.com/malikfajr/beli-mang/internal/usecase"
)

type purchaseHandler struct {
	pool  *pgxpool.Pool
	pcase usecase.PurchaseCase
}

type PurchaseHandler interface {
	GetMerchantNearby(c echo.Context) error
}

func NewPurchasehandler(pool *pgxpool.Pool) PurchaseHandler {
	return &purchaseHandler{
		pool:  pool,
		pcase: usecase.NewPurchaseCase(pool),
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
