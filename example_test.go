package sloghandler_test

import (
	"log/slog"
	"os"

	"github.com/bnixon67/sloghandler"
)

func Example() {
	// Create a custom log handler
	handler := sloghandler.NewLogFormatHandler(slog.LevelDebug, os.Stdout, sloghandler.DefaultTimeFormat)

	// Set up the logger with the custom handler
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Use default logger
	slog.Info("example", "level", "info")
	slog.Debug("example", slog.String("level", "debug"))
}
