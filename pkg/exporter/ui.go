package exporter

import (
	_ "embed"
	"net/http"
)

//go:embed dashboard.html
var dashboardHTML string

func serveDashboard(w http.ResponseWriter, r *http.Request) {
	// Only serve dashboard on exact root path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(dashboardHTML))
}
