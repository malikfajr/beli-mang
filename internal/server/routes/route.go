package routes

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/malikfajr/beli-mang/internal/server/handler"
)

func NewRoutes(e *echo.Echo, pool *pgxpool.Pool) {
	adminHandler := handler.NewAdminHanlder(pool)

	admin := e.Group("/admin")
	admin.POST("/register", adminHandler.Register)
	admin.POST("/login", adminHandler.Login)

	userHandler := handler.NewUserHanlder(pool)
	user := e.Group("/user")
	user.POST("/register", userHandler.Register)
	user.POST("/login", userHandler.Login)
}
