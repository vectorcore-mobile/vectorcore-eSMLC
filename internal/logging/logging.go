// Package logging builds this service's *slog.Logger. Logging always goes
// to the configured logfile, at the level set in the YAML config
// (service.log_level); that file level is never affected by the console
// debug flag. The console (stderr) only receives log output at all when the
// operator passes -d, and when it does, it is always at debug level
// regardless of service.log_level — the flag widens what the console shows,
// it does not narrow or otherwise change what the logfile records.
package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/vectorcore/esmlc/internal/config"
)

// New opens cfg.LogFile and returns a logger that always writes to it at
// Level(cfg.LogLevel), and additionally to stderr at debug level when
// debugConsole is true. The returned io.Closer must be closed by the caller
// on shutdown.
func New(cfg config.Service, debugConsole bool) (*slog.Logger, io.Closer, error) {
	f, e := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if e != nil {
		return nil, nil, fmt.Errorf("logging: open log file: %w", e)
	}
	handlers := []slog.Handler{slog.NewJSONHandler(f, &slog.HandlerOptions{Level: Level(cfg.LogLevel)})}
	if debugConsole {
		handlers = append(handlers, slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return slog.New(newMultiHandler(handlers...)), f, nil
}

// Level maps a config log level string to its slog.Level, defaulting to
// info for the empty/unrecognized case (config.Validate already rejects
// unrecognized non-empty values).
func Level(v string) slog.Level {
	switch v {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// multiHandler fans a single log record out to every wrapped handler,
// applying each handler's own level filter independently. This is what lets
// the logfile and console run at different levels simultaneously.
type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) *multiHandler {
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		if e := h.Handle(ctx, r.Clone()); e != nil {
			return e
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithAttrs(attrs)
	}
	return newMultiHandler(next...)
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithGroup(name)
	}
	return newMultiHandler(next...)
}
