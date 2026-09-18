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

// Target is a configured FortiWeb appliance, ready to be scraped: a client
// plus the vdom it's scoped to (used to label vdom-specific gauges).
type Target struct {
	Client collector.StatusGetter
	VDOM   string
}

// ProbeHandler serves ProbePath?target=NAME, scraping only the FortiWeb
// appliance registered under that name.
type ProbeHandler struct {
	targets map[string]Target
}

// NewProbeHandler builds a ProbeHandler over a pre-built set of targets, one
// per configured FortiWeb target name.
func NewProbeHandler(targets map[string]Target) *ProbeHandler {
	return &ProbeHandler{targets: targets}
}

// ServeHTTP implements http.Handler.
func (h *ProbeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("target")
	if name == "" {
		http.Error(w, "target parameter is required", http.StatusBadRequest)
		return
	}

	target, ok := h.targets[name]
	if !ok {
		http.Error(w, fmt.Sprintf("unknown target %q", name), http.StatusBadRequest)
		return
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector.NewCollector(target.Client, target.VDOM))
	promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}
