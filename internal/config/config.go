package config

import (
	"fmt"
	"net"
)

type Manager struct {
	Host string
	Port int
}

type Worker struct {
	ProxyHost string
	ProxyPort int
	TunDevice string
}

func DefaultManager() Manager {
	return Manager{
		Host: "0.0.0.0",
		Port: 8787,
	}
}

func (cfg Manager) HTTPAddress() string {
	return net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
}

func (cfg Manager) Validate() error {
	if cfg.Host == "" {
		return fmt.Errorf("host is required")
	}
	return validatePort("port", cfg.Port)
}

func DefaultWorker() Worker {
	return Worker{
		ProxyHost: "0.0.0.0",
		ProxyPort: 7928,
		TunDevice: "tun0",
	}
}

func (cfg Worker) ProxyAddress() string {
	return net.JoinHostPort(cfg.ProxyHost, fmt.Sprintf("%d", cfg.ProxyPort))
}

func (cfg Worker) Validate() error {
	if cfg.ProxyHost == "" {
		return fmt.Errorf("proxy host is required")
	}
	if cfg.TunDevice == "" {
		return fmt.Errorf("tun device is required")
	}
	return validatePort("proxy port", cfg.ProxyPort)
}

func validatePort(name string, port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("%s must be between 1 and 65535", name)
	}
	return nil
}
