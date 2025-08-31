package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Config holds logging configuration
type Config struct {
	Level      string
	Format     string // "json" or "text"
	Output     string // "stdout", "stderr", or file path
	AddSource  bool
	ServiceName string
}

// Logger wraps slog.Logger with additional context
type Logger struct {
	*slog.Logger
	config Config
}

// NewLogger creates a new structured logger optimized for Loki
func NewLogger(config Config) *Logger {
	var handler slog.Handler
	var writer io.Writer

	// Determine output writer
	switch strings.ToLower(config.Output) {
	case "stderr":
		writer = os.Stderr
	case "stdout":
		writer = os.Stdout
	default:
		// Treat as file path
		// Ensure directory exists
		dir := filepath.Dir(config.Output)
		if err := os.MkdirAll(dir, 0755); err != nil {
			// Fallback to stdout if directory creation fails
			writer = os.Stdout
		} else {
			file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				// Fallback to stdout if file creation fails
				writer = os.Stdout
			} else {
				writer = file
			}
		}
	}

	// Parse log level
	var level slog.Level
	switch strings.ToLower(config.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: config.AddSource,
	}

	// Choose handler format
	switch strings.ToLower(config.Format) {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	default:
		handler = slog.NewTextHandler(writer, opts)
	}

	// Add service name to all logs
	if config.ServiceName != "" {
		handler = &serviceHandler{
			handler:      handler,
			serviceName: config.ServiceName,
		}
	}

	return &Logger{
		Logger: slog.New(handler),
		config: config,
	}
}

// serviceHandler adds service name to all log entries
type serviceHandler struct {
	handler     slog.Handler
	serviceName string
}

func (h *serviceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *serviceHandler) Handle(ctx context.Context, r slog.Record) error {
	// Add service name to the record
	r.AddAttrs(slog.String("service", h.serviceName))
	return h.handler.Handle(ctx, r)
}

func (h *serviceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &serviceHandler{
		handler:     h.handler.WithAttrs(attrs),
		serviceName: h.serviceName,
	}
}

func (h *serviceHandler) WithGroup(name string) slog.Handler {
	return &serviceHandler{
		handler:     h.handler.WithGroup(name),
		serviceName: h.serviceName,
	}
}

// WithContext returns a logger with context
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		Logger: l.Logger.With(),
		config: l.config,
	}
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields ...slog.Attr) *Logger {
	// Convert slog.Attr to any for slog.Logger.With
	args := make([]any, len(fields))
	for i, field := range fields {
		args[i] = field
	}
	return &Logger{
		Logger: l.Logger.With(args...),
		config: l.config,
	}
}

// RequestLogger logs HTTP requests in Loki-compatible format
func (l *Logger) RequestLogger(method, path string, status int, latencyMs int64, requestID string) {
	level := slog.LevelInfo
	if status >= 500 {
		level = slog.LevelError
	} else if status >= 400 {
		level = slog.LevelWarn
	}

	l.Logger.LogAttrs(context.Background(), level, "http_request",
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", status),
		slog.Int64("latency_ms", latencyMs),
		slog.String("request_id", requestID),
	)
}

// DatabaseLogger logs database operations
func (l *Logger) DatabaseLogger(operation, table string, durationMs int64, err error) {
	attrs := []slog.Attr{
		slog.String("operation", operation),
		slog.String("table", table),
		slog.Int64("duration_ms", durationMs),
	}

	if err != nil {
		l.Logger.LogAttrs(context.Background(), slog.LevelError, "database_operation",
			append(attrs, slog.String("error", err.Error()))...,
		)
	} else {
		l.Logger.LogAttrs(context.Background(), slog.LevelDebug, "database_operation", attrs...)
	}
}

// BusinessLogger logs business logic events
func (l *Logger) BusinessLogger(event string, userID, loanID string, amount float64, details map[string]interface{}) {
	attrs := []slog.Attr{
		slog.String("event", event),
		slog.String("user_id", userID),
		slog.String("loan_id", loanID),
		slog.Float64("amount", amount),
	}

	// Add additional details
	for k, v := range details {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, slog.String(k, val))
		case int:
			attrs = append(attrs, slog.Int(k, val))
		case float64:
			attrs = append(attrs, slog.Float64(k, val))
		case bool:
			attrs = append(attrs, slog.Bool(k, val))
		}
	}

	l.Logger.LogAttrs(context.Background(), slog.LevelInfo, "business_event", attrs...)
}
