package exporter

import (
	_ "embed"
	"net/http"
	"strings"
)

//go:embed dashboard.html
var dashboardHTML string

//go:embed dashboard.css
var dashboardCSS string

func serveDashboard(w http.ResponseWriter, r *http.Request) {
	// Only serve dashboard on exact root path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Replace the CSS placeholder with the actual CSS content
	html := strings.Replace(dashboardHTML, "/* EMBEDDED_CSS */", dashboardCSS, 1)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}
