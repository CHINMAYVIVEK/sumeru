// Package controller provides an addon-extensible HTTP route registry.
package controller

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
	AuthPublic  AuthMode = "public"
)

// Route is one registered HTTP endpoint.
type Route struct {
	Method  string
	Path    string
	Auth    AuthMode
	Handler http.HandlerFunc
}

var (
	mu     sync.RWMutex
	routes []Route
)

// Register adds a route. Call from addon init() or engine bootstrap.
// Later registration for the same Path+Method replaces the earlier handler.
func Register(method, path string, auth AuthMode, h http.HandlerFunc) {
	if h == nil || strings.TrimSpace(path) == "" {
		return
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = http.MethodGet
	}
	mu.Lock()
	defer mu.Unlock()
	for i, rt := range routes {
		if rt.Path == path && rt.Method == method {
			routes[i] = Route{Method: method, Path: path, Auth: auth, Handler: h}
			return
		}
	}
	routes = append(routes, Route{Method: method, Path: path, Auth: auth, Handler: h})
}

// All returns a copy of registered routes.
func All() []Route {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Route, len(routes))
	copy(out, routes)
	return out
}

// Clear removes all registered routes (tests only).
func Clear() {
	mu.Lock()
	defer mu.Unlock()
	routes = nil
}

// RequireSession is set by the web package to gate session-only routes.
var RequireSession func(w http.ResponseWriter, r *http.Request) bool

// ResolveUID returns the authenticated user id (session and/or API key).
var ResolveUID func(r *http.Request) int

// Apply mounts registered routes on mux, grouping handlers that share a path.
func Apply(mux *http.ServeMux) {
	if mux == nil {
		mux = http.DefaultServeMux
	}
	mu.RLock()
	list := make([]Route, len(routes))
	copy(list, routes)
	mu.RUnlock()

	byPath := map[string][]Route{}
	var order []string
	for _, rt := range list {
		if _, ok := byPath[rt.Path]; !ok {
			order = append(order, rt.Path)
		}
		byPath[rt.Path] = append(byPath[rt.Path], rt)
	}
	for _, path := range order {
		mux.HandleFunc(path, dispatch(byPath[path]))
	}
}

func dispatch(rts []Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, rt := range rts {
			if rt.Method != "" && r.Method != rt.Method {
				continue
			}
			if !gate(w, r, rt.Auth) {
				return
			}
			rt.Handler(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func gate(w http.ResponseWriter, r *http.Request, auth AuthMode) bool {
	switch auth {
	case AuthSession:
		if RequireSession != nil && !RequireSession(w, r) {
			return false
		}
	case AuthAPIKey:
		if ResolveUID != nil && ResolveUID(r) <= 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return false
		}
	}
	return true
}
