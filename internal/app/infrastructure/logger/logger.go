package logger

import (
	"context"

	"github.com/duchuongnguyen/dhcp2p/internal/app/infrastructure/config"
	"github.com/mattn/go-colorable"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func getLoggerLevel(cfg *config.AppConfig) zap.AtomicLevel {
	switch cfg.LogLevel {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	return zap.NewAtomicLevelAt(zap.InfoLevel)
}

func NewLogger(lc fx.Lifecycle, cfg *config.AppConfig) *zap.Logger {
	stdout := zapcore.AddSync(colorable.NewColorableStdout())
	file := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     7, // days
	})

	level := getLoggerLevel(cfg)

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	productionCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, level),
		zapcore.NewCore(fileEncoder, file, level),
	)

	logger := zap.New(core)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return logger.Sync()
		},
	})

	return logger
}
