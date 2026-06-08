package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/juanlou1217/exitfleet/internal/node"
)

const (
	ExitRunning = "running"
)

type Options struct {
	Source      node.Source
	ProbeRunner node.ProbeRunner
	Launcher    WorkerLauncher
	ProxyBase   int
}

type WorkerLauncher interface {
	StartWorker(ctx context.Context, request LaunchRequest) (WorkerRecord, error)
	StopWorker(ctx context.Context, workerID string) error
}

type LaunchRequest struct {
	Node      node.Node
	ProxyPort int
}

type WorkerRecord struct {
	ID        string `json:"id"`
	ProxyPort int    `json:"proxy_port"`
	State     string `json:"state"`
}

type Exit struct {
	ID        string    `json:"id"`
	NodeID    string    `json:"node_id"`
	WorkerID  string    `json:"worker_id"`
	ProxyPort int       `json:"proxy_port"`
	State     string    `json:"state"`
	StartedAt time.Time `json:"started_at"`
}

type LogEvent struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Module  string    `json:"module"`
	Message string    `json:"message"`
}

type Snapshot struct {
	Nodes []node.Node `json:"nodes"`
	Exits []Exit      `json:"exits"`
	Logs  []LogEvent  `json:"logs"`
}

type Manager struct {
	mu          sync.RWMutex
	source      node.Source
	probeRunner node.ProbeRunner
	launcher    WorkerLauncher
	proxyBase   int
	nextExit    int
	nodes       map[string]node.Node
	nodeOrder   []string
	exits       map[string]Exit
	logs        []LogEvent
}

func New(options Options) *Manager {
	proxyBase := options.ProxyBase
	if proxyBase == 0 {
		proxyBase = 7928
	}
	return &Manager{
		source:      options.Source,
		probeRunner: options.ProbeRunner,
		launcher:    options.Launcher,
		proxyBase:   proxyBase,
		nodes:       make(map[string]node.Node),
		exits:       make(map[string]Exit),
	}
}

func (m *Manager) RefreshNodes(ctx context.Context) ([]node.Node, error) {
	nodes, err := node.FetchCandidates(ctx, m.source)
	if err != nil {
		m.log("ERROR", "Node", err.Error())
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.nodes = make(map[string]node.Node, len(nodes))
	m.nodeOrder = m.nodeOrder[:0]
	for _, n := range nodes {
		m.nodes[n.ID] = n
		m.nodeOrder = append(m.nodeOrder, n.ID)
	}
	m.appendLog("INFO", "Node", fmt.Sprintf("refreshed %d candidate nodes", len(nodes)))
	return copyNodes(nodes), nil
}

func (m *Manager) TestNodes(ctx context.Context, ids []string) ([]node.Node, error) {
	if m.probeRunner == nil {
		return nil, fmt.Errorf("probe runner is required")
	}

	tested := make([]node.Node, 0, len(ids))
	for _, id := range ids {
		current, ok := m.nodeByID(id)
		if !ok {
			return nil, fmt.Errorf("node %q not found", id)
		}
		updated := node.ProbeNode(ctx, current, m.probeRunner)
		m.mu.Lock()
		m.nodes[id] = updated
		m.mu.Unlock()
		tested = append(tested, updated)
	}
	m.log("INFO", "Node", fmt.Sprintf("tested %d nodes", len(tested)))
	return tested, nil
}

func (m *Manager) StartExit(ctx context.Context, nodeID string) (Exit, error) {
	if m.launcher == nil {
		return Exit{}, fmt.Errorf("worker launcher is required")
	}
	n, ok := m.nodeByID(nodeID)
	if !ok {
		return Exit{}, fmt.Errorf("node %q not found", nodeID)
	}
	if m.activeExitForNode(nodeID) {
		return Exit{}, fmt.Errorf("node %q already has an active exit", nodeID)
	}

	proxyPort := m.nextProxyPort()
	if proxyPort == 0 {
		return Exit{}, fmt.Errorf("no proxy port available")
	}
	worker, err := m.launcher.StartWorker(ctx, LaunchRequest{Node: n, ProxyPort: proxyPort})
	if err != nil {
		m.log("ERROR", "Worker", err.Error())
		return Exit{}, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextExit++
	exit := Exit{
		ID:        fmt.Sprintf("exit-%d", m.nextExit),
		NodeID:    nodeID,
		WorkerID:  worker.ID,
		ProxyPort: worker.ProxyPort,
		State:     ExitRunning,
		StartedAt: time.Now().UTC(),
	}
	m.exits[exit.ID] = exit
	m.appendLog("INFO", "Worker", fmt.Sprintf("started worker %s for node %s", worker.ID, nodeID))
	return exit, nil
}

func (m *Manager) StopExit(ctx context.Context, exitID string) error {
	m.mu.RLock()
	exit, ok := m.exits[exitID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("exit %q not found", exitID)
	}
	if m.launcher != nil {
		if err := m.launcher.StopWorker(ctx, exit.WorkerID); err != nil {
			m.log("ERROR", "Worker", err.Error())
			return err
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.exits, exitID)
	m.appendLog("INFO", "Worker", fmt.Sprintf("stopped worker %s", exit.WorkerID))
	return nil
}

func (m *Manager) Snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	nodes := make([]node.Node, 0, len(m.nodeOrder))
	for _, id := range m.nodeOrder {
		nodes = append(nodes, m.nodes[id])
	}
	exits := make([]Exit, 0, len(m.exits))
	for _, exit := range m.exits {
		exits = append(exits, exit)
	}
	logs := append([]LogEvent(nil), m.logs...)
	return Snapshot{Nodes: nodes, Exits: exits, Logs: logs}
}

func (m *Manager) nodeByID(id string) (node.Node, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n, ok := m.nodes[id]
	return n, ok
}

func (m *Manager) nextProxyPort() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	used := make(map[int]struct{}, len(m.exits))
	for _, exit := range m.exits {
		used[exit.ProxyPort] = struct{}{}
	}
	for port := m.proxyBase; port <= 65535; port++ {
		if _, ok := used[port]; !ok {
			return port
		}
	}
	return 0
}

func (m *Manager) log(level, module, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.appendLog(level, module, message)
}

func (m *Manager) appendLog(level, module, message string) {
	m.logs = append(m.logs, LogEvent{
		Time:    time.Now().UTC(),
		Level:   level,
		Module:  module,
		Message: message,
	})
}

func copyNodes(nodes []node.Node) []node.Node {
	return append([]node.Node(nil), nodes...)
}

func (m *Manager) activeExitForNode(nodeID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, exit := range m.exits {
		if exit.NodeID == nodeID {
			return true
		}
	}
	return false
}
