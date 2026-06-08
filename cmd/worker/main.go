package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/juanlou1217/exitfleet/internal/config"
	"github.com/juanlou1217/exitfleet/internal/harness"
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

	fmt.Printf("%s worker proxy listening on %s via %s\n", harness.ProjectName, cfg.ProxyAddress(), cfg.TunDevice)
	w := worker.New(worker.Options{
		Config: worker.OpenVPNConfig{
			Command:    cfg.OpenVPNCommand,
			AuthFile:   cfg.OpenVPNAuthFile,
			TunDevice:  cfg.TunDevice,
			ConfigFile: "",
		},
	})

	mux := http.NewServeMux()
	mux.Handle("/healthz", w.HealthHandler())
	mux.Handle("/readyz", w.ReadyHandler())

	server := &http.Server{
		Addr:              cfg.ProxyAddress(),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "worker server error: %v\n", err)
		os.Exit(1)
	}
}
