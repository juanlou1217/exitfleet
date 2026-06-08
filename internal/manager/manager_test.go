package manager

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/juanlou1217/exitfleet/internal/node"
)

func TestRefreshNodesDoesNotCreateWorker(t *testing.T) {
	source := staticSource()
	launcher := &fakeLauncher{}
	mgr := New(Options{Source: source, Launcher: launcher})

	if _, err := mgr.RefreshNodes(context.Background()); err != nil {
		t.Fatalf("RefreshNodes() error = %v", err)
	}

	snapshot := mgr.Snapshot()
	if len(snapshot.Nodes) != 1 {
		t.Fatalf("len(nodes) = %d, want 1", len(snapshot.Nodes))
	}
	if len(snapshot.Exits) != 0 {
		t.Fatalf("len(exits) = %d, want 0", len(snapshot.Exits))
	}
	if launcher.started != 0 {
		t.Fatalf("launcher.started = %d, want 0", launcher.started)
	}
}

func TestStartExitCreatesOneWorkerRecord(t *testing.T) {
	launcher := &fakeLauncher{}
	mgr := New(Options{Source: staticSource(), Launcher: launcher})
	nodes, err := mgr.RefreshNodes(context.Background())
	if err != nil {
		t.Fatalf("RefreshNodes() error = %v", err)
	}

	exit, err := mgr.StartExit(context.Background(), nodes[0].ID)
	if err != nil {
		t.Fatalf("StartExit() error = %v", err)
	}

	if exit.NodeID != nodes[0].ID || exit.WorkerID == "" {
		t.Fatalf("exit = %#v, want node id and worker id", exit)
	}
	if launcher.started != 1 {
		t.Fatalf("launcher.started = %d, want 1", launcher.started)
	}
	if len(mgr.Snapshot().Exits) != 1 {
		t.Fatalf("len(exits) = %d, want 1", len(mgr.Snapshot().Exits))
	}
}

func TestStopExitReleasesWorkerRecord(t *testing.T) {
	launcher := &fakeLauncher{}
	mgr := New(Options{Source: staticSource(), Launcher: launcher})
	nodes, _ := mgr.RefreshNodes(context.Background())
	exit, _ := mgr.StartExit(context.Background(), nodes[0].ID)

	if err := mgr.StopExit(context.Background(), exit.ID); err != nil {
		t.Fatalf("StopExit() error = %v", err)
	}

	if len(mgr.Snapshot().Exits) != 0 {
		t.Fatalf("len(exits) = %d, want 0", len(mgr.Snapshot().Exits))
	}
	if launcher.stopped != 1 {
		t.Fatalf("launcher.stopped = %d, want 1", launcher.stopped)
	}
}

func TestStartExitRejectsDuplicateActiveNode(t *testing.T) {
	launcher := &fakeLauncher{}
	mgr := New(Options{Source: staticSource(), Launcher: launcher})
	nodes, _ := mgr.RefreshNodes(context.Background())
	if _, err := mgr.StartExit(context.Background(), nodes[0].ID); err != nil {
		t.Fatalf("first StartExit() error = %v", err)
	}

	if _, err := mgr.StartExit(context.Background(), nodes[0].ID); err == nil {
		t.Fatal("second StartExit() error = nil, want duplicate active node error")
	}
	if launcher.started != 1 {
		t.Fatalf("launcher.started = %d, want 1", launcher.started)
	}
}

