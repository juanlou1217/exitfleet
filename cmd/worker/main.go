package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/juanlou1217/exitfleet/internal/config"
	"github.com/juanlou1217/exitfleet/internal/harness"
	"github.com/juanlou1217/exitfleet/internal/proxy"
	"github.com/juanlou1217/exitfleet/internal/worker"
)

func main() {
	env, err := config.LoadEnvFile(".env")
	if err != nil {
		fmt.Fprintf(os.Stderr, "worker environment error: %v\n", err)
		os.Exit(1)
	}
	cfg, err := config.LoadWorker(env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "worker configuration error: %v\n", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "worker configuration error: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("%s worker proxy listening on %s via %s\n", harness.ProjectName, cfg.ProxyAddress(), cfg.TunDevice)
	fmt.Printf("%s worker health listening on %s\n", harness.ProjectName, cfg.HealthAddress())

	healthChecker := worker.NewHealthChecker(worker.HealthOptions{
		URLs:          cfg.HealthURLs,
		FailThreshold: cfg.HealthFailThreshold,
		Timeout:       time.Duration(cfg.HealthTimeoutSeconds) * time.Second,
	})
	w := worker.New(worker.Options{
		Config: worker.OpenVPNConfig{
			Command:    cfg.OpenVPNCommand,
			AuthFile:   cfg.OpenVPNAuthFile,
			TunDevice:  cfg.TunDevice,
			ConfigFile: "",
		},
		HealthChecker: healthChecker,
	})
	startHealthLoop(ctx, w, time.Duration(cfg.HealthIntervalSeconds)*time.Second)

	proxyServer := proxy.NewServer(proxy.Config{
		BindHost:       cfg.ProxyHost,
		BindPort:       cfg.ProxyPort,
		OutboundIP:     cfg.ProxyOutboundIP,
		MaxConnections: cfg.SOCKSMaxConnections,
	})
	errCh := make(chan error, 2)
	go func() {
		errCh <- proxyServer.ListenAndServe(ctx)
	}()

	mux := http.NewServeMux()
	mux.Handle("/healthz", w.HealthHandler())
	mux.Handle("/readyz", w.ReadyHandler())

	server := &http.Server{
		Addr:              cfg.HealthAddress(),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	go func() {
		errCh <- server.ListenAndServe()
	}()

	if err := <-errCh; err != nil && err != http.ErrServerClosed {
		stop()
		fmt.Fprintf(os.Stderr, "worker server error: %v\n", err)
		os.Exit(1)
	}
}

func startHealthLoop(ctx context.Context, w *worker.Worker, interval time.Duration) {
	if interval <= 0 {
		interval = 10 * time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w.CheckHealth(ctx)
			}
		}
	}()
}
