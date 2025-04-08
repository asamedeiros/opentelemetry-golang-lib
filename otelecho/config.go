package otelecho

import (
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func ConfigOtel(e *echo.Echo, appName string) {
	e.Use(otelecho.Middleware(appName))
}
