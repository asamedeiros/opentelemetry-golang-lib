package otelconfig

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerI interface {
	With(name, value string) LoggerI
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Sync() error
	Error(msg string)
	Errorf(template string, args ...interface{})
	Warn(msg string)
	Warnf(template string, args ...interface{})
	Fatal(msg string)
	Info(msg string)
	Infof(template string, args ...interface{})
	Level() zapcore.Level
}

func NewLogger(zapLogger *zap.Logger) LoggerI {
	return &logger{
		zapLogger: zapLogger,
	}
}

type logger struct {
	zapLogger *zap.Logger
}

func (c *logger) Level() zapcore.Level {
	return c.zapLogger.Level()
}

func (c *logger) with(field zap.Field) LoggerI {
	return NewLogger(c.zapLogger.With(field))
}

func (c *logger) With(name, value string) LoggerI {
	return c.with(zap.String(name, value))
}

func (c *logger) Sync() error {
	return c.zapLogger.Sync()
}

func (c *logger) Error(msg string) {
	c.zapLogger.Error(msg)
}

func (c *logger) Errorf(template string, args ...interface{}) {
	c.zapLogger.Error(fmt.Sprintf(template, args...))
}

func (c *logger) Warn(msg string) {
	c.zapLogger.Warn(msg)
}

func (c *logger) Warnf(template string, args ...interface{}) {
	c.zapLogger.Warn(fmt.Sprintf(template, args...))
}

func (c *logger) Debug(args ...interface{}) {
	c.zapLogger.Debug(fmt.Sprint(args...))
}

func (c *logger) Debugf(template string, args ...interface{}) {
	c.zapLogger.Debug(fmt.Sprintf(template, args...))
}

func (c *logger) Fatal(msg string) {
	c.zapLogger.Fatal(msg)
}

func (c *logger) Info(msg string) {
	c.zapLogger.Info(msg)
}

func (c *logger) Infof(template string, args ...interface{}) {
	c.zapLogger.Info(fmt.Sprintf(template, args...))
}
