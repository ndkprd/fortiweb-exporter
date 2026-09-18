package handler

// Fixed HTTP paths the exporter serves, following the convention used by
// blackbox_exporter: ProbePath scrapes a specific target selected via a
// "target" query parameter, MetricsPath exposes the exporter's own internal
// (Go runtime/process) metrics.
const (
	ProbePath   = "/probe"
	MetricsPath = "/metrics"
)
