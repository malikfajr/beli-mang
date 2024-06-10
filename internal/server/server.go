package server

import (
	"context"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/malikfajr/beli-mang/internal/driver/db"
	"github.com/malikfajr/beli-mang/internal/pkg/customvalidator"
	"github.com/malikfajr/beli-mang/internal/server/routes"
)

func Run() {
	e := echo.New()

	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Use(middleware.CORS())

	e.Validator = customvalidator.NewCustomValidator(validator.New())

	dbAddress := db.Address()
	pool := db.NewPool(context.Background(), dbAddress)
	defer pool.Close()

	pool.Ping(context.Background())

	routes.NewRoutes(e, pool)

	if err := e.Start(":8080"); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
