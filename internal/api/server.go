package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/juanlou1217/exitfleet/internal/manager"
	"github.com/juanlou1217/exitfleet/internal/node"
)

type ManagerHandler struct {
	manager *manager.Manager
}

func NewManagerHandler(mgr *manager.Manager) http.Handler {
	return &ManagerHandler{manager: mgr}
}

func (h *ManagerHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/state":
		writeJSON(rw, http.StatusOK, publicSnapshot(h.manager.Snapshot()))
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/nodes":
		writeJSON(rw, http.StatusOK, map[string]any{"nodes": publicNodes(h.manager.Snapshot().Nodes)})
	case r.Method == http.MethodPost && r.URL.Path == "/api/v1/nodes/refresh":
		nodes, err := h.manager.RefreshNodes(r.Context())
		if err != nil {
			writeError(rw, http.StatusInternalServerError, err)
			return
		}
		writeJSON(rw, http.StatusOK, map[string]any{"nodes": publicNodes(nodes)})
	case r.Method == http.MethodPost && r.URL.Path == "/api/v1/nodes/test":
		var payload struct {
			IDs []string `json:"ids"`
		}
		if err := readJSON(r, &payload); err != nil {
			writeError(rw, http.StatusBadRequest, err)
			return
		}
		nodes, err := h.manager.TestNodes(r.Context(), payload.IDs)
		if err != nil {
			writeError(rw, http.StatusBadRequest, err)
			return
		}
		writeJSON(rw, http.StatusOK, map[string]any{"nodes": publicNodes(nodes)})
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/exits":
		writeJSON(rw, http.StatusOK, map[string]any{"exits": h.manager.Snapshot().Exits})
	case r.Method == http.MethodPost && r.URL.Path == "/api/v1/exits":
		var payload struct {
			NodeID string `json:"node_id"`
		}
		if err := readJSON(r, &payload); err != nil {
			writeError(rw, http.StatusBadRequest, err)
			return
		}
		if strings.TrimSpace(payload.NodeID) == "" {
			writeError(rw, http.StatusBadRequest, fmt.Errorf("node_id is required"))
			return
		}
		exit, err := h.manager.StartExit(r.Context(), payload.NodeID)
		if err != nil {
			writeError(rw, http.StatusBadRequest, err)
			return
		}
		writeJSON(rw, http.StatusCreated, map[string]any{"exit": exit})
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/v1/exits/"):
		exitID := strings.TrimPrefix(r.URL.Path, "/api/v1/exits/")
		if exitID == "" {
			writeError(rw, http.StatusBadRequest, fmt.Errorf("exit id is required"))
			return
		}
		if err := h.manager.StopExit(r.Context(), exitID); err != nil {
			writeError(rw, http.StatusNotFound, err)
			return
		}
		writeJSON(rw, http.StatusOK, map[string]any{"ok": true})
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/logs/stream":
		h.writeLogStream(rw)
	default:
		writeError(rw, http.StatusNotFound, fmt.Errorf("not found"))
	}
}

func (h *ManagerHandler) writeLogStream(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.WriteHeader(http.StatusOK)
	for _, event := range h.manager.Snapshot().Logs {
		data, _ := json.Marshal(event)
		_, _ = fmt.Fprintf(rw, "event: log\ndata: %s\n\n", data)
	}
}

func readJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func writeJSON(rw http.ResponseWriter, status int, value any) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	_ = json.NewEncoder(rw).Encode(value)
}

func writeError(rw http.ResponseWriter, status int, err error) {
	writeJSON(rw, status, map[string]string{"error": err.Error()})
}

func publicSnapshot(snapshot manager.Snapshot) manager.Snapshot {
	snapshot.Nodes = publicNodes(snapshot.Nodes)
	return snapshot
}

func publicNodes(nodes []node.Node) []node.Node {
	out := append([]node.Node(nil), nodes...)
	for i := range out {
		out[i].ConfigText = ""
	}
	return out
}
