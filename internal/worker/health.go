package worker

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type HealthProbe interface {
	Check(ctx context.Context, url string) error
}

type HTTPHealthProbe struct {
	Client  *http.Client
	Timeout time.Duration
}

func (p HTTPHealthProbe) Check(ctx context.Context, url string) error {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("health url %s returned status %d", url, resp.StatusCode)
	}
	return nil
}

type HealthOptions struct {
	URLs          []string
	FailThreshold int
	Timeout       time.Duration
	Probe         HealthProbe
}

type HealthResult struct {
	Healthy             bool
	SuccessfulURL       string
	ConsecutiveFailures int
	LastError           string
}

type HealthChecker struct {
	mu               sync.RWMutex
	urls             []string
	failThreshold    int
	probe            HealthProbe
	consecutiveFails int
	lastError        string
}

func NewHealthChecker(options HealthOptions) *HealthChecker {
	failThreshold := options.FailThreshold
	if failThreshold <= 0 {
		failThreshold = 3
	}
	probe := options.Probe
	if probe == nil {
		probe = HTTPHealthProbe{Timeout: options.Timeout}
	}
	urls := make([]string, 0, len(options.URLs))
	for _, url := range options.URLs {
		if trimmed := strings.TrimSpace(url); trimmed != "" {
			urls = append(urls, trimmed)
		}
	}
	return &HealthChecker{urls: urls, failThreshold: failThreshold, probe: probe}
}

func (c *HealthChecker) CheckOnce(ctx context.Context) HealthResult {
	if c == nil {
		return HealthResult{Healthy: true}
	}
	if len(c.urls) == 0 {
		c.setResult(0, "")
		return HealthResult{Healthy: true}
	}

	var lastErr error
	for _, url := range c.urls {
		if err := c.probe.Check(ctx, url); err == nil {
			c.setResult(0, "")
			return HealthResult{Healthy: true, SuccessfulURL: url}
		} else {
			lastErr = err
		}
	}

	c.mu.Lock()
	c.consecutiveFails++
	c.lastError = lastErr.Error()
	result := HealthResult{
		Healthy:             false,
		ConsecutiveFailures: c.consecutiveFails,
		LastError:           c.lastError,
	}
	c.mu.Unlock()
	return result
}

func (c *HealthChecker) Ready() bool {
	if c == nil {
		return true
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.consecutiveFails < c.failThreshold
}

func (c *HealthChecker) setResult(consecutiveFailures int, lastError string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.consecutiveFails = consecutiveFailures
	c.lastError = lastError
}
