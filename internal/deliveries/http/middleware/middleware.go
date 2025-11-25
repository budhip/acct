package middleware

import (
	"bitbucket.org/Amartha/go-accounting/internal/contract"

	"github.com/labstack/echo/v4"
)

type AppMiddleware struct {
	e *echo.Echo
	c *contract.Contract
}

func New(e *echo.Echo, c *contract.Contract) AppMiddleware {
	return AppMiddleware{
		e: e,
		c: c,
	}
}