func TestStartExitAllocatesFirstFreeProxyPort(t *testing.T) {
	launcher := &fakeLauncher{}
	mgr := New(Options{
		Source: node.SourceFunc(func(ctx context.Context) (string, error) {
			return strings.Join([]string{
				"#HostName,IP,Score,Ping,Speed,CountryLong,CountryShort,NumVpnSessions,Uptime,TotalUsers,TotalTraffic,LogType,Operator,Message,OpenVPN_ConfigData_Base64",
				"host-a,203.0.113.10,900,42,1000000,Japan,JP,1,2,3,4,2,op,msg,Y2xpZW50Cg==",
				"host-b,203.0.113.11,800,43,900000,Korea,KR,1,2,3,4,2,op,msg,Y2xpZW50Cg==",
				"host-c,203.0.113.12,700,44,800000,United States,US,1,2,3,4,2,op,msg,Y2xpZW50Cg==",
			}, "\n"), nil
		}),
		Launcher: launcher,
	})
	nodes, _ := mgr.RefreshNodes(context.Background())
	first, _ := mgr.StartExit(context.Background(), nodes[0].ID)
	second, _ := mgr.StartExit(context.Background(), nodes[1].ID)
	if first.ProxyPort != 7928 || second.ProxyPort != 7929 {
		t.Fatalf("initial ports = %d/%d, want 7928/7929", first.ProxyPort, second.ProxyPort)
	}
	if err := mgr.StopExit(context.Background(), first.ID); err != nil {
		t.Fatalf("StopExit() error = %v", err)
	}

	third, err := mgr.StartExit(context.Background(), nodes[2].ID)
	if err != nil {
		t.Fatalf("third StartExit() error = %v", err)
	}
	if third.ProxyPort != 7928 {
		t.Fatalf("third.ProxyPort = %d, want first free port 7928", third.ProxyPort)
	}
}

func TestTestNodesUpdatesProbeStatus(t *testing.T) {
	mgr := New(Options{
		Source: staticSource(),
		ProbeRunner: node.ProbeFunc(func(ctx context.Context, n node.Node) (node.ProbeResult, error) {
			return node.ProbeResult{Latency: 14 * time.Millisecond}, nil
		}),
		Launcher: &fakeLauncher{},
	})
	nodes, _ := mgr.RefreshNodes(context.Background())

	tested, err := mgr.TestNodes(context.Background(), []string{nodes[0].ID})
	if err != nil {
		t.Fatalf("TestNodes() error = %v", err)
	}
	if tested[0].ProbeStatus != node.ProbeAvailable || tested[0].LatencyMS != 14 {
		t.Fatalf("tested node = %#v, want available latency", tested[0])
	}
}

func TestStartExitRejectsUnknownNode(t *testing.T) {
	mgr := New(Options{Launcher: &fakeLauncher{}})

	if _, err := mgr.StartExit(context.Background(), "missing"); err == nil {
		t.Fatal("StartExit() error = nil, want unknown node error")
	}
}

func TestLogsRecordLifecycleEvents(t *testing.T) {
	mgr := New(Options{Source: staticSource(), Launcher: &fakeLauncher{}})
	nodes, _ := mgr.RefreshNodes(context.Background())
	_, _ = mgr.StartExit(context.Background(), nodes[0].ID)

	if len(mgr.Snapshot().Logs) == 0 {
		t.Fatal("logs are empty, want lifecycle events")
	}
}

func staticSource() node.Source {
	return node.SourceFunc(func(ctx context.Context) (string, error) {
		return strings.Join([]string{
			"#HostName,IP,Score,Ping,Speed,CountryLong,CountryShort,NumVpnSessions,Uptime,TotalUsers,TotalTraffic,LogType,Operator,Message,OpenVPN_ConfigData_Base64",
			"host-a,203.0.113.10,900,42,1000000,Japan,JP,1,2,3,4,2,op,msg,Y2xpZW50Cg==",
		}, "\n"), nil
	})
}

type fakeLauncher struct {
	started int
	stopped int
	fail    bool
}

func (f *fakeLauncher) StartWorker(ctx context.Context, request LaunchRequest) (WorkerRecord, error) {
	if f.fail {
		return WorkerRecord{}, errors.New("launch failed")
	}
	f.started++
	return WorkerRecord{ID: "worker-" + request.Node.ID, ProxyPort: request.ProxyPort, State: "ready"}, nil
}

func (f *fakeLauncher) StopWorker(ctx context.Context, workerID string) error {
	f.stopped++
	return nil
}
