package middleware

import (
	"fmt"
	"net/http"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"github.com/labstack/echo/v4"
)

func (m *AppMiddleware) InternalAuth() func(echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			secretKey := c.Request().Header.Get("X-Secret-Key")
			statusCode := http.StatusUnauthorized
			if secretKey == "" {
				return commonhttp.RestErrorResponse(c, statusCode, fmt.Errorf("%s", "required secret key"))
			}

			if secretKey != m.c.Config.SecretKey {
				return commonhttp.RestErrorResponse(c, statusCode, fmt.Errorf("%s", "invalid secret key"))
			}

			return next(c)
		}
	}
}
