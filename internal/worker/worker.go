package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

const (
	StateStopped  = "stopped"
	StateStarting = "starting"
	StateReady    = "ready"
	StateFailed   = "failed"
)

type OpenVPNConfig struct {
	Command     string
	ConfigFile  string
	AuthFile    string
	TunDevice   string
	RouteNoPull bool
}

type RunResult struct {
	Ready   bool
	Message string
	LogTail []string
}

type Runner interface {
	Run(ctx context.Context, command []string) (RunResult, error)
}

type RunnerFunc func(ctx context.Context, command []string) (RunResult, error)

func (fn RunnerFunc) Run(ctx context.Context, command []string) (RunResult, error) {
	return fn(ctx, command)
}

type Status struct {
	State     string `json:"state"`
	TunDevice string `json:"tun_device"`
	LastError string `json:"last_error,omitempty"`
	Message   string `json:"message,omitempty"`
}

type Options struct {
	Config OpenVPNConfig
	Runner Runner
}

type Worker struct {
	mu     sync.RWMutex
	config OpenVPNConfig
	runner Runner
	status Status
}

func New(options Options) *Worker {
	cfg := options.Config
	if cfg.Command == "" {
		cfg.Command = "openvpn"
	}
	if cfg.TunDevice == "" {
		cfg.TunDevice = "tun0"
	}
	return &Worker{
		config: cfg,
		runner: options.Runner,
		status: Status{State: StateStopped, TunDevice: cfg.TunDevice},
	}
}

func BuildOpenVPNCommand(cfg OpenVPNConfig) ([]string, error) {
	if cfg.Command == "" {
		cfg.Command = "openvpn"
	}
	if cfg.TunDevice == "" {
		cfg.TunDevice = "tun0"
	}
	if strings.TrimSpace(cfg.ConfigFile) == "" {
		return nil, fmt.Errorf("openvpn config file is required")
	}

	command := []string{
		cfg.Command,
		"--config", cfg.ConfigFile,
		"--dev", cfg.TunDevice,
		"--dev-type", "tun",
	}
	if cfg.AuthFile != "" {
		command = append(command, "--auth-user-pass", cfg.AuthFile)
	}
	if cfg.RouteNoPull {
		command = append(command, "--route-nopull")
	}
	return command, nil
}

func (w *Worker) Start(ctx context.Context) error {
	w.setStatus(Status{State: StateStarting, TunDevice: w.config.TunDevice, Message: "starting openvpn"})

	command, err := BuildOpenVPNCommand(w.config)
	if err != nil {
		w.setStatus(Status{State: StateFailed, TunDevice: w.config.TunDevice, LastError: err.Error()})
		return err
	}
	if w.runner == nil {
		err := fmt.Errorf("worker runner is required")
		w.setStatus(Status{State: StateFailed, TunDevice: w.config.TunDevice, LastError: err.Error()})
		return err
	}

	result, err := w.runner.Run(ctx, command)
	if err != nil {
		w.setStatus(Status{State: StateFailed, TunDevice: w.config.TunDevice, LastError: err.Error()})
		return err
	}
	if result.Ready || logsContainReady(result.LogTail) {
		w.setStatus(Status{State: StateReady, TunDevice: w.config.TunDevice, Message: result.Message})
		return nil
	}
	err = fmt.Errorf("openvpn did not become ready")
	if result.Message != "" {
		err = errors.New(result.Message)
	}
	w.setStatus(Status{State: StateFailed, TunDevice: w.config.TunDevice, LastError: err.Error()})
	return err
}

func (w *Worker) Stop(ctx context.Context) error {
	w.setStatus(Status{State: StateStopped, TunDevice: w.config.TunDevice})
	return nil
}

func (w *Worker) Status() Status {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.status
}

func (w *Worker) HealthHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		writeJSON(rw, http.StatusOK, w.Status())
	})
}

func (w *Worker) ReadyHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		status := w.Status()
		if status.State != StateReady {
			writeJSON(rw, http.StatusServiceUnavailable, status)
			return
		}
		writeJSON(rw, http.StatusOK, status)
	})
}

func (w *Worker) setStatus(status Status) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.status = status
}

func logsContainReady(lines []string) bool {
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "initialization sequence completed") {
			return true
		}
	}
	return false
}

func writeJSON(rw http.ResponseWriter, status int, value any) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	_ = json.NewEncoder(rw).Encode(value)
}
