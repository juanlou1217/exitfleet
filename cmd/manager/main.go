package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/juanlou1217/exitfleet/internal/api"
	"github.com/juanlou1217/exitfleet/internal/config"
	"github.com/juanlou1217/exitfleet/internal/harness"
	"github.com/juanlou1217/exitfleet/internal/manager"
	"github.com/juanlou1217/exitfleet/internal/node"
)

func main() {
	cfg := config.DefaultManager()
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "manager configuration error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s manager listening on %s\n", harness.ProjectName, cfg.HTTPAddress())
	mgr := manager.New(manager.Options{
		Source: node.HTTPSource{URL: "https://www.vpngate.net/api/iphone/"},
		ProbeRunner: node.ProbeFunc(func(_ context.Context, n node.Node) (node.ProbeResult, error) {
			latency := time.Duration(n.PingMS) * time.Millisecond
			if latency == 0 {
				latency = time.Millisecond
			}
			return node.ProbeResult{Latency: latency}, nil
		}),
		Launcher:  skeletonLauncher{},
		ProxyBase: config.DefaultWorker().ProxyPort,
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddress(),
		Handler:           api.NewManagerHandler(mgr),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "manager server error: %v\n", err)
		os.Exit(1)
	}
}

type skeletonLauncher struct{}

func (skeletonLauncher) StartWorker(_ context.Context, request manager.LaunchRequest) (manager.WorkerRecord, error) {
	return manager.WorkerRecord{
		ID:        "worker-" + request.Node.ID,
		ProxyPort: request.ProxyPort,
		State:     "ready",
	}, nil
}

func (skeletonLauncher) StopWorker(_ context.Context, workerID string) error {
	return nil
}
