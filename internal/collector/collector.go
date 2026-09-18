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
	serverPolicyCount    *prometheus.Desc
	contentRoutingCount  *prometheus.Desc
	serverPoolCount      *prometheus.Desc
}

// NewCollector builds a Collector that scrapes status from client, labeling
// its resource-count gauges with vdom.
func NewCollector(client StatusGetter, vdom string) *Collector {
	return &Collector{
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
		serverPolicyCount: prometheus.NewDesc(
			"fortiweb_server_policy_count",
			"Current number of server-policy objects configured on the vdom.",
			[]string{"vdom"}, nil,
		),
		contentRoutingCount: prometheus.NewDesc(
			"fortiweb_content_routing_count",
			"Current number of content-routing-policy objects configured on the vdom.",
			[]string{"vdom"}, nil,
		),
		serverPoolCount: prometheus.NewDesc(
			"fortiweb_server_pool_count",
			"Current number of server-pool objects configured on the vdom.",
			[]string{"vdom"}, nil,
		),
	}
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
	ch <- c.serverPolicyCount
	ch <- c.contentRoutingCount
	ch <- c.serverPoolCount
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

	serverPolicyCount, err := c.client.GetServerPolicyCount(ctx)
	if err != nil {
		log.Error().Str("event", "fortiweb_scrape_failed").Err(err).Msg("failed to scrape FortiWeb server-policy count")
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		return
	}

	contentRoutingCount, err := c.client.GetContentRoutingCount(ctx)
	if err != nil {
		log.Error().Str("event", "fortiweb_scrape_failed").Err(err).Msg("failed to scrape FortiWeb content-routing count")
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		return
	}

	serverPoolCount, err := c.client.GetServerPoolCount(ctx)
	if err != nil {
		log.Error().Str("event", "fortiweb_scrape_failed").Err(err).Msg("failed to scrape FortiWeb server-pool count")
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		return
	}

	c.collectStatus(ch, status)
	ch <- prometheus.MustNewConstMetric(c.serverPolicyCount, prometheus.GaugeValue, float64(serverPolicyCount), c.vdom)
	ch <- prometheus.MustNewConstMetric(c.contentRoutingCount, prometheus.GaugeValue, float64(contentRoutingCount), c.vdom)
	ch <- prometheus.MustNewConstMetric(c.serverPoolCount, prometheus.GaugeValue, float64(serverPoolCount), c.vdom)
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
