package node

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestParseVPNGateCSVDecodesConfig(t *testing.T) {
	text := strings.Join([]string{
		"*vpn_servers",
		"#HostName,IP,Score,Ping,Speed,CountryLong,CountryShort,NumVpnSessions,Uptime,TotalUsers,TotalTraffic,LogType,Operator,Message,OpenVPN_ConfigData_Base64",
		"host-a,203.0.113.10,900,42,1000000,Japan,JP,1,2,3,4,2,op,msg,Y2xpZW50Cg==",
		"*",
	}, "\n")

	nodes, err := ParseVPNGateCSV(text)
	if err != nil {
		t.Fatalf("ParseVPNGateCSV() error = %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("len(nodes) = %d, want 1", len(nodes))
	}
	got := nodes[0]
	if got.ID == "" {
		t.Fatal("node ID must not be empty")
	}
	if got.HostName != "host-a" || got.IP != "203.0.113.10" || got.Country == "" {
		t.Fatalf("unexpected node fields: %#v", got)
	}
	if got.Score != 900 || got.PingMS != 42 || got.Speed != 1000000 {
		t.Fatalf("unexpected metrics: %#v", got)
	}
	if got.ConfigText != "client\n" {
		t.Fatalf("ConfigText = %q, want decoded config", got.ConfigText)
	}
}

func TestParseVPNGateCSVRejectsMissingConfigColumn(t *testing.T) {
	text := strings.Join([]string{
		"#HostName,IP,Score",
		"host-a,203.0.113.10,900",
	}, "\n")

	if _, err := ParseVPNGateCSV(text); err == nil {
		t.Fatal("ParseVPNGateCSV() error = nil, want missing config column error")
	}
}

func TestDecodeOpenVPNConfigRejectsInvalidBase64(t *testing.T) {
	if _, err := DecodeOpenVPNConfig("not base64"); err == nil {
		t.Fatal("DecodeOpenVPNConfig() error = nil, want invalid base64 error")
	}
}

func TestProbeNodeMarksAvailable(t *testing.T) {
	runner := ProbeFunc(func(ctx context.Context, n Node) (ProbeResult, error) {
		return ProbeResult{Latency: 25 * time.Millisecond}, nil
	})

	got := ProbeNode(context.Background(), Node{ID: "node-1"}, runner)

	if got.ProbeStatus != ProbeAvailable {
		t.Fatalf("ProbeStatus = %q, want %q", got.ProbeStatus, ProbeAvailable)
	}
	if got.LatencyMS != 25 {
		t.Fatalf("LatencyMS = %d, want 25", got.LatencyMS)
	}
	if got.LastError != "" {
		t.Fatalf("LastError = %q, want empty", got.LastError)
	}
}

func TestProbeNodeMarksUnavailable(t *testing.T) {
	runner := ProbeFunc(func(ctx context.Context, n Node) (ProbeResult, error) {
		return ProbeResult{}, errors.New("timeout")
	})

	got := ProbeNode(context.Background(), Node{ID: "node-1"}, runner)

	if got.ProbeStatus != ProbeUnavailable {
		t.Fatalf("ProbeStatus = %q, want %q", got.ProbeStatus, ProbeUnavailable)
	}
	if got.LastError != "timeout" {
		t.Fatalf("LastError = %q, want timeout", got.LastError)
	}
}

func TestFetchCandidatesUsesSourceAndParser(t *testing.T) {
	source := SourceFunc(func(ctx context.Context) (string, error) {
		return strings.Join([]string{
			"#HostName,IP,Score,Ping,Speed,CountryLong,CountryShort,NumVpnSessions,Uptime,TotalUsers,TotalTraffic,LogType,Operator,Message,OpenVPN_ConfigData_Base64",
			"host-a,203.0.113.10,900,42,1000000,Japan,JP,1,2,3,4,2,op,msg,Y2xpZW50Cg==",
		}, "\n"), nil
	})

	nodes, err := FetchCandidates(context.Background(), source)
	if err != nil {
		t.Fatalf("FetchCandidates() error = %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("len(nodes) = %d, want 1", len(nodes))
	}
}

func TestHTTPSourceFetchesText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, _ = rw.Write([]byte("csv text"))
	}))
	defer server.Close()

	source := HTTPSource{URL: server.URL, Client: server.Client()}
	got, err := source.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if got != "csv text" {
		t.Fatalf("Fetch() = %q, want csv text", got)
	}
}
