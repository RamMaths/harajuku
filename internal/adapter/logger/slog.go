package logger

import (
	"log/slog"
	"os"

	"harajuku/backend/internal/adapter/config"
	slogmulti "github.com/samber/slog-multi"
	"gopkg.in/natefinch/lumberjack.v2"
)

// logger is the default logger used by the application
var logger *slog.Logger

// Set sets the logger configuration based on the environment
func Set(config *config.App) {
	// Create common options
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug, // This enables debug logging
	}

	// Default logger (development)
	logger = slog.New(
		slog.NewTextHandler(os.Stderr, opts), // Added options here
	)

	if config.Env == "production" {
		logRotate := &lumberjack.Logger{
			Filename:   "log/app.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}

		// Production uses Info level by default
		prodOpts := &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}

		logger = slog.New(
			slogmulti.Fanout(
				slog.NewJSONHandler(logRotate, prodOpts),
				slog.NewTextHandler(os.Stderr, opts), // Keep debug for stderr
			),
		)
	}

	slog.SetDefault(logger)
}
