// Package collector implements a prometheus.Collector that scrapes a
// FortiWeb appliance's system resource status on every collection.
package collector

import (
	"context"

	"github.com/ndkprd/fortiweb_exporter/internal/fortiweb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

const availableStatus = "Available"

// StatusGetter fetches FortiWeb system resource status. *fortiweb.Client
// satisfies this interface; tests can supply a fake.
type StatusGetter interface {
	GetSystemResourceStatus(ctx context.Context) (*fortiweb.SystemResourceStatus, error)
}

// Collector adapts a StatusGetter into a prometheus.Collector.
type Collector struct {
	client StatusGetter

	up                   *prometheus.Desc
	cpuUsagePercent      *prometheus.Desc
	memoryUsagePercent   *prometheus.Desc
	diskUsagePercent     *prometheus.Desc
	sessionCount         *prometheus.Desc
	connectionsPerSecond *prometheus.Desc
	logDiskAvailable     *prometheus.Desc
	dbStatusAvailable    *prometheus.Desc
}

// NewCollector builds a Collector that scrapes status from client.
func NewCollector(client StatusGetter) *Collector {
	return &Collector{
		client: client,
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
}

// Collect implements prometheus.Collector. It never panics and always emits
// at least fortiweb_up, degrading gracefully when the FortiWeb API call
// fails.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	status, err := c.client.GetSystemResourceStatus(context.Background())
	if err != nil {
		log.Error().Str("event", "fortiweb_scrape_failed").Err(err).Msg("failed to scrape FortiWeb system resource status")
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		return
	}

	c.collectStatus(ch, status)
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
