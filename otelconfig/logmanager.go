package otelconfig

import (
	"context"

	"go.uber.org/zap"
)

func NewLogManager(zapLogger *zap.Logger) *LogManager {
	return &LogManager{
		zapLogger: zapLogger,
	}
}

type LogManager struct {
	zapLogger *zap.Logger
}

func (c *LogManager) WithHttpData(req *HttpRequest) *LogManager {

	if req == nil {
		return c
	}
	return NewLogManager(c.zapLogger.With(zap.String("network.client.ip", req.RemoteAddr)).
		With(zap.String("http.request.host", req.URL.Host)).
		With(zap.String("http.request.method", req.Method)).
		With(zap.String("http.request.path", req.URL.Path)).
		With(zap.String("http.request.query", req.URL.RawQuery)).
		With(zap.String("http.request.headers.x-request-id", req.GetHeader("X-Request-Id"))).
		With(zap.String("http.request.headers.user-agent", req.GetHeader("User-Agent"))).
		With(zap.String("http.request.headers.cf-ray", req.GetHeader("CF-Ray"))).
		With(zap.String("http.request.headers.aws-xray", req.GetHeader("X-Amzn-Trace-Id"))))
}

func (c *LogManager) WithTracer(ctxTrace context.Context) LoggerI {

	if ctxTrace == nil {
		return NewLogger(c.zapLogger)
	}
	return NewLogger(c.zapLogger.With(zap.Any("context", ctxTrace)))
}

func (c *LogManager) NoTracer() LoggerI {
	return NewLogger(c.zapLogger)
}
