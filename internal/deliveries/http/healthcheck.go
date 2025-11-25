package http

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-accounting/internal/contract"
	"bitbucket.org/Amartha/go-accounting/internal/deliveries/http/health"
	"bitbucket.org/Amartha/go-accounting/internal/deliveries/http/pprof"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/graceful"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type httpHealthCheck struct {
	e *echo.Echo
	c *contract.Contract
}

func NewHealthCheck(c *contract.Contract) ProcessStartStopper {
	return &httpHealthCheck{
		e: echo.New(),
		c: c,
	}
}

func (h *httpHealthCheck) Start(ctx context.Context) (graceful.ProcessStarter, graceful.ProcessStopper) {

	h.e.Use(middleware.Recover())
	h.e.Use(middleware.Logger())

	// api group
	gAPI := h.e.Group("/api")

	// pprof
	pprof.New(h.e)

	// health check
	health.New(gAPI, h.c.Service.HealthService)

	// metrics
	h.e.GET("/metrics", echoprometheus.NewHandler())

	return func() error {
			return h.e.Start(fmt.Sprintf(":%d", h.c.Config.Kafka.HealthCheckPort))
		},
		func(ctx context.Context) error {
			return h.e.Shutdown(ctx)
		}
}
