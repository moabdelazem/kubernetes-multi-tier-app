package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// Init initializes the global logger
func Init(env string) error {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set encoding to JSON for production, console for development
	if env == "production" {
		config.Encoding = "json"
	} else {
		config.Encoding = "console"
	}

	logger, err := config.Build()
	if err != nil {
		return err
	}

	Log = logger
	return nil
}

// InitWithLevel initializes the logger with a specific level
func InitWithLevel(env string, level zapcore.Level) error {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.Level = zap.NewAtomicLevelAt(level)

	if env == "production" {
		config.Encoding = "json"
	} else {
		config.Encoding = "console"
	}

	logger, err := config.Build()
	if err != nil {
		return err
	}

	Log = logger
	return nil
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if Log == nil {
		// Fallback to development logger if not initialized
		Log, _ = zap.NewDevelopment()
	}
	return Log
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// With creates a child logger with the given fields
func With(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

// FromEnv initializes the logger based on environment variables
func FromEnv() error {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	return Init(env)
}
