package docker

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/juanlou1217/exitfleet/internal/manager"
	"github.com/juanlou1217/exitfleet/internal/node"
)

func TestStartWorkerRunsDockerContainer(t *testing.T) {
	runner := &fakeRunner{output: "container-123\n"}
	launcher := NewLauncher(Config{
		Command:                     "docker",
		Image:                       "exitfleet-worker:test",
		ContainerPrefix:             "exitfleet-worker",
		WorkerProxyPort:             7928,
		WorkerTunDevice:             "tun0",
		OpenVPNCommand:              "/usr/sbin/openvpn",
		OpenVPNAuthFile:             "/run/exitfleet/auth.txt",
		WorkerHealthHost:            "0.0.0.0",
		WorkerHealthPort:            8790,
		WorkerHealthURLs:            []string{"https://one.test", "https://two.test"},
		WorkerHealthFailThreshold:   5,
		WorkerHealthIntervalSeconds: 11,
		WorkerHealthTimeoutSeconds:  7,
		WorkerProxyOutboundIP:       "10.8.0.2",
		WorkerSOCKSMaxConnections:   77,
		Network:                     "exitfleet-net",
	}, runner)

	record, err := launcher.StartWorker(context.Background(), manager.LaunchRequest{
		Node:      node.Node{ID: "node-abc", Country: "Japan"},
		ProxyPort: 8028,
	})
	if err != nil {
		t.Fatalf("StartWorker() error = %v", err)
	}
	if record.ID != "container-123" {
		t.Fatalf("record.ID = %q, want container-123", record.ID)
	}
	if record.ProxyPort != 8028 || record.State != "starting" {
		t.Fatalf("record = %#v, want proxy port and starting state", record)
	}

	wantArgs := []string{
		"run", "-d",
		"--name", "exitfleet-worker-node-abc-8028",
		"--label", "exitfleet.managed=true",
		"--label", "exitfleet.node_id=node-abc",
		"--cap-add", "NET_ADMIN",
		"--device", "/dev/net/tun",
		"-p", "8028:7928",
		"-e", "EXITFLEET_WORKER_PROXY_HOST=0.0.0.0",
		"-e", "EXITFLEET_WORKER_PROXY_PORT=7928",
		"-e", "EXITFLEET_WORKER_TUN_DEVICE=tun0",
		"-e", "EXITFLEET_OPENVPN_CMD=/usr/sbin/openvpn",
		"-e", "EXITFLEET_OPENVPN_AUTH_FILE=/run/exitfleet/auth.txt",
		"-e", "EXITFLEET_WORKER_HEALTH_HOST=0.0.0.0",
		"-e", "EXITFLEET_WORKER_HEALTH_PORT=8790",
		"-e", "EXITFLEET_WORKER_HEALTH_URLS=https://one.test,https://two.test",
		"-e", "EXITFLEET_WORKER_HEALTH_FAIL_THRESHOLD=5",
		"-e", "EXITFLEET_WORKER_HEALTH_INTERVAL_SECONDS=11",
		"-e", "EXITFLEET_WORKER_HEALTH_TIMEOUT_SECONDS=7",
		"-e", "EXITFLEET_WORKER_PROXY_OUTBOUND_IP=10.8.0.2",
		"-e", "EXITFLEET_WORKER_SOCKS_MAX_CONNECTIONS=77",
		"--network", "exitfleet-net",
		"exitfleet-worker:test",
	}
	if runner.name != "docker" || !reflect.DeepEqual(runner.args, wantArgs) {
		t.Fatalf("command = %s %#v, want docker %#v", runner.name, runner.args, wantArgs)
	}
}

func TestStartWorkerFallsBackToContainerNameWhenOutputIsEmpty(t *testing.T) {
	runner := &fakeRunner{}
	launcher := NewLauncher(DefaultConfig(), runner)

	record, err := launcher.StartWorker(context.Background(), manager.LaunchRequest{
		Node:      node.Node{ID: "node-abc"},
		ProxyPort: 8028,
	})
	if err != nil {
		t.Fatalf("StartWorker() error = %v", err)
	}
	if record.ID != "exitfleet-worker-node-abc-8028" {
		t.Fatalf("record.ID = %q, want container name", record.ID)
	}
}

func TestStartWorkerPropagatesDockerError(t *testing.T) {
	runner := &fakeRunner{err: errors.New("docker not found")}
	launcher := NewLauncher(DefaultConfig(), runner)

	if _, err := launcher.StartWorker(context.Background(), manager.LaunchRequest{
		Node:      node.Node{ID: "node-abc"},
		ProxyPort: 8028,
	}); err == nil {
		t.Fatal("StartWorker() error = nil, want docker error")
	}
}

func TestStopWorkerRemovesContainer(t *testing.T) {
	runner := &fakeRunner{}
	launcher := NewLauncher(DefaultConfig(), runner)

	if err := launcher.StopWorker(context.Background(), "container-123"); err != nil {
		t.Fatalf("StopWorker() error = %v", err)
	}
	wantArgs := []string{"rm", "-f", "container-123"}
	if runner.name != "docker" || !reflect.DeepEqual(runner.args, wantArgs) {
		t.Fatalf("command = %s %#v, want docker %#v", runner.name, runner.args, wantArgs)
	}
}

type fakeRunner struct {
	name   string
	args   []string
	output string
	err    error
}

func (r *fakeRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	r.name = name
	r.args = append([]string(nil), args...)
	return r.output, r.err
}
