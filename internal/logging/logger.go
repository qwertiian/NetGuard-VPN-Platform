package logging

import (
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

var Logger *slog.Logger

var wgKeyRegex = regexp.MustCompile(`[A-Za-z0-9+/]{43}=`)

// Init initializes the package-level logger with the specified level.
func Init(level string) {
	var l slog.Level
	switch strings.ToUpper(level) {
	case "DEBUG":
		l = slog.LevelDebug
	case "INFO":
		l = slog.LevelInfo
	case "WARN":
		l = slog.LevelWarn
	case "ERROR":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: l,
	})
	Logger = slog.New(handler)
}

// SetOutput sets the output writer for the logger.
// It resets the level to INFO by default.
func SetOutput(w io.Writer) {
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Logger = slog.New(handler)
}

// ScrubSecrets replaces WireGuard private keys (base64 44-char patterns) with [REDACTED].
func ScrubSecrets(s string) string {
	return wgKeyRegex.ReplaceAllString(s, "[REDACTED]")
}
