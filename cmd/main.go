package main

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/bnixon67/sloghandler"
)

func main() {
	// Create a custom log handler
	handler := sloghandler.NewLogFormatHandler(slog.LevelDebug, os.Stdout)

	// Set up the logger with the custom handler
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Use default logger
	slog.Info("testing", "date", time.Now())

	// Use custom logger with various levels and attributes
	logger.Debug("Debug message", "foo", "bar")
	logger.Info("App started", slog.String("version", "1.0.0"))
	logger.Warn("Low space", slog.Float64("available", 1.2))
	logger.Error("Error Message", "error", errors.New("some error"))
}
