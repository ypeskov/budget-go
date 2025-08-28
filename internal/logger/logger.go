package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

var Logger *slog.Logger

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[94m"  // Light blue (bright blue)
	ColorGray   = "\033[37m"
	ColorBold   = "\033[1m"
)

// ColoredTextHandler wraps slog.TextHandler to add colors to log levels
type ColoredTextHandler struct {
	slog.Handler
	w io.Writer
}

// NewColoredTextHandler creates a new colored text handler
func NewColoredTextHandler(w io.Writer, opts *slog.HandlerOptions) *ColoredTextHandler {
	return &ColoredTextHandler{
		Handler: slog.NewTextHandler(w, opts),
		w:       w,
	}
}

// Handle implements the slog.Handler interface with colored output
func (h *ColoredTextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Get color based on level
	var levelColor string
	var levelName string
	switch r.Level {
	case slog.LevelDebug:
		levelColor = ColorGray
		levelName = "DEBUG"
	case slog.LevelInfo:
		levelColor = ColorBlue
		levelName = "INFO "
	case slog.LevelWarn:
		levelColor = ColorYellow
		levelName = "WARN "
	case slog.LevelError:
		levelColor = ColorRed
		levelName = "ERROR"
	default:
		levelColor = ColorReset
		levelName = r.Level.String()
	}

	// Format timestamp
	timestamp := r.Time.Format(time.RFC3339)

	// Start building the log line
	logLine := fmt.Sprintf("time=%s %slevel=%s%s msg=\"%s\"",
		timestamp,
		levelColor,
		levelName,
		ColorReset,
		r.Message,
	)

	// Add attributes
	r.Attrs(func(a slog.Attr) bool {
		logLine += fmt.Sprintf(" %s=%v", a.Key, a.Value)
		return true
	})

	// Write the colored log line
	fmt.Fprintln(h.w, logLine)
	return nil
}

func Init(logLevel string) {
	var level slog.Level
	
	switch strings.ToLower(logLevel) {
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
		Level: level,
	}

	// Use colored handler if we're outputting to a terminal
	var handler slog.Handler
	if isTerminal(os.Stdout) {
		handler = NewColoredTextHandler(os.Stdout, opts)
	} else {
		// Use standard text handler for non-terminal output (logs, files, etc.)
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	Logger = slog.New(handler)
	
	// Set as default logger
	slog.SetDefault(Logger)
}

// Convenience functions for common logging patterns
func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	Logger.Error(msg, args...)
	os.Exit(1)
}

// With returns a logger with the given attributes
func With(args ...any) *slog.Logger {
	return Logger.With(args...)
}

// isTerminal checks if the output is a terminal (for color support)
func isTerminal(f *os.File) bool {
	// Simple check - this works for most cases
	// On Unix-like systems, check if it's a character device
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}