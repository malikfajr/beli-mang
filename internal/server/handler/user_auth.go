package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/malikfajr/beli-mang/internal/entity"
	"github.com/malikfajr/beli-mang/internal/exception"
	"github.com/malikfajr/beli-mang/internal/pkg/token"
	"github.com/malikfajr/beli-mang/internal/usecase"
)

type userHandler struct {
	pool *pgxpool.Pool
}

func NewUserHanlder(pool *pgxpool.Pool) *userHandler {
	return &userHandler{
		pool: pool,
	}
}

func (a *userHandler) Register(c echo.Context) error {
	payload := &entity.User{}

	if err := c.Bind(payload); err != nil {
		return c.JSON(http.StatusBadRequest, exception.BadRequest("request doesn’t pass validation"))
	}

	if err := c.Validate(payload); err != nil {
		return c.JSON(http.StatusBadRequest, exception.BadRequest("request doesn’t pass validation"))

	}

	userAuth := usecase.NewUserAuth(a.pool)

	err := userAuth.Insert(c.Request().Context(), payload)
	if err != nil {
		ex, ok := err.(*exception.CustomError)
		if ok {
			return c.JSON(ex.StatusCode, ex)
		}
		panic(err)
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"token": token.CreateToken(payload.Username, true),
	})
}

func (a *userHandler) Login(c echo.Context) error {
	payload := &entity.UserLogin{}

	if err := c.Bind(payload); err != nil {
		return c.JSON(http.StatusBadRequest, exception.BadRequest("request doesn’t pass validation"))
	}

	if err := c.Validate(payload); err != nil {
		return c.JSON(http.StatusBadRequest, exception.BadRequest("request doesn’t pass validation"))
	}

	userAuth := usecase.NewUserAuth(a.pool)

	_, err := userAuth.Login(c.Request().Context(), payload)
	if err != nil {
		ex, ok := err.(*exception.CustomError)
		if ok {
			return c.JSON(ex.StatusCode, ex)
		}
		panic(err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": token.CreateToken(payload.Username, true),
	})
}
