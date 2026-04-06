// Package logging provides structured logging utilities for the Catalogizer API.
// It wraps zap logger with application-specific configurations and helpers.
package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the global structured logger instance
var Logger *zap.Logger

// SugaredLogger is the sugared version of the global logger for convenient use
var SugaredLogger *zap.SugaredLogger

// Init initializes the global logger with the specified environment configuration
func Init(env string) error {
	config := zap.NewProductionConfig()

	if env == "development" || env == "dev" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.StacktraceKey = "stacktrace"

	var err error
	Logger, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return err
	}

	SugaredLogger = Logger.Sugar()
	return nil
}

// InitWithConfig initializes the logger with a custom configuration
func InitWithConfig(cfg zap.Config) error {
	var err error
	Logger, err = cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return err
	}

	SugaredLogger = Logger.Sugar()
	return nil
}

// Sync flushes any buffered log entries
func Sync() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	} else {
		os.Exit(1)
	}
}

// Debugf logs a formatted debug message
func Debugf(template string, args ...interface{}) {
	if SugaredLogger != nil {
		SugaredLogger.Debugf(template, args...)
	}
}

// Infof logs a formatted info message
func Infof(template string, args ...interface{}) {
	if SugaredLogger != nil {
		SugaredLogger.Infof(template, args...)
	}
}

// Warnf logs a formatted warning message
func Warnf(template string, args ...interface{}) {
	if SugaredLogger != nil {
		SugaredLogger.Warnf(template, args...)
	}
}

// Errorf logs a formatted error message
func Errorf(template string, args ...interface{}) {
	if SugaredLogger != nil {
		SugaredLogger.Errorf(template, args...)
	}
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(template string, args ...interface{}) {
	if SugaredLogger != nil {
		SugaredLogger.Fatalf(template, args...)
	} else {
		os.Exit(1)
	}
}

// With creates a child logger with the provided fields
func With(fields ...zap.Field) *zap.Logger {
	if Logger != nil {
		return Logger.With(fields...)
	}
	return nil
}

// WithFields creates a child logger with the provided fields (sugared version)
func WithFields(fields map[string]interface{}) *zap.SugaredLogger {
	if SugaredLogger != nil {
		return SugaredLogger.With(fields)
	}
	return nil
}

// Field helpers for common field types
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

func ErrorField(err error) zap.Field {
	return zap.Error(err)
}

func Duration(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

func Float64(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}
