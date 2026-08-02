package main

import (
	"context"
	"errors"
	"flag"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/vectorcore/esmlc/internal/config"
	"github.com/vectorcore/esmlc/internal/observability"
	"github.com/vectorcore/esmlc/internal/sls"
	"log/slog"
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

	server := sls.New(cfg, log)
	if cfg.Observability.Enabled {
		startObservabilityServer(ctx, cfg, log, server)
	}
	if e = server.Listen(ctx); e != nil {
		log.Error("esmlc stopped", "error", e)
		os.Exit(1)
	}
}

// startObservabilityServer runs the /metrics, /healthz, /readyz HTTP
// listener in the background, independent of the SCTP transport. It shuts
// down alongside the same signal context the SCTP listener uses.
func startObservabilityServer(ctx context.Context, cfg config.Config, log *slog.Logger, server *sls.Server) {
	addr := net.JoinHostPort(cfg.Observability.ListenAddress, strconv.Itoa(cfg.Observability.Port))
	httpServer := &http.Server{Addr: addr, Handler: observability.Handler(server)}
	go func() {
		log.Info("esmlc.observability.listening", "address", addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("esmlc.observability.failed", "error", err)
		}
	}()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()
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
