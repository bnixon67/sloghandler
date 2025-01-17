// Package sloghandler provides a custom slog.Handler implementation that
// formats log output similarly to the default log package. It supports log
// levels, attributes, grouping, and ensures thread-safe writes.
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

const DefaultTimeFormat = "2006/01/02 15:04:05"

// LogFormatHandler is a custom log handler that formats logs to similar to
// the  default log output.
type LogFormatHandler struct {
	level      slog.Level   // Min log level that this handler processes.
	writer     io.Writer    // Destination for log messages.
	attrs      []slog.Attr  // Additional attrs included in every log entry.
	group      string       // Optional name for grouping log messages.
	timeFormat string       // Format for timestamps in log messages.
	mu         sync.RWMutex // Protects concurrent writes to the log output.
}

// Handle processes a log record, formats it, and writes it to the output.
func (h *LogFormatHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level < h.level {
		return nil
	}

	var buf bytes.Buffer

	// Format timestamp, level, and message
	buf.WriteString(r.Time.Format(h.timeFormat))
	buf.WriteString(" ")
	buf.WriteString(strings.ToUpper(r.Level.String()))
	buf.WriteString(" ")

	if h.group != "" {
		buf.WriteString("[")
		buf.WriteString(h.group)
		buf.WriteString("] ")
	}

	buf.WriteString(r.Message)

	// Append record attributes
	r.Attrs(func(a slog.Attr) bool {
		if a.Key != "" {
			buf.WriteString(fmt.Sprintf(" %s=%v", a.Key, a.Value))
		}
		return true
	})

	// Append handler-level attributes
	for _, attr := range h.attrs {
		if attr.Key != "" {
			buf.WriteString(fmt.Sprintf(" %s=%v", attr.Key, attr.Value))
		}
	}

	buf.WriteString("\n")

	// Write the log with thread safety
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, err := h.writer.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write log: %w", err)
	}

	return nil
}

// Enabled reports whether the handler processes logs at the given level.
func (h *LogFormatHandler) Enabled(ctx context.Context, level slog.Level) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return level >= h.level
}

// WithAttrs returns a new handler with the given attributes added.
func (h *LogFormatHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := append(h.attrs, attrs...)
	return &LogFormatHandler{
		level:      h.level,
		writer:     h.writer,
		attrs:      newAttrs,
		group:      h.group,
		timeFormat: h.timeFormat,
		mu:         h.mu,
	}
}

// WithGroup returns a new handler with the specified group name.
func (h *LogFormatHandler) WithGroup(name string) slog.Handler {
	return &LogFormatHandler{
		level:      h.level,
		writer:     h.writer,
		attrs:      h.attrs,
		group:      name,
		timeFormat: h.timeFormat,
		mu:         h.mu,
	}
}

// NewLogFormatHandler creates a new LogFormatHandler with the given log level,
// writer, and timestamp format.
func NewLogFormatHandler(level slog.Level, writer io.Writer, timeFormat string) *LogFormatHandler {
	if timeFormat == "" {
		timeFormat = DefaultTimeFormat
	}

	return &LogFormatHandler{
		level:      level,
		writer:     writer,
		timeFormat: timeFormat,
	}
}
