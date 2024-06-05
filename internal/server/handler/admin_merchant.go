package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/malikfajr/beli-mang/internal/entity"
	"github.com/malikfajr/beli-mang/internal/entity/converter"
	"github.com/malikfajr/beli-mang/internal/exception"
	"github.com/malikfajr/beli-mang/internal/pkg/token"
	"github.com/malikfajr/beli-mang/internal/usecase"
)

type merchantHandler struct {
	pool           *pgxpool.Pool
	manageMerchant usecase.ManageMerchant
}

func NewMerchantHandler(pool *pgxpool.Pool) *merchantHandler {
	return &merchantHandler{
		pool:           pool,
		manageMerchant: usecase.NewManageMerchant(pool),
	}
}

func (m *merchantHandler) Create(c echo.Context) error {
	payload := &entity.AddMerchantPayload{}

	if err := c.Bind(payload); err != nil {
		return c.JSON(http.StatusBadRequest, exception.BadRequest("request doesn't pass validation"))
	}

	if err := c.Validate(payload); err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusBadRequest, exception.BadRequest("request doesn't pass validation"))
	}

	user := c.Get("user").(*token.JwtClaim)

	merchant, err := m.manageMerchant.Create(c.Request().Context(), user.Username, payload)
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

	merchants, total, err := m.manageMerchant.GetAll(c.Request().Context(), user.Username, params)
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
		Total:  total,
	}

	response := &converter.MerchantResponse{
		Data: merchants,
		Meta: meta,
	}

	return c.JSON(http.StatusOK, response)
}

func (m *merchantHandler) AddProduct(c echo.Context) error {
	payload := &entity.AddProductPayload{}
	merchantId := c.Param("merchantId")

	if err := c.Bind(payload); err != nil {
		return c.JSON(http.StatusBadRequest, exception.BadRequest("request doesn't pass validation"))
	}

	if err := c.Validate(payload); err != nil {
		return c.JSON(http.StatusBadRequest, exception.BadRequest("request doesn't pass validation"))
	}

	data, err := m.manageMerchant.AddProduct(c.Request().Context(), merchantId, payload)
	if err != nil {
		ex, ok := err.(*exception.CustomError)
		if ok {
			return c.JSON(ex.StatusCode, ex)
		}
		panic(err)
	}

	return c.JSON(http.StatusCreated, &entity.ProductResponse{
		Id: data.Id,
	})
}

func (m *merchantHandler) GetProducts(c echo.Context) error {
	user := c.Get("user").(*token.JwtClaim)
	params := &entity.ProductParams{}

	c.Bind(params)

	data, total, err := m.manageMerchant.GetProducts(c.Request().Context(), user.Username, params)
	if err != nil {
		ex, ok := err.(*exception.CustomError)
		if ok {
			return c.JSON(ex.StatusCode, ex)
		}
		panic(err)
	}

	return c.JSON(http.StatusOK, &converter.ProductResponse{
		Data: data,
		Meta: &converter.Meta{
			Limit:  params.Limit,
			Offset: params.Offset,
			Total:  total,
		},
	})
}

func (m *merchantHandler) ResetCache(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			<-ticker.C
			m.manageMerchant.ResetData()
		}
	}()
}
