package otelhttp

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewOtelHTTPTransport(tr *http.Transport) *otelhttp.Transport {
	// var newTr http.RoundTripper = tr
	return otelhttp.NewTransport(tr)
}
