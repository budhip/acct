package httpv1

import (
	"bitbucket.org/Amartha/go-accounting/internal/contract"

	v1account "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/v1/account"
	v1accounting "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/v1/accounting"
	v1cache "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/v1/cache"
	v1category "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/v1/category"
	v1coatype "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/v1/coatype"
	v1entity "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/v1/entity"
	v1loanpartneraccount "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/v1/loanpartneraccount"
	v1migration "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/v1/migration"
	v1producttype "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/v1/producttype"
	v1publisher "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/v1/publisher"
	v1subcategory "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/v1/sub_category"

	"github.com/labstack/echo/v4"
)

// v1Group register api
func Route(g *echo.Group, c *contract.Contract) {
	v1account.New(g, c.Service.Account)
	v1accounting.New(g, c.Service.Accounting, c.Service.Journal, c.Service.TrialBalance)
	v1cache.New(g, c.Service.Account)
	v1category.New(g, c.Service.Category)
	v1coatype.New(g, c.Service.COAType)
	v1entity.New(g, c.Service.Entity)
	v1loanpartneraccount.New(g, c.Service.LoanPartnerAccount)
	v1migration.New(g, c.Service.Migration)
	v1producttype.New(g, c.Service.ProductType)
	v1publisher.New(g, c.Service.PublisherService)
	v1subcategory.New(g, c.Service.SubCategory)
}
