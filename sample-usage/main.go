package main

import (
	"net/http"
	"os"

	"github.com/asamedeiros/opentelemetry-golang-lib/otelconfig"
	"github.com/asamedeiros/opentelemetry-golang-lib/otelecho"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"go.opentelemetry.io/otel/metric"
)

const (
	APP_NAME    = "opentelemetry-golang-sample-usage"
	APP_VERSION = "1.0.0"
)

/* func echoContextToHTTPRequest(c echo.Context) *otelconfig.HttpRequest {

	h := c.Request().Header
	rHeader := make(map[string]string)
	for k := range h {
		rHeader[k] = h[k][0]
	}

	rPath := c.Request().URL.Path
	rMethod := c.Request().Method
	rHost := c.Request().Host
	rRawQuery := c.QueryString()
	rRemoteAddr := c.RealIP()

	return &otelconfig.HttpRequest{
		Header: rHeader,
		Method: rMethod,
		// RawBody: rBody,
		URL: &url.URL{
			Host:     rHost,
			Path:     rPath,
			RawQuery: rRawQuery,
		},
		RemoteAddr: rRemoteAddr,
	}
} */

type Response struct {
	Status string `json:"status"`
}

func main() {

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "staging"
	}

	logManager, tracer, meter := otelconfig.StartOpenTelemetry(APP_NAME, APP_VERSION, environment, false)
	defer otelconfig.StopOpenTelemetry()

	apiCounter, _ := meter.Int64Counter(
		"api.counter",
		metric.WithDescription("Number of API calls."),
		metric.WithUnit("{call}"),
	)

	e := echo.New()

	otelecho.ConfigOtel(e, APP_NAME)

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.IPExtractor = echo.ExtractIPFromRealIPHeader()
	e.HideBanner = true

	e.GET("/*", func(c echo.Context) error {
		// logManager.NoTracer().Infof("Request received for %s", c.Request().URL.Path)
		return c.JSON(http.StatusOK, &Response{Status: "healthy"})
	})

	e.GET("/test", func(c echo.Context) error {

		ctxTracer, span := tracer.Start(c.Request().Context(), "endpoint_name")
		defer span.End()

		logManager := logManager.WithHttpData(otelecho.EchoContextToHTTPRequest(c))

		// This is an example of how to create a request whitout the otelecho utility
		// logManager := logManager.WithHttpData(echoContextToHTTPRequest(c))

		apiCounter.Add(c.Request().Context(), 1)

		loggerNoTracer := logManager.NoTracer()
		loggerNoTracer.Info("Log without tracer")

		loggerWithTracer := logManager.WithTracer(ctxTracer)
		loggerWithTracer.Info("Log with tracer")
		loggerWithTracer.With("context.user_id", "q98349286493").Info("Log with tracer and additional attribute")

		return c.JSON(http.StatusOK, &Response{Status: "ok"})
	})

	logManager.NoTracer().Info("Starting application")
	e.Logger.Fatal(e.Start(":3000"))
}
