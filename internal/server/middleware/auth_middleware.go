package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/malikfajr/beli-mang/internal/exception"
	jwt "github.com/malikfajr/beli-mang/internal/pkg/token"
)

func Auth(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			Authorization := c.Request().Header.Get("Authorization")

			if len(Authorization) < 9 || Authorization[:7] != "Bearer " {
				return c.JSON(http.StatusUnauthorized, exception.Unauthorized("Invalid token"))
			}

			token := Authorization[7:]
			claim, err := jwt.ClaimToken(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, exception.Unauthorized("Invalid token"))
			}

			c.Set("user", claim)

			// check the admin role
			if role == "admin" && claim.Admin {
				return next(c)
			}

			// default user passing middleware if token is valid
			return next(c)
		}
	}
}
