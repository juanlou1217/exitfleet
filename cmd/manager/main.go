package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/juanlou1217/exitfleet/internal/api"
	"github.com/juanlou1217/exitfleet/internal/config"
	workerdocker "github.com/juanlou1217/exitfleet/internal/docker"
	"github.com/juanlou1217/exitfleet/internal/harness"
	"github.com/juanlou1217/exitfleet/internal/manager"
	"github.com/juanlou1217/exitfleet/internal/node"
)

func main() {
	env, err := config.LoadEnvFile(".env")
	if err != nil {
		fmt.Fprintf(os.Stderr, "manager environment error: %v\n", err)
		os.Exit(1)
	}
	cfg, err := config.LoadManager(env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "manager configuration error: %v\n", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "manager configuration error: %v\n", err)
		os.Exit(1)
	}
	workerCfg, err := config.LoadWorker(env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "worker configuration error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s manager listening on %s\n", harness.ProjectName, cfg.HTTPAddress())
	mgr := manager.New(manager.Options{
		Source: node.HTTPSource{URL: cfg.VPNGateURL},
		ProbeRunner: node.ProbeFunc(func(_ context.Context, n node.Node) (node.ProbeResult, error) {
			latency := time.Duration(n.PingMS) * time.Millisecond
			if latency == 0 {
				latency = time.Millisecond
			}
			return node.ProbeResult{Latency: latency}, nil
		}),
		Launcher: workerdocker.NewLauncher(workerdocker.Config{
			Command:         cfg.DockerCommand,
			Image:           cfg.DockerWorkerImage,
			ContainerPrefix: cfg.DockerContainerPrefix,
			WorkerProxyPort: workerCfg.ProxyPort,
			WorkerTunDevice: workerCfg.TunDevice,
			OpenVPNCommand:  workerCfg.OpenVPNCommand,
			OpenVPNAuthFile: workerCfg.OpenVPNAuthFile,
			Network:         cfg.DockerNetwork,
		}, nil),
		ProxyBase: cfg.ProxyBasePort,
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
