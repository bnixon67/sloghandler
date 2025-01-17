package sloghandler

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestLogFormatHandler_Handle tests log message formatting and attribute handling.
func TestLogFormatHandler_Handle(t *testing.T) {
	tests := []struct {
		name            string
		level           slog.Level
		record          slog.Record
		group           string
		attrs           []slog.Attr
		wantOutput      string
		removeTimestamp bool
	}{
		{
			name:  "Info level log without attributes",
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
			name:  "Error level log with attributes",
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
			name:  "Debug level ignored due to higher handler level",
			level: slog.LevelInfo,
			record: slog.NewRecord(
				time.Now(),
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
				time.Now(),
				slog.LevelInfo,
				"Grouped message",
				0,
			),
			group:           "TestGroup",
			wantOutput:      "INFO [TestGroup] Grouped message\n",
			removeTimestamp: true,
		},
		{
			name:  "Handler-level attributes applied",
			level: slog.LevelInfo,
			record: slog.NewRecord(
				time.Now(),
				slog.LevelInfo,
				"Handler attributes",
				0,
			),
			attrs:           []slog.Attr{slog.String("env", "production")},
			wantOutput:      "INFO Handler attributes env=production\n",
			removeTimestamp: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			handler := NewLogFormatHandler(tc.level, &buf, DefaultTimeFormat)

			// Apply grouping and attributes if specified
			if tc.group != "" {
				handler = handler.WithGroup(tc.group).(*LogFormatHandler)
			}
			if len(tc.attrs) > 0 {
				handler = handler.WithAttrs(tc.attrs).(*LogFormatHandler)
			}

			// Call Handle
			err := handler.Handle(context.Background(), tc.record)
			if err != nil {
				t.Fatalf("Handle() error = %v", err)
			}

			gotOutput := buf.String()
			if tc.removeTimestamp {
				gotOutput = removeTimestamp(gotOutput)
			}

			if gotOutput != tc.wantOutput {
				t.Errorf("Output mismatch\nGot:  %q\nWant: %q", gotOutput, tc.wantOutput)
			}
		})
	}
}

// removeTimestamp removes timestamp from log output for consistent testing.
func removeTimestamp(output string) string {
	re := regexp.MustCompile(`^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} `)
	return re.ReplaceAllString(output, "")
}

// TestLogFormatHandler_Enabled tests log level filtering.
func TestLogFormatHandler_Enabled(t *testing.T) {
	tests := []struct {
		name      string
		handlerLv slog.Level
		recordLv  slog.Level
		want      bool
	}{
		{"Enabled at same level", slog.LevelInfo, slog.LevelInfo, true},
		{"Enabled for higher levels", slog.LevelInfo, slog.LevelError, true},
		{"Disabled for lower levels", slog.LevelInfo, slog.LevelDebug, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewLogFormatHandler(tc.handlerLv, os.Stdout, DefaultTimeFormat)
			if got := handler.Enabled(context.Background(), tc.recordLv); got != tc.want {
				t.Errorf("Enabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestLogFormatHandler_WithAttrs verifies that attributes persist across instances.
func TestLogFormatHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	handler := NewLogFormatHandler(slog.LevelInfo, &buf, DefaultTimeFormat)
	handler = handler.WithAttrs([]slog.Attr{slog.String("app", "testApp")}).(*LogFormatHandler)

	// Log a message
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "Attribute test", 0)
	_ = handler.Handle(context.Background(), record)

	gotOutput := removeTimestamp(buf.String())
	wantOutput := "INFO Attribute test app=testApp\n"
	if gotOutput != wantOutput {
		t.Errorf("WithAttrs() output mismatch\nGot:  %q\nWant: %q", gotOutput, wantOutput)
	}
}

// TestLogFormatHandler_WithGroup verifies that group names are correctly applied.
func TestLogFormatHandler_WithGroup(t *testing.T) {
	var buf bytes.Buffer
	handler := NewLogFormatHandler(slog.LevelInfo, &buf, DefaultTimeFormat)
	handler = handler.WithGroup("MyGroup").(*LogFormatHandler)

	// Log a message
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "Grouped log", 0)
	_ = handler.Handle(context.Background(), record)

	gotOutput := removeTimestamp(buf.String())
	wantOutput := "INFO [MyGroup] Grouped log\n"
	if gotOutput != wantOutput {
		t.Errorf("WithGroup() output mismatch\nGot:  %q\nWant: %q", gotOutput, wantOutput)
	}
}

// TestLogFormatHandler_Concurrent verifies thread safety for concurrent logging.
func TestLogFormatHandler_Concurrent(t *testing.T) {
	var buf bytes.Buffer
	handler := NewLogFormatHandler(slog.LevelInfo, &buf, DefaultTimeFormat)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			record := slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprintf("Concurrent log %d", i), 0)
			_ = handler.Handle(context.Background(), record)
		}(i)
	}
	wg.Wait()

	// Ensure at least 10 logs were written
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 10 {
		t.Errorf("Expected 10 log entries, got %d", len(lines))
	}
}
