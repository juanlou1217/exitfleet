package config

import (
	"os"
	"path/filepath"
	"testing"
)

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

func TestLoadEnvFileParsesDotEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := []byte("EXITFLEET_MANAGER_HOST=127.0.0.1\nEXITFLEET_MANAGER_PORT=9000\n# comment\nEXITFLEET_VPNGATE_URL=https://example.test/api\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	env, err := LoadEnvFile(path)
	if err != nil {
		t.Fatalf("LoadEnvFile() error = %v", err)
	}
	if env["EXITFLEET_MANAGER_HOST"] != "127.0.0.1" {
		t.Fatalf("manager host = %q", env["EXITFLEET_MANAGER_HOST"])
	}
	if env["EXITFLEET_VPNGATE_URL"] != "https://example.test/api" {
		t.Fatalf("vpngate url = %q", env["EXITFLEET_VPNGATE_URL"])
	}
}

func TestLoadManagerUsesEnvFileValues(t *testing.T) {
	t.Setenv(EnvManagerHost, "127.0.0.1")
	t.Setenv(EnvManagerPort, "9000")
	t.Setenv(EnvVPNGateURL, "https://example.test/api")
	t.Setenv(EnvProxyBasePort, "8100")
	t.Setenv(EnvDockerCommand, "docker")
	t.Setenv(EnvDockerWorkerImage, "exitfleet-worker:test")
	t.Setenv(EnvDockerContainerPrefix, "ef-worker")
	t.Setenv(EnvDockerNetwork, "exitfleet-net")

	cfg, err := LoadManager(Env{
		"EXITFLEET_MANAGER_HOST":            "127.0.0.1",
		"EXITFLEET_MANAGER_PORT":            "9000",
		"EXITFLEET_VPNGATE_URL":             "https://example.test/api",
		"EXITFLEET_PROXY_BASE_PORT":         "8100",
		"EXITFLEET_DOCKER_CMD":              "docker",
		"EXITFLEET_DOCKER_WORKER_IMAGE":     "exitfleet-worker:test",
		"EXITFLEET_DOCKER_CONTAINER_PREFIX": "ef-worker",
		"EXITFLEET_DOCKER_NETWORK":          "exitfleet-net",
	})
	if err != nil {
		t.Fatalf("LoadManager() error = %v", err)
	}
	if cfg.HTTPAddress() != "127.0.0.1:9000" {
		t.Fatalf("HTTPAddress() = %q", cfg.HTTPAddress())
	}
	if cfg.VPNGateURL != "https://example.test/api" {
		t.Fatalf("VPNGateURL = %q", cfg.VPNGateURL)
	}
	if cfg.ProxyBasePort != 8100 {
		t.Fatalf("ProxyBasePort = %d", cfg.ProxyBasePort)
	}
	if cfg.DockerWorkerImage != "exitfleet-worker:test" {
		t.Fatalf("DockerWorkerImage = %q", cfg.DockerWorkerImage)
	}
	if cfg.DockerContainerPrefix != "ef-worker" {
		t.Fatalf("DockerContainerPrefix = %q", cfg.DockerContainerPrefix)
	}
	if cfg.DockerNetwork != "exitfleet-net" {
		t.Fatalf("DockerNetwork = %q", cfg.DockerNetwork)
	}
}

func TestLoadWorkerUsesEnvFileValues(t *testing.T) {
	t.Setenv(EnvWorkerProxyHost, "127.0.0.1")
	t.Setenv(EnvWorkerProxyPort, "8101")
	t.Setenv(EnvWorkerTunDevice, "tun7")
	t.Setenv(EnvOpenVPNCommand, "/usr/sbin/openvpn")
	t.Setenv(EnvOpenVPNAuthFile, "/run/exitfleet/auth.txt")

	cfg, err := LoadWorker(Env{
		"EXITFLEET_WORKER_PROXY_HOST": "127.0.0.1",
		"EXITFLEET_WORKER_PROXY_PORT": "8101",
		"EXITFLEET_WORKER_TUN_DEVICE": "tun7",
		"EXITFLEET_OPENVPN_CMD":       "/usr/sbin/openvpn",
		"EXITFLEET_OPENVPN_AUTH_FILE": "/run/exitfleet/auth.txt",
	})
	if err != nil {
		t.Fatalf("LoadWorker() error = %v", err)
	}
	if cfg.ProxyAddress() != "127.0.0.1:8101" {
		t.Fatalf("ProxyAddress() = %q", cfg.ProxyAddress())
	}
	if cfg.TunDevice != "tun7" {
		t.Fatalf("TunDevice = %q", cfg.TunDevice)
	}
	if cfg.OpenVPNCommand != "/usr/sbin/openvpn" {
		t.Fatalf("OpenVPNCommand = %q", cfg.OpenVPNCommand)
	}
	if cfg.OpenVPNAuthFile != "/run/exitfleet/auth.txt" {
		t.Fatalf("OpenVPNAuthFile = %q", cfg.OpenVPNAuthFile)
	}
}

func TestProcessEnvOverridesEnvFile(t *testing.T) {
	t.Setenv("EXITFLEET_MANAGER_PORT", "9100")

	cfg, err := LoadManager(Env{"EXITFLEET_MANAGER_PORT": "9000"})
	if err != nil {
		t.Fatalf("LoadManager() error = %v", err)
	}
	if cfg.Port != 9100 {
		t.Fatalf("Port = %d, want process env override 9100", cfg.Port)
	}
}

func TestLoadManagerRejectsInvalidPort(t *testing.T) {
	t.Setenv(EnvManagerPort, "bad")

	if _, err := LoadManager(Env{}); err == nil {
		t.Fatal("LoadManager() error = nil, want invalid port error")
	}
}
