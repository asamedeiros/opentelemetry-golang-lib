package otelecho

import (
	"net/url"

	"github.com/asamedeiros/opentelemetry-golang-lib/otelconfig"
	"github.com/labstack/echo/v4"
)

func EchoContextToHTTPRequest(c echo.Context) *otelconfig.HttpRequest {

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
		URL: &url.URL{
			Host:     rHost,
			Path:     rPath,
			RawQuery: rRawQuery,
		},
		RemoteAddr: rRemoteAddr,
	}
}
