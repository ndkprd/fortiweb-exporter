// Package handler serves the exporter's HTTP endpoints: a landing page at
// "/", a per-target probe at ProbePath, and the exporter's own internal
// metrics at MetricsPath.
package handler

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/endekasoft/fortiweb-exporter/internal/collector"
)

// ProbeHandler serves ProbePath?target=NAME, scraping only the FortiWeb
// appliance registered under that name.
type ProbeHandler struct {
	clients map[string]collector.StatusGetter
}

// NewProbeHandler builds a ProbeHandler over a pre-built set of clients, one
// per configured FortiWeb target name.
func NewProbeHandler(clients map[string]collector.StatusGetter) *ProbeHandler {
	return &ProbeHandler{clients: clients}
}

// ServeHTTP implements http.Handler.
func (h *ProbeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "target parameter is required", http.StatusBadRequest)
		return
	}

	client, ok := h.clients[target]
	if !ok {
		http.Error(w, fmt.Sprintf("unknown target %q", target), http.StatusBadRequest)
		return
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector.NewCollector(client))
	promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}
