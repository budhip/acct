package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/labstack/echo/v4"
)

func (m *AppMiddleware) timeout() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			// Use a goroutine to execute the handler
			done := make(chan error, 1)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						done <- fmt.Errorf("panic: %v", r)
					}
				}()
				done <- next(c)
			}()

			select {
			case err := <-done:
				if err != nil {
					if errors.Is(ctx.Err(), context.DeadlineExceeded) {
						// Handle context timeout error here
						return commonhttp.RestErrorResponse(c, http.StatusGatewayTimeout, models.ErrTimeout)
					}
					return err
				}
			case <-ctx.Done():
				return commonhttp.RestErrorResponse(c, http.StatusGatewayTimeout, models.ErrTimeout)
			}

			return nil
		}
	}
}
