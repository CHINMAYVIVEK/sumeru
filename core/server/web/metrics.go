package web

import (
	"net/http"

	"sumeru/core/metrics"
)

// MetricsHandler exposes Prometheus text metrics to session users in base.group_system.
// Access is enforced by requireSystemAdmin (groupSystemXML).
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireSystemAdmin(w, r, false) {
		return
	}
	metrics.Handler(w, r)
}
