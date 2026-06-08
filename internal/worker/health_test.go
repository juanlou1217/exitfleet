package worker

import (
	"context"
	"errors"
	"testing"
)

func TestHealthCheckerTreatsAnySuccessfulURLAsHealthy(t *testing.T) {
	probe := &fakeHealthProbe{results: map[string]error{
		"https://one.test": errors.New("down"),
		"https://two.test": nil,
	}}
	checker := NewHealthChecker(HealthOptions{
		URLs:          []string{"https://one.test", "https://two.test"},
		FailThreshold: 2,
		Probe:         probe,
	})

	result := checker.CheckOnce(context.Background())
	if !result.Healthy {
		t.Fatalf("Healthy = false, want true")
	}
	if result.ConsecutiveFailures != 0 {
		t.Fatalf("ConsecutiveFailures = %d, want 0", result.ConsecutiveFailures)
	}
	if !checker.Ready() {
		t.Fatal("Ready() = false, want true after successful URL")
	}
}

func TestHealthCheckerNotReadyAfterFailureThreshold(t *testing.T) {
	checker := NewHealthChecker(HealthOptions{
		URLs:          []string{"https://one.test"},
		FailThreshold: 2,
		Probe:         &fakeHealthProbe{err: errors.New("down")},
	})

	first := checker.CheckOnce(context.Background())
	if first.ConsecutiveFailures != 1 || !checker.Ready() {
		t.Fatalf("after first failure = %#v ready=%v, want 1 and ready", first, checker.Ready())
	}
	second := checker.CheckOnce(context.Background())
	if second.ConsecutiveFailures != 2 || checker.Ready() {
		t.Fatalf("after second failure = %#v ready=%v, want threshold and not ready", second, checker.Ready())
	}
}

func TestWorkerCheckHealthUpdatesStatus(t *testing.T) {
	w := New(Options{
		Config: OpenVPNConfig{ConfigFile: "/tmp/node.ovpn"},
		Runner: RunnerFunc(func(ctx context.Context, command []string) (RunResult, error) {
			return RunResult{Ready: true}, nil
		}),
		HealthChecker: NewHealthChecker(HealthOptions{
			URLs:          []string{"https://one.test"},
			FailThreshold: 1,
			Probe:         &fakeHealthProbe{err: errors.New("down")},
		}),
	})
	if err := w.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	result := w.CheckHealth(context.Background())
	if result.Healthy {
		t.Fatal("Healthy = true, want false")
	}
	status := w.Status()
	if status.State != StateFailed {
		t.Fatalf("state = %q, want failed after health threshold", status.State)
	}
	if status.HealthFailures != 1 {
		t.Fatalf("HealthFailures = %d, want 1", status.HealthFailures)
	}
}

type fakeHealthProbe struct {
	results map[string]error
	err     error
}

func (p *fakeHealthProbe) Check(ctx context.Context, url string) error {
	if p.results != nil {
		if err, ok := p.results[url]; ok {
			return err
		}
	}
	return p.err
}
