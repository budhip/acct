package health

import (
	"net/http"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/services"

	"github.com/labstack/echo/v4"
)

type healthHandler struct {
	service services.HealthService
}

// New health handler will initialize the health/ resources endpoint
func New(app *echo.Group, hs services.HealthService) {
	hh := healthHandler{
		service: hs,
	}
	health := app.Group("/health")
	health.GET("/readiness", hh.readinessCheck)
	health.GET("/liveness", hh.livenessCheck)
}

type (
	DoReadinessCheckResponse struct {
		Kind   string `json:"kind" example:"health"`
		Status string `json:"status" example:"server is up and running"`
	}

	DoLivenessCheckResponse struct {
		Kind   string      `json:"kind" example:"health"`
		Status interface{} `json:"status" swaggertype:"object,string" example:"mysql:mysql is up and running,redis:redis is up and running"`
	}
)

// @Summary 	Get the status readiness of server
// @Description	Get the status readiness of server
// @Tags 		Health
// @Accept		json
// @Produce		json
// @Success 	200 {object} DoReadinessCheckResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Router 		/health/readiness [get]
func (hh healthHandler) readinessCheck(c echo.Context) error {
	return commonhttp.RestSuccessResponse(c, http.StatusOK, DoReadinessCheckResponse{
		Kind:   "health",
		Status: "server is up and running",
	})
}

// @Summary 	Get the status liveness of server
// @Description	Get the status liveness of server
// @Tags 		Health
// @Accept		json
// @Produce		json
// @Success 	200 {object} DoLivenessCheckResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Router 		/health/liveness [get]
func (hh healthHandler) livenessCheck(c echo.Context) error {
	return commonhttp.RestSuccessResponse(c, http.StatusOK, DoLivenessCheckResponse{
		Kind:   "health",
		Status: hh.service.GetHealth(c.Request().Context()),
	})
}
