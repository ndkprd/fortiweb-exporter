// Package collector implements a prometheus.Collector that scrapes a
// FortiWeb appliance's system resource status on every collection.
package collector

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"gitlab.com/endekasoft/fortiweb-exporter/internal/fortiweb"
)

const availableStatus = "Available"

// StatusGetter fetches FortiWeb system resource status and resource counts.
// *fortiweb.Client satisfies this interface; tests can supply a fake.
type StatusGetter interface {
	GetSystemResourceStatus(ctx context.Context) (*fortiweb.SystemResourceStatus, error)
	GetServerPolicyCount(ctx context.Context) (int, error)
	GetContentRoutingCount(ctx context.Context) (int, error)
	GetServerPoolCount(ctx context.Context) (int, error)
	GetProtectionProfileCount(ctx context.Context) (int, error)
	GetAllowedHostsCount(ctx context.Context) (int, error)
	GetVirtualServerCount(ctx context.Context) (int, error)
	GetSignatureCount(ctx context.Context) (int, error)
}

// countMetric pairs a vdom-labeled resource-count gauge with the call that
// fetches its current value.
type countMetric struct {
	desc  *prometheus.Desc
	fetch func(context.Context) (int, error)
	name  string
}

// Collector adapts a StatusGetter into a prometheus.Collector, labeling the
// vdom-scoped resource-count gauges with the client's own configured vdom.
type Collector struct {
	client StatusGetter
	vdom   string

	up                   *prometheus.Desc
	cpuUsagePercent      *prometheus.Desc
	memoryUsagePercent   *prometheus.Desc
	diskUsagePercent     *prometheus.Desc
	sessionCount         *prometheus.Desc
	connectionsPerSecond *prometheus.Desc
	logDiskAvailable     *prometheus.Desc
	dbStatusAvailable    *prometheus.Desc

	counts []countMetric
}

// NewCollector builds a Collector that scrapes status from client, labeling
// its resource-count gauges with vdom.
func NewCollector(client StatusGetter, vdom string) *Collector {
	c := &Collector{
		client: client,
		vdom:   vdom,
		up: prometheus.NewDesc(
			"fortiweb_up",
			"Whether the last scrape of the FortiWeb API succeeded (1 for success, 0 for failure).",
			nil, nil,
		),
		cpuUsagePercent: prometheus.NewDesc(
			"fortiweb_cpu_usage_percent",
			"FortiWeb CPU usage percentage.",
			nil, nil,
		),
		memoryUsagePercent: prometheus.NewDesc(
			"fortiweb_memory_usage_percent",
			"FortiWeb memory usage percentage.",
			nil, nil,
		),
		diskUsagePercent: prometheus.NewDesc(
			"fortiweb_disk_usage_percent",
			"FortiWeb disk usage percentage.",
			nil, nil,
		),
		sessionCount: prometheus.NewDesc(
			"fortiweb_session_count",
			"Current FortiWeb session count.",
			nil, nil,
		),
		connectionsPerSecond: prometheus.NewDesc(
			"fortiweb_connections_per_second",
			"FortiWeb connections per second.",
			nil, nil,
		),
		logDiskAvailable: prometheus.NewDesc(
			"fortiweb_log_disk_available",
			"Whether the FortiWeb log disk is available (1) or not (0).",
			nil, nil,
		),
		dbStatusAvailable: prometheus.NewDesc(
			"fortiweb_db_status_available",
			"Whether the FortiWeb database status is available (1) or not (0).",
			nil, nil,
		),
	}

	type spec struct {
		metric string
		help   string
		fetch  func(context.Context) (int, error)
	}
	specs := []spec{
		{"fortiweb_server_policy_count", "Current number of server-policy objects configured on the vdom.", client.GetServerPolicyCount},
		{"fortiweb_content_routing_count", "Current number of content-routing-policy objects configured on the vdom.", client.GetContentRoutingCount},
		{"fortiweb_server_pool_count", "Current number of server-pool objects configured on the vdom.", client.GetServerPoolCount},
		{"fortiweb_protection_profile_count", "Current number of protection-profile objects configured on the vdom.", client.GetProtectionProfileCount},
		{"fortiweb_allowed_hosts_count", "Current number of allowed-hosts objects configured on the vdom.", client.GetAllowedHostsCount},
		{"fortiweb_virtual_server_count", "Current number of virtual-server objects configured on the vdom.", client.GetVirtualServerCount},
		{"fortiweb_signature_count", "Current number of signature objects configured on the vdom.", client.GetSignatureCount},
	}
	c.counts = make([]countMetric, len(specs))
	for i, s := range specs {
		c.counts[i] = countMetric{
			desc:  prometheus.NewDesc(s.metric, s.help, []string{"vdom"}, nil),
			fetch: s.fetch,
			name:  s.metric,
		}
	}

	return c
}

// Describe implements prometheus.Collector.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.cpuUsagePercent
	ch <- c.memoryUsagePercent
	ch <- c.diskUsagePercent
	ch <- c.sessionCount
	ch <- c.connectionsPerSecond
	ch <- c.logDiskAvailable
	ch <- c.dbStatusAvailable
	for _, m := range c.counts {
		ch <- m.desc
	}
}

// Collect implements prometheus.Collector. It never panics and always emits
// at least fortiweb_up, degrading gracefully (all-or-nothing) when any
// FortiWeb API call fails.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	status, err := c.client.GetSystemResourceStatus(ctx)
	if err != nil {
		log.Error().Str("event", "fortiweb_scrape_failed").Err(err).Msg("failed to scrape FortiWeb system resource status")
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		return
	}

	values := make([]float64, len(c.counts))
	for i, m := range c.counts {
		n, err := m.fetch(ctx)
		if err != nil {
			log.Error().Str("event", "fortiweb_scrape_failed").Str("metric", m.name).Err(err).Msg("failed to scrape FortiWeb resource count")
			ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
			return
		}
		values[i] = float64(n)
	}

	c.collectStatus(ch, status)
	for i, m := range c.counts {
		ch <- prometheus.MustNewConstMetric(m.desc, prometheus.GaugeValue, values[i], c.vdom)
	}
}

func (c *Collector) collectStatus(ch chan<- prometheus.Metric, status *fortiweb.SystemResourceStatus) {
	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 1)
	ch <- prometheus.MustNewConstMetric(c.cpuUsagePercent, prometheus.GaugeValue, status.CPU)
	ch <- prometheus.MustNewConstMetric(c.memoryUsagePercent, prometheus.GaugeValue, status.Mem)
	ch <- prometheus.MustNewConstMetric(c.diskUsagePercent, prometheus.GaugeValue, status.DiskUsage)
	ch <- prometheus.MustNewConstMetric(c.sessionCount, prometheus.GaugeValue, float64(status.SessionCount))
	ch <- prometheus.MustNewConstMetric(c.connectionsPerSecond, prometheus.GaugeValue, float64(status.ConnCntPerSec))
	ch <- prometheus.MustNewConstMetric(c.logDiskAvailable, prometheus.GaugeValue, availableToFloat(status.LogDisk))
	ch <- prometheus.MustNewConstMetric(c.dbStatusAvailable, prometheus.GaugeValue, availableToFloat(status.DBStatus))
}

func availableToFloat(s string) float64 {
	if s == availableStatus {
		return 1
	}
	return 0
}
