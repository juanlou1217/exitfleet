package node

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	ProbeUnknown     = "unknown"
	ProbeAvailable   = "available"
	ProbeUnavailable = "unavailable"
)

type Node struct {
	ID          string `json:"id"`
	HostName    string `json:"host_name"`
	IP          string `json:"ip"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Score       int    `json:"score"`
	PingMS      int    `json:"ping_ms"`
	Speed       int64  `json:"speed"`
	ConfigText  string `json:"config_text,omitempty"`
	ProbeStatus string `json:"probe_status"`
	LatencyMS   int    `json:"latency_ms"`
	LastError   string `json:"last_error,omitempty"`
}

type Source interface {
	Fetch(ctx context.Context) (string, error)
}

type SourceFunc func(ctx context.Context) (string, error)

func (fn SourceFunc) Fetch(ctx context.Context) (string, error) {
	return fn(ctx)
}

type HTTPSource struct {
	URL    string
	Client *http.Client
}

func (s HTTPSource) Fetch(ctx context.Context) (string, error) {
	if strings.TrimSpace(s.URL) == "" {
		return "", fmt.Errorf("source URL is required")
	}
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("source returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

type ProbeResult struct {
	Latency time.Duration
}

type ProbeRunner interface {
	Probe(ctx context.Context, n Node) (ProbeResult, error)
}

type ProbeFunc func(ctx context.Context, n Node) (ProbeResult, error)

func (fn ProbeFunc) Probe(ctx context.Context, n Node) (ProbeResult, error) {
	return fn(ctx, n)
}

func FetchCandidates(ctx context.Context, source Source) ([]Node, error) {
	if source == nil {
		return nil, fmt.Errorf("node source is required")
	}
	text, err := source.Fetch(ctx)
	if err != nil {
		return nil, err
	}
	return ParseVPNGateCSV(text)
}

func ParseVPNGateCSV(text string) ([]Node, error) {
	reader := csv.NewReader(strings.NewReader(text))
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	var header []string
	var nodes []Node
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) == 0 {
			continue
		}
		first := strings.TrimSpace(record[0])
		if first == "" || first == "*" || strings.HasPrefix(first, "*") {
			continue
		}
		if strings.HasPrefix(first, "#") {
			header = append([]string(nil), record...)
			header[0] = strings.TrimPrefix(header[0], "#")
			continue
		}
		if header == nil {
			return nil, fmt.Errorf("vpngate csv header is required")
		}
		row := rowMap(header, record)
		encodedConfig := row["OpenVPN_ConfigData_Base64"]
		if encodedConfig == "" {
			return nil, fmt.Errorf("OpenVPN_ConfigData_Base64 column is required")
		}
		config, err := DecodeOpenVPNConfig(encodedConfig)
		if err != nil {
			return nil, err
		}
		n := Node{
			HostName:    row["HostName"],
			IP:          row["IP"],
			Country:     row["CountryLong"],
			CountryCode: row["CountryShort"],
			Score:       atoi(row["Score"]),
			PingMS:      atoi(row["Ping"]),
			Speed:       atoi64(row["Speed"]),
			ConfigText:  config,
			ProbeStatus: ProbeUnknown,
		}
		n.ID = stableID(n)
		nodes = append(nodes, n)
	}
	if header == nil {
		return nil, fmt.Errorf("vpngate csv header is required")
	}
	return nodes, nil
}

func DecodeOpenVPNConfig(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func ProbeNode(ctx context.Context, n Node, runner ProbeRunner) Node {
	if runner == nil {
		n.ProbeStatus = ProbeUnavailable
		n.LastError = "probe runner is required"
		return n
	}
	result, err := runner.Probe(ctx, n)
	if err != nil {
		n.ProbeStatus = ProbeUnavailable
		n.LastError = err.Error()
		n.LatencyMS = 0
		return n
	}
	n.ProbeStatus = ProbeAvailable
	n.LastError = ""
	n.LatencyMS = int(result.Latency / time.Millisecond)
	return n
}

func rowMap(header []string, record []string) map[string]string {
	row := make(map[string]string, len(header))
	for i, key := range header {
		if i < len(record) {
			row[strings.TrimSpace(key)] = strings.TrimSpace(record[i])
		}
	}
	return row
}

func stableID(n Node) string {
	sum := sha1.Sum([]byte(n.HostName + "|" + n.IP + "|" + n.CountryCode))
	return hex.EncodeToString(sum[:8])
}

func atoi(value string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(value))
	return n
}

func atoi64(value string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return n
}
