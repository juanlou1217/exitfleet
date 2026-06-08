package config

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Manager struct {
	Host                  string
	Port                  int
	VPNGateURL            string
	ProxyBasePort         int
	DockerCommand         string
	DockerWorkerImage     string
	DockerContainerPrefix string
	DockerNetwork         string
}

type Worker struct {
	ProxyHost             string
	ProxyPort             int
	ProxyOutboundIP       string
	SOCKSMaxConnections   int
	HealthHost            string
	HealthPort            int
	HealthURLs            []string
	HealthFailThreshold   int
	HealthIntervalSeconds int
	HealthTimeoutSeconds  int
	TunDevice             string
	OpenVPNCommand        string
	OpenVPNAuthFile       string
}

type Env map[string]string

const (
	DefaultVPNGateURL = "https://www.vpngate.net/api/iphone/"

	EnvManagerHost                 = "EXITFLEET_MANAGER_HOST"
	EnvManagerPort                 = "EXITFLEET_MANAGER_PORT"
	EnvVPNGateURL                  = "EXITFLEET_VPNGATE_URL"
	EnvProxyBasePort               = "EXITFLEET_PROXY_BASE_PORT"
	EnvDockerCommand               = "EXITFLEET_DOCKER_CMD"
	EnvDockerWorkerImage           = "EXITFLEET_DOCKER_WORKER_IMAGE"
	EnvDockerContainerPrefix       = "EXITFLEET_DOCKER_CONTAINER_PREFIX"
	EnvDockerNetwork               = "EXITFLEET_DOCKER_NETWORK"
	EnvWorkerProxyHost             = "EXITFLEET_WORKER_PROXY_HOST"
	EnvWorkerProxyPort             = "EXITFLEET_WORKER_PROXY_PORT"
	EnvWorkerProxyOutboundIP       = "EXITFLEET_WORKER_PROXY_OUTBOUND_IP"
	EnvWorkerSOCKSMaxConnections   = "EXITFLEET_WORKER_SOCKS_MAX_CONNECTIONS"
	EnvWorkerHealthHost            = "EXITFLEET_WORKER_HEALTH_HOST"
	EnvWorkerHealthPort            = "EXITFLEET_WORKER_HEALTH_PORT"
	EnvWorkerHealthURLs            = "EXITFLEET_WORKER_HEALTH_URLS"
	EnvWorkerHealthFailThreshold   = "EXITFLEET_WORKER_HEALTH_FAIL_THRESHOLD"
	EnvWorkerHealthIntervalSeconds = "EXITFLEET_WORKER_HEALTH_INTERVAL_SECONDS"
	EnvWorkerHealthTimeoutSeconds  = "EXITFLEET_WORKER_HEALTH_TIMEOUT_SECONDS"
	EnvWorkerTunDevice             = "EXITFLEET_WORKER_TUN_DEVICE"
	EnvOpenVPNCommand              = "EXITFLEET_OPENVPN_CMD"
	EnvOpenVPNAuthFile             = "EXITFLEET_OPENVPN_AUTH_FILE"
)

func DefaultManager() Manager {
	return Manager{
		Host:                  "0.0.0.0",
		Port:                  8787,
		VPNGateURL:            DefaultVPNGateURL,
		ProxyBasePort:         7928,
		DockerCommand:         "docker",
		DockerWorkerImage:     "exitfleet-worker:local",
		DockerContainerPrefix: "exitfleet-worker",
	}
}

func (cfg Manager) HTTPAddress() string {
	return net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
}

func (cfg Manager) Validate() error {
	if cfg.Host == "" {
		return fmt.Errorf("host is required")
	}
	if cfg.VPNGateURL == "" {
		return fmt.Errorf("vpngate url is required")
	}
	if cfg.DockerCommand == "" {
		return fmt.Errorf("docker command is required")
	}
	if cfg.DockerWorkerImage == "" {
		return fmt.Errorf("docker worker image is required")
	}
	if cfg.DockerContainerPrefix == "" {
		return fmt.Errorf("docker container prefix is required")
	}
	if err := validatePort("port", cfg.Port); err != nil {
		return err
	}
	return validatePort("proxy base port", cfg.ProxyBasePort)
}

func DefaultWorker() Worker {
	return Worker{
		ProxyHost:             "0.0.0.0",
		ProxyPort:             7928,
		SOCKSMaxConnections:   200,
		HealthHost:            "0.0.0.0",
		HealthPort:            8790,
		HealthFailThreshold:   3,
		HealthIntervalSeconds: 10,
		HealthTimeoutSeconds:  8,
		TunDevice:             "tun0",
		OpenVPNCommand:        "openvpn",
	}
}

func (cfg Worker) ProxyAddress() string {
	return net.JoinHostPort(cfg.ProxyHost, fmt.Sprintf("%d", cfg.ProxyPort))
}

func (cfg Worker) HealthAddress() string {
	return net.JoinHostPort(cfg.HealthHost, fmt.Sprintf("%d", cfg.HealthPort))
}

