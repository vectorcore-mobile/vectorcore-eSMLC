package main

import (
	"context"
	"flag"
	"github.com/vectorcore/esmlc/internal/config"
	"github.com/vectorcore/esmlc/internal/sls"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	path := flag.String("config", "config/esmlc.yaml", "YAML configuration path")
	flag.Parse()
	cfg, e := config.Load(*path)
	if e != nil {
		slog.Error("esmlc.startup failed", "error", e)
		os.Exit(1)
	}
	log := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level(cfg.Service.LogLevel)}))
	log.Info("esmlc.startup", "service", cfg.Service.Name)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if e = sls.New(cfg, log).Listen(ctx); e != nil {
		log.Error("esmlc stopped", "error", e)
		os.Exit(1)
	}
}
func level(v string) slog.Level {
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
