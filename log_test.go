package sloghandler

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestLogFormatHandler_Handle(t *testing.T) {
	tests := []struct {
		name       string
		level      slog.Level
		record     slog.Record
		group      string
		attrs      []slog.Attr
		wantOutput string
	}{
		{
			name:  "Info level with no attributes",
			level: slog.LevelInfo,
			record: slog.NewRecord(
				time.Date(2025, 1, 11, 12, 0, 0, 0, time.UTC),
				slog.LevelInfo,
				"Test message",
				0,
			),
			wantOutput: "2025/01/11 12:00:00 INFO Test message\n",
		},
		{
			name:  "Error level with attributes",
			level: slog.LevelError,
			record: func() slog.Record {
				r := slog.NewRecord(
					time.Date(2025, 1, 11, 12, 0, 0, 0, time.UTC),
					slog.LevelError,
					"Error occurred",
					0,
				)
				r.AddAttrs(slog.String("key", "value"), slog.Int("code", 123))
				return r
			}(),
			wantOutput: "2025/01/11 12:00:00 ERROR Error occurred key=value code=123\n",
		},
		{
			name:  "Debug level ignored due to handler level",
			level: slog.LevelInfo,
			record: slog.NewRecord(
				time.Date(2025, 1, 11, 12, 0, 0, 0, time.UTC),
				slog.LevelDebug,
				"Should not log",
				0,
			),
			wantOutput: "",
		},
		{
			name:  "Grouped log message",
			level: slog.LevelInfo,
			record: slog.NewRecord(
				time.Date(2025, 1, 11, 12, 0, 0, 0, time.UTC),
				slog.LevelInfo,
				"Grouped message",
				0,
			),
			group:      "TestGroup",
			wantOutput: "2025/01/11 12:00:00 INFO [TestGroup] Grouped message\n",
		},
		{
			name:  "Handler-level attributes",
			level: slog.LevelInfo,
			record: slog.NewRecord(
				time.Date(2025, 1, 11, 12, 0, 0, 0, time.UTC),
				slog.LevelInfo,
				"Handler attributes",
				0,
			),
			attrs:      []slog.Attr{slog.String("env", "production")},
			wantOutput: "2025/01/11 12:00:00 INFO Handler attributes env=production\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			handler := NewLogFormatHandler(tt.level, os.Stdout)
			handler.writer = &buf // Override writer for testing

			// Apply grouping and attributes if specified
			if tt.group != "" {
				handler = handler.WithGroup(tt.group).(*LogFormatHandler)
			}
			if len(tt.attrs) > 0 {
				handler = handler.WithAttrs(tt.attrs).(*LogFormatHandler)
			}

			// Call Handle
			err := handler.Handle(context.Background(), tt.record)
			if err != nil {
				t.Errorf("Handle() error = %v", err)
				return
			}

			// Check output
			gotOutput := buf.String()
			if gotOutput != tt.wantOutput {
				t.Errorf("Output mismatch\nGot:  %q\nWant: %q", gotOutput, tt.wantOutput)
			}
		})
	}
}

func TestLogFormatHandler_Enabled(t *testing.T) {
	tests := []struct {
		name      string
		handlerLv slog.Level
		recordLv  slog.Level
		want      bool
	}{
		{"Enabled for equal levels", slog.LevelInfo, slog.LevelInfo, true},
		{"Enabled for higher levels", slog.LevelInfo, slog.LevelError, true},
		{"Disabled for lower levels", slog.LevelInfo, slog.LevelDebug, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewLogFormatHandler(tt.handlerLv, os.Stdout)
			if got := handler.Enabled(context.Background(), tt.recordLv); got != tt.want {
				t.Errorf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
