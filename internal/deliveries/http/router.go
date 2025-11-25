package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/contract"
	"bitbucket.org/Amartha/go-accounting/internal/deliveries/http/health"
	"bitbucket.org/Amartha/go-accounting/internal/deliveries/http/middleware"
	"bitbucket.org/Amartha/go-accounting/internal/deliveries/http/pprof"
	httpv1 "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/v1"
	httpv2 "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/v2"

	"bitbucket.org/Amartha/go-accounting/internal/pkg/graceful"
	"bitbucket.org/Amartha/go-x/environment"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"

	// for swagger docs
	_ "bitbucket.org/Amartha/go-accounting/docs"
)

type ProcessStartStopper interface {
	Start(ctx context.Context) (graceful.ProcessStarter, graceful.ProcessStopper)
}

type svc struct {
	e               *echo.Echo
	addr            string
	gracefulTimeout time.Duration
}

var _ ProcessStartStopper = (*svc)(nil)

func (s *svc) Start(ctx context.Context) (graceful.ProcessStarter, graceful.ProcessStopper) {
	return func() error {
			if err := s.e.Start(s.addr); err != http.ErrServerClosed {
				return err
			}
			return nil
		},
		func(ctx context.Context) error {
			return s.e.Shutdown(ctx)
		}
}

// @title GO ACCOUNTING API DUCUMENTATION
// @version 1.0
// @description This is a go accounting api docs.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host go-accounting-dev.amartha.id
// @BasePath /api
// @schemes https
func NewHTTPServer(
	ctx context.Context,
	contract *contract.Contract,
) *svc {
	e := echo.New()
	svc := &svc{e: e, addr: fmt.Sprintf(":%d", contract.Config.App.HTTPPort), gracefulTimeout: contract.Config.App.GracefulTimeout}

	// options middleware
	m := middleware.New(e, contract)
	m.Default()

	e.GET("/", func(c echo.Context) error {
		message := fmt.Sprintf("Welcome to %s", contract.Config.App.Name)
		return c.String(http.StatusOK, message)
	})

	e.GET("/metrics", echoprometheus.NewHandler())

	if environment.ToEnvironment(contract.Config.App.Env) == environment.LOCAL_ENV {
		// swagger
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	// api group
	gAPI := e.Group("/api")

	// health check
	health.New(gAPI, contract.Service.HealthService)

	// pprof
	pprof.New(e)

	// v1 group
	gV1 := gAPI.Group("/v1")
	// v1Group middleware
	gV1.Use(m.InternalAuth())
	httpv1.Route(gV1, contract)

	// v2 group
	gV2 := gAPI.Group("/v2")
	gV2.Use(m.InternalAuth())
	httpv2.Route(gV2, contract)

	return svc
}
