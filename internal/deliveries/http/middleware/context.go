package middleware

import (
	"bitbucket.org/Amartha/go-x/log/ctxdata"
	"context"

	"github.com/labstack/echo/v4"
)

func (m *AppMiddleware) setContext() func(echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctxWithTimeout, cancel := context.WithTimeout(c.Request().Context(), m.c.Config.App.HTTPTimeout)
			defer cancel()
			c.SetRequest(c.Request().WithContext(ctxdata.SetContextFromHTTP(ctxWithTimeout, c.Request(), m.c.Config.GcloudProjectID)))
			return next(c)
		}
	}
}
