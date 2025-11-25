package httpv2

import (
	"bitbucket.org/Amartha/go-accounting/internal/contract"
	v2accounting "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/v2/accounting"
	"github.com/labstack/echo/v4"
)

func Route(g *echo.Group, c *contract.Contract) {
	v2accounting.New(g, c.Service.Accounting)
}
