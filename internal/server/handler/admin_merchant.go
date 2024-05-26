package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/malikfajr/beli-mang/internal/entity"
	"github.com/malikfajr/beli-mang/internal/entity/converter"
	"github.com/malikfajr/beli-mang/internal/exception"
	"github.com/malikfajr/beli-mang/internal/pkg/token"
	"github.com/malikfajr/beli-mang/internal/usecase"
)

type merchantHandler struct {
	pool *pgxpool.Pool
}

func NewMerchantHandler(pool *pgxpool.Pool) *merchantHandler {
	return &merchantHandler{
		pool: pool,
	}
}

func (m *merchantHandler) Create(c echo.Context) error {
	payload := &entity.AddMerchantPayload{}

	if err := c.Bind(payload); err != nil {
		return c.JSON(http.StatusBadRequest, exception.BadRequest("request doesn't pass validation"))
	}

	if err := c.Validate(payload); err != nil {
		return c.JSON(http.StatusBadRequest, exception.BadRequest("request doesn't pass validation"))
	}

	user := c.Get("user").(*token.JwtClaim)

	manageMerchant := usecase.NewManageMerchant(m.pool)
	merchant, err := manageMerchant.Create(c.Request().Context(), user.Username, payload)
	if err != nil {
		ex, ok := err.(*exception.CustomError)
		if ok {
			return c.JSON(ex.StatusCode, ex)
		}
		panic(err)
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"merchantId": merchant.Id,
	})
}

func (m *merchantHandler) GetAll(c echo.Context) error {
	user := c.Get("user").(*token.JwtClaim)
	params := &entity.MerchantParams{}

	c.Bind(params)

	manageMerchant := usecase.NewManageMerchant(m.pool)
	merchants, err := manageMerchant.GetAll(c.Request().Context(), user.Username, params)
	if err != nil {
		ex, ok := err.(*exception.CustomError)
		if ok {
			return c.JSON(ex.StatusCode, ex)
		}
		panic(err)
	}

	meta := &converter.Meta{
		Limit:  params.Limit,
		Offset: params.Offset,
		Total:  len(*merchants),
	}

	response := &converter.MerchantResponse{
		Data: merchants,
		Meta: meta,
	}

	return c.JSON(http.StatusOK, response)
}
