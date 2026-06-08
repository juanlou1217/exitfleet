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
	Command         string
	Image           string
	ContainerPrefix string
	WorkerProxyPort int
	WorkerTunDevice string
	OpenVPNCommand  string
	OpenVPNAuthFile string
	Network         string
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
		Command:         "docker",
		Image:           "exitfleet-worker:local",
		ContainerPrefix: "exitfleet-worker",
		WorkerProxyPort: 7928,
		WorkerTunDevice: "tun0",
		OpenVPNCommand:  "openvpn",
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
