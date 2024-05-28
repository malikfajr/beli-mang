package routes

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/malikfajr/beli-mang/internal/server/handler"
	"github.com/malikfajr/beli-mang/internal/server/middleware"
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

	merchantHandler := handler.NewMerchantHandler(pool)

	adminMerchant := e.Group("/admin/merchants", middleware.Auth("admin"))
	adminMerchant.POST("", merchantHandler.Create)
	adminMerchant.GET("", merchantHandler.GetAll)
	adminMerchant.POST("/:merchantId/items", merchantHandler.AddProduct)
	adminMerchant.GET("/:merchantId/items", merchantHandler.GetProducts)

	merchantHandler.ResetCache(3 * time.Minute)

	imageHandler := &handler.ImageHandler{}
	e.POST("/image", imageHandler.Upload, middleware.Auth("admin"))
}
