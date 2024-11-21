package bootstrap

import (
	"bank_test/internal/conf"
	"bank_test/internal/enum"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapLogger creates a new zap logger with the specified log level.
func NewZapLogger() (*zap.SugaredLogger, error) {
	pe := zap.NewProductionEncoderConfig()

	pe.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(pe)

	level := zap.InfoLevel
	if conf.GlobalConfig.LogLevel == enum.Debug {
		level = zap.DebugLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
	)

	return zap.New(core, zap.AddCaller()).Sugar(), nil
}

var Logger *zap.SugaredLogger
