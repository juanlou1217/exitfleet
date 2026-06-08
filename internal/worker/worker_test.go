package worker

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestBuildOpenVPNCommandUsesTunAuthAndRouteNoPull(t *testing.T) {
	cmd, err := BuildOpenVPNCommand(OpenVPNConfig{
		Command:     "openvpn",
		ConfigFile:  "/tmp/node.ovpn",
		AuthFile:    "/tmp/auth.txt",
		TunDevice:   "tun0",
		RouteNoPull: true,
	})
	if err != nil {
		t.Fatalf("BuildOpenVPNCommand() error = %v", err)
	}

	want := []string{
		"openvpn",
		"--config", "/tmp/node.ovpn",
		"--dev", "tun0",
		"--dev-type", "tun",
		"--auth-user-pass", "/tmp/auth.txt",
		"--route-nopull",
	}
	if !reflect.DeepEqual(cmd, want) {
		t.Fatalf("command = %#v, want %#v", cmd, want)
	}
}

func TestBuildOpenVPNCommandRejectsMissingConfigFile(t *testing.T) {
	if _, err := BuildOpenVPNCommand(OpenVPNConfig{}); err == nil {
		t.Fatal("BuildOpenVPNCommand() error = nil, want missing config error")
	}
}

func TestWorkerStartTransitionsToReady(t *testing.T) {
	runner := RunnerFunc(func(ctx context.Context, command []string) (RunResult, error) {
		return RunResult{Ready: true, Message: "Initialization Sequence Completed"}, nil
	})
	w := New(Options{
		Config: OpenVPNConfig{ConfigFile: "/tmp/node.ovpn", AuthFile: "/tmp/auth.txt"},
		Runner: runner,
	})

	if err := w.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if got := w.Status(); got.State != StateReady {
		t.Fatalf("state = %q, want %q", got.State, StateReady)
	}
}

func TestWorkerStartTransitionsToFailed(t *testing.T) {
	runner := RunnerFunc(func(ctx context.Context, command []string) (RunResult, error) {
		return RunResult{}, errors.New("openvpn missing")
	})
	w := New(Options{
		Config: OpenVPNConfig{ConfigFile: "/tmp/node.ovpn", AuthFile: "/tmp/auth.txt"},
		Runner: runner,
	})

	if err := w.Start(context.Background()); err == nil {
		t.Fatal("Start() error = nil, want runner error")
	}
	if got := w.Status(); got.State != StateFailed || got.LastError != "openvpn missing" {
		t.Fatalf("status = %#v, want failed with error", got)
	}
}

func TestWorkerStopIsIdempotent(t *testing.T) {
	w := New(Options{
		Config: OpenVPNConfig{ConfigFile: "/tmp/node.ovpn"},
		Runner: RunnerFunc(func(ctx context.Context, command []string) (RunResult, error) {
			return RunResult{Ready: true}, nil
		}),
	})

	if err := w.Stop(context.Background()); err != nil {
		t.Fatalf("first Stop() error = %v", err)
	}
	if err := w.Stop(context.Background()); err != nil {
		t.Fatalf("second Stop() error = %v", err)
	}
	if got := w.Status(); got.State != StateStopped {
		t.Fatalf("state = %q, want %q", got.State, StateStopped)
	}
}

func TestHealthAndReadyHandlersReflectState(t *testing.T) {
	w := New(Options{
		Config: OpenVPNConfig{ConfigFile: "/tmp/node.ovpn"},
		Runner: RunnerFunc(func(ctx context.Context, command []string) (RunResult, error) {
			return RunResult{Ready: true}, nil
		}),
	})

	health := httptest.NewRecorder()
	w.HealthHandler().ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health status = %d, want 200", health.Code)
	}

	notReady := httptest.NewRecorder()
	w.ReadyHandler().ServeHTTP(notReady, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if notReady.Code != http.StatusServiceUnavailable {
		t.Fatalf("ready before start status = %d, want 503", notReady.Code)
	}

	if err := w.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	ready := httptest.NewRecorder()
	w.ReadyHandler().ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if ready.Code != http.StatusOK {
		t.Fatalf("ready after start status = %d, want 200", ready.Code)
	}
}