func (cfg Worker) Validate() error {
	if cfg.ProxyHost == "" {
		return fmt.Errorf("proxy host is required")
	}
	if cfg.HealthHost == "" {
		return fmt.Errorf("health host is required")
	}
	if cfg.TunDevice == "" {
		return fmt.Errorf("tun device is required")
	}
	if cfg.OpenVPNCommand == "" {
		return fmt.Errorf("openvpn command is required")
	}
	if cfg.SOCKSMaxConnections < 1 {
		return fmt.Errorf("socks max connections must be positive")
	}
	if cfg.HealthFailThreshold < 1 {
		return fmt.Errorf("health fail threshold must be positive")
	}
	if cfg.HealthIntervalSeconds < 1 {
		return fmt.Errorf("health interval seconds must be positive")
	}
	if cfg.HealthTimeoutSeconds < 1 {
		return fmt.Errorf("health timeout seconds must be positive")
	}
	if err := validatePort("proxy port", cfg.ProxyPort); err != nil {
		return err
	}
	return validatePort("health port", cfg.HealthPort)
}

func LoadEnvFile(path string) (Env, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Env{}, nil
		}
		return nil, err
	}
	defer file.Close()

	env := Env{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("invalid env line %q", line)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("empty env key")
		}
		env[key] = trimEnvValue(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return env, nil
}

func LoadManager(env Env) (Manager, error) {
	cfg := DefaultManager()
	cfg.Host = envString(env, EnvManagerHost, cfg.Host)
	cfg.VPNGateURL = envString(env, EnvVPNGateURL, cfg.VPNGateURL)
	cfg.DockerCommand = envString(env, EnvDockerCommand, cfg.DockerCommand)
	cfg.DockerWorkerImage = envString(env, EnvDockerWorkerImage, cfg.DockerWorkerImage)
	cfg.DockerContainerPrefix = envString(env, EnvDockerContainerPrefix, cfg.DockerContainerPrefix)
	cfg.DockerNetwork = envString(env, EnvDockerNetwork, cfg.DockerNetwork)

	var err error
	cfg.Port, err = envInt(env, EnvManagerPort, cfg.Port)
	if err != nil {
		return Manager{}, err
	}
	cfg.ProxyBasePort, err = envInt(env, EnvProxyBasePort, cfg.ProxyBasePort)
	if err != nil {
		return Manager{}, err
	}
	return cfg, cfg.Validate()
}

func LoadWorker(env Env) (Worker, error) {
	cfg := DefaultWorker()
	cfg.ProxyHost = envString(env, EnvWorkerProxyHost, cfg.ProxyHost)
	cfg.ProxyOutboundIP = envString(env, EnvWorkerProxyOutboundIP, cfg.ProxyOutboundIP)
	cfg.HealthHost = envString(env, EnvWorkerHealthHost, cfg.HealthHost)
	cfg.HealthURLs = envStringList(env, EnvWorkerHealthURLs, cfg.HealthURLs)
	cfg.TunDevice = envString(env, EnvWorkerTunDevice, cfg.TunDevice)
	cfg.OpenVPNCommand = envString(env, EnvOpenVPNCommand, cfg.OpenVPNCommand)
	cfg.OpenVPNAuthFile = envString(env, EnvOpenVPNAuthFile, cfg.OpenVPNAuthFile)

	var err error
	cfg.ProxyPort, err = envInt(env, EnvWorkerProxyPort, cfg.ProxyPort)
	if err != nil {
		return Worker{}, err
	}
	cfg.SOCKSMaxConnections, err = envInt(env, EnvWorkerSOCKSMaxConnections, cfg.SOCKSMaxConnections)
	if err != nil {
		return Worker{}, err
	}
	cfg.HealthPort, err = envInt(env, EnvWorkerHealthPort, cfg.HealthPort)
	if err != nil {
		return Worker{}, err
	}
	cfg.HealthFailThreshold, err = envInt(env, EnvWorkerHealthFailThreshold, cfg.HealthFailThreshold)
	if err != nil {
		return Worker{}, err
	}
	cfg.HealthIntervalSeconds, err = envInt(env, EnvWorkerHealthIntervalSeconds, cfg.HealthIntervalSeconds)
	if err != nil {
		return Worker{}, err
	}
	cfg.HealthTimeoutSeconds, err = envInt(env, EnvWorkerHealthTimeoutSeconds, cfg.HealthTimeoutSeconds)
	if err != nil {
		return Worker{}, err
	}
	return cfg, cfg.Validate()
}

func validatePort(name string, port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("%s must be between 1 and 65535", name)
	}
	return nil
}

func envString(env Env, key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value)
	}
	if value, ok := env[key]; ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func envInt(env Env, key string, fallback int) (int, error) {
	raw := envString(env, key, "")
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}
	return value, nil
}

func envStringList(env Env, key string, fallback []string) []string {
	raw := envString(env, key, "")
	if raw == "" {
		return fallback
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func trimEnvValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}
