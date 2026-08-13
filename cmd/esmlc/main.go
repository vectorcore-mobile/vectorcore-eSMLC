package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/vectorcore/esmlc/internal/config"
	"github.com/vectorcore/esmlc/internal/logging"
	"github.com/vectorcore/esmlc/internal/observability"
	"github.com/vectorcore/esmlc/internal/sls"
	"log/slog"
)

func main() {
	fmt.Println("Starting VectoreCore ESMLC")
	path := flag.String("c", "config/esmlc.yaml", "YAML configuration path")
	debugConsole := flag.Bool("d", false, "enable debug logging on the console")
	flag.Parse()
	cfg, e := config.Load(*path)
	if e != nil {
		slog.Error("esmlc.startup failed", "error", e)
		os.Exit(1)
	}
	log, logFile, e := logging.New(cfg.Service, *debugConsole)
	if e != nil {
		slog.Error("esmlc.startup failed", "error", e)
		os.Exit(1)
	}
	defer func() { _ = logFile.Close() }()
	log.Info("esmlc.startup", "service", cfg.Service.Name, "debug_console", *debugConsole)
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

