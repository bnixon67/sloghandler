package sloghandler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
)

// LogFormatHandler is a custom log handler that formats logs to similar to
// the  default log output.
type LogFormatHandler struct {
	level  slog.Level  // Minimum log level to handle
	writer io.Writer   // Destination for log output
	attrs  []slog.Attr // Additional attributes for the handler
	group  string      // Optional log group for grouping messages
	mu     sync.Mutex  // Ensures thread-safe writes
}

// Handle formats the log message and writes it to the specified writer.
func (h *LogFormatHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level < h.level {
		return nil
	}

	// Use a buffer for efficient string building
	var buf bytes.Buffer

	// Format timestamp, level, and message
	buf.WriteString(r.Time.Format("2006/01/02 15:04:05"))
	buf.WriteString(" ")
	buf.WriteString(strings.ToUpper(r.Level.String()))
	buf.WriteString(" ")

	if h.group != "" {
		buf.WriteString("[")
		buf.WriteString(h.group)
		buf.WriteString("] ")
	}

	buf.WriteString(r.Message)

	// Collect record attributes
	r.Attrs(func(a slog.Attr) bool {
		buf.WriteString(fmt.Sprintf(" %s=%v", a.Key, a.Value))
		return true
	})

	// Add handler-level attributes
	for _, attr := range h.attrs {
		buf.WriteString(fmt.Sprintf(" %s=%v", attr.Key, attr.Value))
	}

	buf.WriteString("\n")

	// Write the log to the output with thread safety
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, err := h.writer.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write log: %w", err)
	}

	return nil
}

// Enabled checks if the specified log level is enabled for the handler.
func (h *LogFormatHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

// WithAttrs returns a new handler with additional attributes.
func (h *LogFormatHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := append(h.attrs, attrs...)
	return &LogFormatHandler{
		level:  h.level,
		writer: h.writer,
		attrs:  newAttrs,
		group:  h.group,
		mu:     sync.Mutex{},
	}
}

// WithGroup returns a new handler with a log group prefix.
func (h *LogFormatHandler) WithGroup(name string) slog.Handler {
	return &LogFormatHandler{
		level:  h.level,
		writer: h.writer,
		attrs:  h.attrs,
		group:  name,
		mu:     sync.Mutex{},
	}
}

// NewLogFormatHandler creates a new handler to emulate default log output.
func NewLogFormatHandler(level slog.Level, writer io.Writer) *LogFormatHandler {
	return &LogFormatHandler{
		level:  level,
		writer: writer,
		attrs:  []slog.Attr{},
		group:  "",
		mu:     sync.Mutex{},
	}
}
