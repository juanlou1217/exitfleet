package config

import "testing"

func TestDefaultManagerUsesManagementPort(t *testing.T) {
	cfg := DefaultManager()

	if err := cfg.Validate(); err != nil {
		t.Fatalf("DefaultManager().Validate() error = %v", err)
	}
	if got := cfg.HTTPAddress(); got != "0.0.0.0:8787" {
		t.Fatalf("DefaultManager().HTTPAddress() = %q, want %q", got, "0.0.0.0:8787")
	}
}

func TestDefaultWorkerUsesFirstProxyPortAndTun0(t *testing.T) {
	cfg := DefaultWorker()

	if err := cfg.Validate(); err != nil {
		t.Fatalf("DefaultWorker().Validate() error = %v", err)
	}
	if got := cfg.ProxyAddress(); got != "0.0.0.0:7928" {
		t.Fatalf("DefaultWorker().ProxyAddress() = %q, want %q", got, "0.0.0.0:7928")
	}
	if cfg.TunDevice != "tun0" {
		t.Fatalf("DefaultWorker().TunDevice = %q, want tun0", cfg.TunDevice)
	}
}

func TestWorkerRejectsMissingTunDevice(t *testing.T) {
	cfg := DefaultWorker()
	cfg.TunDevice = ""

	if err := cfg.Validate(); err == nil {
		t.Fatal("Worker.Validate() error = nil, want missing tun device error")
	}
}
