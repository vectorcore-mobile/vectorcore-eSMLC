package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vectorcore/esmlc/internal/config"
)

func TestNewAlwaysWritesToLogFileAtConfiguredLevel(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "esmlc.log")
	cfg := config.Service{LogLevel: "warn", LogFile: logFile}

	log, closer, e := New(cfg, false)
	if e != nil {
		t.Fatalf("New: %v", e)
	}
	defer closer.Close()

	log.Info("should be filtered out by warn level")
	log.Warn("should be written")

	b, e := os.ReadFile(logFile)
	if e != nil {
		t.Fatalf("read log file: %v", e)
	}
	out := string(b)
	if strings.Contains(out, "should be filtered out") {
		t.Fatalf("info record was written despite warn level: %q", out)
	}
	if !strings.Contains(out, "should be written") {
		t.Fatalf("warn record missing from log file: %q", out)
	}
}

func TestDebugConsoleFlagControlsWhetherAConsoleHandlerExists(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Service{LogLevel: "info", LogFile: filepath.Join(dir, "esmlc.log")}

	withoutConsole, closer, e := New(cfg, false)
	if e != nil {
		t.Fatalf("New: %v", e)
	}
	defer closer.Close()
	mh, ok := withoutConsole.Handler().(*multiHandler)
	if !ok {
		t.Fatalf("expected *multiHandler, got %T", withoutConsole.Handler())
	}
	if len(mh.handlers) != 1 {
		t.Fatalf("expected only the file handler without -d, got %d handlers", len(mh.handlers))
	}

	withConsole, closer2, e := New(cfg, true)
	if e != nil {
		t.Fatalf("New: %v", e)
	}
	defer closer2.Close()
	mh2, ok := withConsole.Handler().(*multiHandler)
	if !ok {
		t.Fatalf("expected *multiHandler, got %T", withConsole.Handler())
	}
	if len(mh2.handlers) != 2 {
		t.Fatalf("expected file+console handlers with -d, got %d handlers", len(mh2.handlers))
	}
}

func TestDebugConsoleDoesNotChangeLogFileLevel(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "esmlc.log")
	cfg := config.Service{LogLevel: "error", LogFile: logFile}

	log, closer, e := New(cfg, true)
	if e != nil {
		t.Fatalf("New: %v", e)
	}
	defer closer.Close()

	log.Debug("debug record: file should still drop this")
	log.Warn("warn record: file should still drop this")
	log.Error("error record: file should keep this")

	b, e := os.ReadFile(logFile)
	if e != nil {
		t.Fatalf("read log file: %v", e)
	}
	out := string(b)
	if strings.Contains(out, "debug record") || strings.Contains(out, "warn record") {
		t.Fatalf("log_level=error in YAML was widened by -d console flag: %q", out)
	}
	if !strings.Contains(out, "error record") {
		t.Fatalf("error record missing from log file: %q", out)
	}
}

func TestLevel(t *testing.T) {
	cases := map[string]bool{"debug": true, "info": true, "warn": true, "error": true, "": true, "bogus": true}
	for v := range cases {
		_ = Level(v) // must not panic for any input; config.Validate rejects unknown non-empty values upstream
	}
}
