package docker

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/juanlou1217/exitfleet/internal/manager"
)

type Config struct {
	Command                     string
	Image                       string
	ContainerPrefix             string
	WorkerProxyPort             int
	WorkerProxyOutboundIP       string
	WorkerSOCKSMaxConnections   int
	WorkerHealthHost            string
	WorkerHealthPort            int
	WorkerHealthURLs            []string
	WorkerHealthFailThreshold   int
	WorkerHealthIntervalSeconds int
	WorkerHealthTimeoutSeconds  int
	WorkerTunDevice             string
	OpenVPNCommand              string
	OpenVPNAuthFile             string
	Network                     string
}

type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

type Launcher struct {
	config Config
	runner CommandRunner
}

func DefaultConfig() Config {
	return Config{
		Command:                     "docker",
		Image:                       "exitfleet-worker:local",
		ContainerPrefix:             "exitfleet-worker",
		WorkerProxyPort:             7928,
		WorkerSOCKSMaxConnections:   200,
		WorkerHealthHost:            "0.0.0.0",
		WorkerHealthPort:            8790,
		WorkerHealthFailThreshold:   3,
		WorkerHealthIntervalSeconds: 10,
		WorkerHealthTimeoutSeconds:  8,
		WorkerTunDevice:             "tun0",
		OpenVPNCommand:              "openvpn",
	}
}

func NewLauncher(config Config, runner CommandRunner) *Launcher {
	defaults := DefaultConfig()
	if config.Command == "" {
		config.Command = defaults.Command
	}
	if config.Image == "" {
		config.Image = defaults.Image
	}
	if config.ContainerPrefix == "" {
		config.ContainerPrefix = defaults.ContainerPrefix
	}
	if config.WorkerProxyPort == 0 {
		config.WorkerProxyPort = defaults.WorkerProxyPort
	}
	if config.WorkerSOCKSMaxConnections == 0 {
		config.WorkerSOCKSMaxConnections = defaults.WorkerSOCKSMaxConnections
	}
	if config.WorkerHealthHost == "" {
		config.WorkerHealthHost = defaults.WorkerHealthHost
	}
	if config.WorkerHealthPort == 0 {
		config.WorkerHealthPort = defaults.WorkerHealthPort
	}
	if config.WorkerHealthFailThreshold == 0 {
		config.WorkerHealthFailThreshold = defaults.WorkerHealthFailThreshold
	}
	if config.WorkerHealthIntervalSeconds == 0 {
		config.WorkerHealthIntervalSeconds = defaults.WorkerHealthIntervalSeconds
	}
	if config.WorkerHealthTimeoutSeconds == 0 {
		config.WorkerHealthTimeoutSeconds = defaults.WorkerHealthTimeoutSeconds
	}
	if config.WorkerTunDevice == "" {
		config.WorkerTunDevice = defaults.WorkerTunDevice
	}
	if config.OpenVPNCommand == "" {
		config.OpenVPNCommand = defaults.OpenVPNCommand
	}
	if runner == nil {
		runner = ExecRunner{}
	}
	return &Launcher{config: config, runner: runner}
}

func (l *Launcher) StartWorker(ctx context.Context, request manager.LaunchRequest) (manager.WorkerRecord, error) {
	containerName := l.containerName(request)
	args := []string{
		"run", "-d",
		"--name", containerName,
		"--label", "exitfleet.managed=true",
		"--label", "exitfleet.node_id=" + request.Node.ID,
		"--cap-add", "NET_ADMIN",
		"--device", "/dev/net/tun",
		"-p", fmt.Sprintf("%d:%d", request.ProxyPort, l.config.WorkerProxyPort),
		"-e", "EXITFLEET_WORKER_PROXY_HOST=0.0.0.0",
		"-e", fmt.Sprintf("EXITFLEET_WORKER_PROXY_PORT=%d", l.config.WorkerProxyPort),
		"-e", "EXITFLEET_WORKER_TUN_DEVICE=" + l.config.WorkerTunDevice,
		"-e", "EXITFLEET_OPENVPN_CMD=" + l.config.OpenVPNCommand,
	}
	if l.config.OpenVPNAuthFile != "" {
		args = append(args, "-e", "EXITFLEET_OPENVPN_AUTH_FILE="+l.config.OpenVPNAuthFile)
	}
	args = append(args,
		"-e", "EXITFLEET_WORKER_HEALTH_HOST="+l.config.WorkerHealthHost,
		"-e", fmt.Sprintf("EXITFLEET_WORKER_HEALTH_PORT=%d", l.config.WorkerHealthPort),
		"-e", "EXITFLEET_WORKER_HEALTH_URLS="+strings.Join(l.config.WorkerHealthURLs, ","),
		"-e", fmt.Sprintf("EXITFLEET_WORKER_HEALTH_FAIL_THRESHOLD=%d", l.config.WorkerHealthFailThreshold),
		"-e", fmt.Sprintf("EXITFLEET_WORKER_HEALTH_INTERVAL_SECONDS=%d", l.config.WorkerHealthIntervalSeconds),
		"-e", fmt.Sprintf("EXITFLEET_WORKER_HEALTH_TIMEOUT_SECONDS=%d", l.config.WorkerHealthTimeoutSeconds),
		"-e", "EXITFLEET_WORKER_PROXY_OUTBOUND_IP="+l.config.WorkerProxyOutboundIP,
		"-e", fmt.Sprintf("EXITFLEET_WORKER_SOCKS_MAX_CONNECTIONS=%d", l.config.WorkerSOCKSMaxConnections),
	)
	if l.config.Network != "" {
		args = append(args, "--network", l.config.Network)
	}
	args = append(args, l.config.Image)

	output, err := l.runner.Run(ctx, l.config.Command, args...)
	if err != nil {
		return manager.WorkerRecord{}, err
	}
	id := strings.TrimSpace(output)
	if id == "" {
		id = containerName
	}
	return manager.WorkerRecord{
		ID:        id,
		ProxyPort: request.ProxyPort,
		State:     "starting",
	}, nil
}

func (l *Launcher) StopWorker(ctx context.Context, workerID string) error {
	_, err := l.runner.Run(ctx, l.config.Command, "rm", "-f", workerID)
	return err
}

func (l *Launcher) containerName(request manager.LaunchRequest) string {
	return sanitizeName(fmt.Sprintf("%s-%s-%d", l.config.ContainerPrefix, request.Node.ID, request.ProxyPort))
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("%s %s failed: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

var invalidNameChars = regexp.MustCompile(`[^a-zA-Z0-9_.-]+`)

func sanitizeName(value string) string {
	value = invalidNameChars.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "exitfleet-worker"
	}
	return value
}
