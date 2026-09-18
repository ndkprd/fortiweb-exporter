// Package handler serves per-target Prometheus metrics over HTTP, selecting
// which configured FortiWeb appliance to scrape via a "target" query
// parameter.
package handler

import (
	"fmt"
	"net/http"

	"github.com/ndkprd/fortiweb_exporter/internal/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler serves /metrics?target=NAME, scraping only the FortiWeb
// appliance registered under that name.
type MetricsHandler struct {
	clients map[string]collector.StatusGetter
}

// NewMetricsHandler builds a MetricsHandler over a pre-built set of clients,
// one per configured FortiWeb target name.
func NewMetricsHandler(clients map[string]collector.StatusGetter) *MetricsHandler {
	return &MetricsHandler{clients: clients}
}

// ServeHTTP implements http.Handler.
func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
