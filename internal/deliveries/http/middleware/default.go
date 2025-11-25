package middleware

import (
	"bitbucket.org/Amartha/go-accounting/internal/pkg/metrics"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4/middleware"
	"github.com/newrelic/go-agent/v3/integrations/nrecho-v4"
)

func (m *AppMiddleware) Default() {
	m.e.Use(middleware.Recover())
	m.e.Use(m.setContext())
	m.e.Use(m.timeout())
	if m.c.NewRelic != nil {
		m.e.Use(nrecho.Middleware(m.c.NewRelic))
	}
	m.e.Use(m.logger())
	m.e.Use(middleware.BodyLimit("201M"))
	m.e.Use(echoprometheus.NewMiddleware(metrics.BuildFQName(m.c.Config.App.Name))) // adds middleware to gather metrics
}
