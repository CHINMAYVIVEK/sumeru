package web

import (
	"net/http"

	"github.com/gorilla/websocket"
)

const swcBusRoute = "/web/swc/bus"

func registerSwcBusRoute() {
	registerSession(http.MethodGet, swcBusRoute, SwcBusHandler)
}

// SwcBusHandler upgrades GET /web/swc/bus to a WebSocket for live outbox events.
func SwcBusHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !websocket.IsWebSocketUpgrade(r) {
		http.NotFound(w, r)
		return
	}
	serveSwcBusWebSocket(w, r, AuthenticatedUserID(r))
}
