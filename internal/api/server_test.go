package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/juanlou1217/exitfleet/internal/manager"
	"github.com/juanlou1217/exitfleet/internal/node"
)

func TestManagerAPIRefreshStartStopFlow(t *testing.T) {
	mgr := testManager()
	handler := NewManagerHandler(mgr)

	refresh := httptest.NewRecorder()
	handler.ServeHTTP(refresh, httptest.NewRequest(http.MethodPost, "/api/v1/nodes/refresh", nil))
	if refresh.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, want 200", refresh.Code)
	}

	nodesResp := httptest.NewRecorder()
	handler.ServeHTTP(nodesResp, httptest.NewRequest(http.MethodGet, "/api/v1/nodes", nil))
	var nodesBody struct {
		Nodes []node.Node `json:"nodes"`
	}
	if err := json.Unmarshal(nodesResp.Body.Bytes(), &nodesBody); err != nil {
		t.Fatalf("decode nodes response: %v", err)
	}
	if len(nodesBody.Nodes) != 1 {
		t.Fatalf("len(nodes) = %d, want 1", len(nodesBody.Nodes))
	}

	start := httptest.NewRecorder()
	body := strings.NewReader(`{"node_id":"` + nodesBody.Nodes[0].ID + `"}`)
	handler.ServeHTTP(start, httptest.NewRequest(http.MethodPost, "/api/v1/exits", body))
	if start.Code != http.StatusCreated {
		t.Fatalf("start status = %d, want 201 body=%s", start.Code, start.Body.String())
	}
	var startBody struct {
		Exit manager.Exit `json:"exit"`
	}
	if err := json.Unmarshal(start.Body.Bytes(), &startBody); err != nil {
		t.Fatalf("decode start response: %v", err)
	}
	if startBody.Exit.WorkerID == "" {
		t.Fatalf("exit = %#v, want worker id", startBody.Exit)
	}

	stop := httptest.NewRecorder()
	handler.ServeHTTP(stop, httptest.NewRequest(http.MethodDelete, "/api/v1/exits/"+startBody.Exit.ID, nil))
	if stop.Code != http.StatusOK {
		t.Fatalf("stop status = %d, want 200 body=%s", stop.Code, stop.Body.String())
	}
}

func TestManagerAPITestNodes(t *testing.T) {
	mgr := testManager()
	handler := NewManagerHandler(mgr)
	_, _ = mgr.RefreshNodes(context.Background())
	nodeID := mgr.Snapshot().Nodes[0].ID

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/test", strings.NewReader(`{"ids":["`+nodeID+`"]}`))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("test nodes status = %d, want 200 body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Nodes []node.Node `json:"nodes"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Nodes[0].ProbeStatus != node.ProbeAvailable {
		t.Fatalf("probe status = %q, want available", body.Nodes[0].ProbeStatus)
	}
}

func TestManagerAPINodesDoesNotExposeOpenVPNConfigText(t *testing.T) {
	mgr := testManager()
	handler := NewManagerHandler(mgr)
	_, _ = mgr.RefreshNodes(context.Background())

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/nodes", nil))

	if strings.Contains(rec.Body.String(), "config_text") || strings.Contains(rec.Body.String(), "client") {
		t.Fatalf("nodes response exposed config text: %s", rec.Body.String())
	}
}

func TestManagerAPIRejectsMissingExitNodeID(t *testing.T) {
	rec := httptest.NewRecorder()
	NewManagerHandler(testManager()).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/exits", strings.NewReader(`{}`)))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestManagerAPINotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	NewManagerHandler(testManager()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/missing", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestLogsStreamReturnsSSE(t *testing.T) {
	mgr := testManager()
	_, _ = mgr.RefreshNodes(context.Background())

	rec := httptest.NewRecorder()
	NewManagerHandler(mgr).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/logs/stream", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("Content-Type = %q, want text/event-stream", got)
	}
	if !strings.Contains(rec.Body.String(), "event: log") || !strings.Contains(rec.Body.String(), "data:") {
		t.Fatalf("SSE body = %q, want event and data", rec.Body.String())
	}
}

func testManager() *manager.Manager {
	return manager.New(manager.Options{
		Source: node.SourceFunc(func(ctx context.Context) (string, error) {
			return strings.Join([]string{
				"#HostName,IP,Score,Ping,Speed,CountryLong,CountryShort,NumVpnSessions,Uptime,TotalUsers,TotalTraffic,LogType,Operator,Message,OpenVPN_ConfigData_Base64",
				"host-a,203.0.113.10,900,42,1000000,Japan,JP,1,2,3,4,2,op,msg,Y2xpZW50Cg==",
			}, "\n"), nil
		}),
		ProbeRunner: node.ProbeFunc(func(ctx context.Context, n node.Node) (node.ProbeResult, error) {
			return node.ProbeResult{Latency: 12 * time.Millisecond}, nil
		}),
		Launcher: testLauncher{},
	})
}

type testLauncher struct{}

func (testLauncher) StartWorker(ctx context.Context, request manager.LaunchRequest) (manager.WorkerRecord, error) {
	return manager.WorkerRecord{ID: "worker-" + request.Node.ID, ProxyPort: request.ProxyPort, State: "ready"}, nil
}

func (testLauncher) StopWorker(ctx context.Context, workerID string) error {
	return nil
}
