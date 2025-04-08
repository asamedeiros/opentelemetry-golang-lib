package otelconfig

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/contrib/detectors/aws/ecs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	otelMetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	otelTrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var shutdowns []func()

func newResource(ctx context.Context, appName, appVersion, environment string) (*resource.Resource, error) {
	return resource.New(ctx,
		// Use the ECS resource detector!
		resource.WithDetectors(ecs.NewResourceDetector()),
		// Keep the default detectors
		resource.WithTelemetrySDK(),
		// Add your own custom attributes to identify your application
		resource.WithAttributes(
			semconv.ServiceName(appName),
			semconv.ServiceVersion(appVersion),
			semconv.DeploymentEnvironment(environment),
		))
}

func configLog(ctx context.Context, rsc *resource.Resource, appName string, grpcSupported bool) (*LogManager, func(), error) {
	// Create a logger provider.
	// You can pass this instance directly when creating bridges.
	var exporter log.Exporter
	var err error
	if grpcSupported {
		exporter, err = otlploggrpc.New(ctx)
	} else {
		exporter, err = otlploghttp.New(ctx)
	}
	if err != nil || os.Getenv("ENVIRONMENT") == "development" {
		zapLog, _ := zap.NewProduction(zap.AddCallerSkip(1))
		// Criando um log emergencial, para conseguir reportar errors, já que não chegará a instanciar o oficial mais abaixo
		return NewLogManager(zapLog), func() {}, err
	}
	processor := log.NewBatchProcessor(exporter)
	loggerProvider := log.NewLoggerProvider(
		log.WithResource(rsc),
		log.WithProcessor(processor),
	)

	// Register as global logger provider so that it can be accessed global.LoggerProvider.
	// Most log bridges use the global logger provider as default.
	// If the global logger provider is not set then a no-op implementation
	// is used, which fails to generate data.
	global.SetLoggerProvider(loggerProvider)

	// Initialize a zap zaplogger with the otelzap bridge core.
	// This method actually doesn't log anything on your STDOUT, as everything
	// is shipped to a configured otel endpoint.
	zaplogger := zap.New(otelzap.NewCore(appName, otelzap.WithLoggerProvider(loggerProvider)))

	log := NewLogManager(zaplogger)

	shutdown := func() {
		if err := log.NoTracer().Sync(); err != nil && !strings.Contains(err.Error(), "sync /dev/stdout: invalid argument") {
			log.NoTracer().Warn("error syncing logger: " + err.Error())
		}

		if loggerProvider != nil {
			if err := loggerProvider.Shutdown(ctx); err != nil {
				log.NoTracer().Warn("error shutting down logger provider: " + err.Error())
			}
		}
	}

	return log, shutdown, nil
}

func configTracerProvider(ctx context.Context, logManager *LogManager, rsc *resource.Resource, grpcSupported bool) (func(), error) {
	var traceExporter trace.SpanExporter
	var err error
	if grpcSupported {
		traceExporter, err = otlptracegrpc.New(ctx)
	} else {
		traceExporter, err = otlptracehttp.New(ctx)
	}
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(rsc),
	)

	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	shutdown := func() {
		if tracerProvider != nil {
			if err := tracerProvider.Shutdown(ctx); err != nil {
				logManager.NoTracer().Warn("error shutting down tracer provider: " + err.Error())
			}
		}
	}

	return shutdown, nil
}

func configMetricProvider(ctx context.Context, logManager *LogManager, rsc *resource.Resource, grpcSupported bool) (func(), error) {
	var metricExporter metric.Exporter
	var err error
	if grpcSupported {
		metricExporter, err = otlpmetricgrpc.New(ctx)
	} else {
		metricExporter, err = otlpmetrichttp.New(ctx)
	}
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		metric.WithResource(rsc),
	)
	otel.SetMeterProvider(meterProvider)

	shutdown := func() {
		if meterProvider != nil {
			if err := meterProvider.Shutdown(ctx); err != nil {
				logManager.NoTracer().Warn("error shutting down metric provider: " + err.Error())
			}
		}
	}

	return shutdown, nil
}

func StartOpenTelemetry(appName, appVersion, environment string, grpcSupported bool) (*LogManager, otelTrace.Tracer, otelMetric.Meter) {

	ctx := context.Background()

	rsc, err := newResource(ctx, appName, appVersion, environment)
	if err != nil {
		fmt.Printf("error in plugin - failed to create resource for log provider: %s", err)
	}

	logManager, shutdownLog, err := configLog(ctx, rsc, appName, grpcSupported)
	if err != nil {
		logManager.NoTracer().Errorf("error in plugin - failed to create log or log provider: %s", err.Error())
	}
	shutdowns = append(shutdowns, shutdownLog)

	shutdownTracerProvider, err := configTracerProvider(ctx, logManager, rsc, grpcSupported)
	if err != nil {
		logManager.NoTracer().Errorf("error in plugin - failed to create trace provider: %s", err.Error())
	}
	shutdowns = append(shutdowns, shutdownTracerProvider)

	tracer := otel.GetTracerProvider().Tracer(appName)

	shutdownMetricProvider, err := configMetricProvider(ctx, logManager, rsc, grpcSupported)
	if err != nil {
		logManager.NoTracer().Errorf("error in plugin - failed to create metric provider: %s", err.Error())
	}
	shutdowns = append(shutdowns, shutdownMetricProvider)

	metric := otel.GetMeterProvider().Meter(appName)

	return logManager, tracer, metric
}

func StopOpenTelemetry() {
	for _, s := range shutdowns {
		s()
	}
}
