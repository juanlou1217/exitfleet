package api

type Route struct {
	Method string
	Path   string
	Owner  string
}

const (
	OwnerManager = "manager"
	OwnerWorker  = "worker"
)

var Routes = []Route{
	{Method: "GET", Path: "/api/v1/state", Owner: OwnerManager},
	{Method: "GET", Path: "/api/v1/nodes", Owner: OwnerManager},
	{Method: "POST", Path: "/api/v1/nodes/refresh", Owner: OwnerManager},
	{Method: "POST", Path: "/api/v1/nodes/test", Owner: OwnerManager},
	{Method: "GET", Path: "/api/v1/exits", Owner: OwnerManager},
	{Method: "POST", Path: "/api/v1/exits", Owner: OwnerManager},
	{Method: "DELETE", Path: "/api/v1/exits/{id}", Owner: OwnerManager},
	{Method: "GET", Path: "/api/v1/logs/stream", Owner: OwnerManager},
	{Method: "GET", Path: "/healthz", Owner: OwnerWorker},
	{Method: "GET", Path: "/readyz", Owner: OwnerWorker},
}

func ManagerRoutes() []Route {
	return routesByOwner(OwnerManager)
}

func WorkerRoutes() []Route {
	return routesByOwner(OwnerWorker)
}

func routesByOwner(owner string) []Route {
	routes := make([]Route, 0, len(Routes))
	for _, route := range Routes {
		if route.Owner == owner {
			routes = append(routes, route)
		}
	}
	return routes
}
