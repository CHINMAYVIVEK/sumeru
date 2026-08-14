// Package router provides an addon-extensible HTTP route registry.
package router

import (
	"net/http"
	"strings"
	"sync"
)

// AuthMode selects how a registered route authenticates the caller.
type AuthMode string

const (
	AuthNone    AuthMode = "none"
	AuthSession AuthMode = "session"
	AuthAPIKey  AuthMode = "apikey" // session or API key (uid > 0)
)

const (
	methodNotAllowedMessage = "Method not allowed"
	unauthorizedMessage     = "Unauthorized"
)

// Route is a method-specific handler mounted at a path with an auth policy.
type Route struct {
	Method  string
	Path    string
	Auth    AuthMode
	Handler http.HandlerFunc
}

var (
	routeMu          sync.RWMutex
	registeredRoutes []Route
)

// RequireSession is set by the web package to gate session-only routes.
var RequireSession func(w http.ResponseWriter, r *http.Request) bool

// ResolveUID returns the authenticated user id (session and/or API key).
var ResolveUID func(r *http.Request) int

// Register adds a route. Call from addon init() or engine bootstrap.
// Later registration for the same path+method replaces the earlier handler.
func Register(method, path string, auth AuthMode, handler http.HandlerFunc) {
	if !isRegisterableRoute(path, handler) {
		return
	}

	route := Route{
		Method:  normalizeHTTPMethod(method),
		Path:    path,
		Auth:    auth,
		Handler: handler,
	}

	routeMu.Lock()
	defer routeMu.Unlock()
	upsertRoute(route)
}

// Clear removes all registered routes (tests only).
func Clear() {
	routeMu.Lock()
	defer routeMu.Unlock()
	registeredRoutes = nil
}

// Apply mounts registered routes on mux, grouping handlers that share a path.
func Apply(mux *http.ServeMux) {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	pathsInOrder, routesByPath := groupRoutesByPath(snapshotRoutes())
	for _, path := range pathsInOrder {
		mux.HandleFunc(path, dispatch(routesByPath[path]))
	}
}

func isRegisterableRoute(path string, handler http.HandlerFunc) bool {
	return handler != nil && strings.TrimSpace(path) != ""
}

func normalizeHTTPMethod(method string) string {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		return http.MethodGet
	}
	return method
}

func upsertRoute(route Route) {
	for index, existing := range registeredRoutes {
		if existing.Path == route.Path && existing.Method == route.Method {
			registeredRoutes[index] = route
			return
		}
	}
	registeredRoutes = append(registeredRoutes, route)
}

func snapshotRoutes() []Route {
	routeMu.RLock()
	defer routeMu.RUnlock()

	routes := make([]Route, len(registeredRoutes))
	copy(routes, registeredRoutes)
	return routes
}

func groupRoutesByPath(routes []Route) (pathsInOrder []string, routesByPath map[string][]Route) {
	routesByPath = make(map[string][]Route)
	for _, route := range routes {
		if _, seen := routesByPath[route.Path]; !seen {
			pathsInOrder = append(pathsInOrder, route.Path)
		}
		routesByPath[route.Path] = append(routesByPath[route.Path], route)
	}
	return pathsInOrder, routesByPath
}

func dispatch(pathRoutes []Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, route := range pathRoutes {
			if !requestMatchesRoute(r, route) {
				continue
			}
			if !authorizeRequest(w, r, route.Auth) {
				return
			}
			route.Handler(w, r)
			return
		}
		http.Error(w, methodNotAllowedMessage, http.StatusMethodNotAllowed)
	}
}

func requestMatchesRoute(r *http.Request, route Route) bool {
	return route.Method == "" || r.Method == route.Method
}

func authorizeRequest(w http.ResponseWriter, r *http.Request, auth AuthMode) bool {
	switch auth {
	case AuthSession:
		return allowSessionRequest(w, r)
	case AuthAPIKey:
		return allowAuthenticatedRequest(w, r)
	default:
		return true
	}
}

func allowSessionRequest(w http.ResponseWriter, r *http.Request) bool {
	if RequireSession == nil {
		return true
	}
	return RequireSession(w, r)
}

func allowAuthenticatedRequest(w http.ResponseWriter, r *http.Request) bool {
	if ResolveUID != nil && ResolveUID(r) <= 0 {
		http.Error(w, unauthorizedMessage, http.StatusUnauthorized)
		return false
	}
	return true
}
