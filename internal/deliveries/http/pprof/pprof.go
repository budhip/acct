package pprof

import (
	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
)

// New is to register debug pprof.
func New(e *echo.Echo) {
	pprof.Register(e)
}
