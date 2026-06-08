package api

import "testing"

func TestManagerRoutesContainExitLifecycle(t *testing.T) {
	routes := ManagerRoutes()

	assertRoute(t, routes, "GET", "/api/v1/exits")
	assertRoute(t, routes, "POST", "/api/v1/exits")
	assertRoute(t, routes, "DELETE", "/api/v1/exits/{id}")
}

func TestWorkerRoutesExposeHealthAndReadiness(t *testing.T) {
	routes := WorkerRoutes()

	assertRoute(t, routes, "GET", "/healthz")
	assertRoute(t, routes, "GET", "/readyz")
}

func assertRoute(t *testing.T, routes []Route, method string, path string) {
	t.Helper()
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return
		}
	}
	t.Fatalf("route %s %s not found in %#v", method, path, routes)
}

