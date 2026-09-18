package collector

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"gitlab.com/endekasoft/fortiweb-exporter/internal/fortiweb"
)

type fakeStatusGetter struct {
	status *fortiweb.SystemResourceStatus
	err    error

	serverPolicyCount   int
	contentRoutingCount int
	serverPoolCount     int
	countErr            error
}

func (f *fakeStatusGetter) GetSystemResourceStatus(context.Context) (*fortiweb.SystemResourceStatus, error) {
	return f.status, f.err
}

func (f *fakeStatusGetter) GetServerPolicyCount(context.Context) (int, error) {
	return f.serverPolicyCount, f.countErr
}

func (f *fakeStatusGetter) GetContentRoutingCount(context.Context) (int, error) {
	return f.contentRoutingCount, f.countErr
}

func (f *fakeStatusGetter) GetServerPoolCount(context.Context) (int, error) {
	return f.serverPoolCount, f.countErr
}

func TestCollect_Success(t *testing.T) {
	fake := &fakeStatusGetter{
		status: &fortiweb.SystemResourceStatus{
			CPU:           17,
			Mem:           78,
			DiskUsage:     79,
			SessionCount:  21051,
			ConnCntPerSec: 482,
			LogDisk:       "Available",
			DBStatus:      "Available",
		},
		serverPolicyCount:   5,
		contentRoutingCount: 6,
		serverPoolCount:     7,
	}
	collector := NewCollector(fake, "root")

	expected := `
# HELP fortiweb_up Whether the last scrape of the FortiWeb API succeeded (1 for success, 0 for failure).
# TYPE fortiweb_up gauge
fortiweb_up 1
# HELP fortiweb_cpu_usage_percent FortiWeb CPU usage percentage.
# TYPE fortiweb_cpu_usage_percent gauge
fortiweb_cpu_usage_percent 17
# HELP fortiweb_memory_usage_percent FortiWeb memory usage percentage.
# TYPE fortiweb_memory_usage_percent gauge
fortiweb_memory_usage_percent 78
# HELP fortiweb_disk_usage_percent FortiWeb disk usage percentage.
# TYPE fortiweb_disk_usage_percent gauge
fortiweb_disk_usage_percent 79
# HELP fortiweb_session_count Current FortiWeb session count.
# TYPE fortiweb_session_count gauge
fortiweb_session_count 21051
# HELP fortiweb_connections_per_second FortiWeb connections per second.
# TYPE fortiweb_connections_per_second gauge
fortiweb_connections_per_second 482
# HELP fortiweb_log_disk_available Whether the FortiWeb log disk is available (1) or not (0).
# TYPE fortiweb_log_disk_available gauge
fortiweb_log_disk_available 1
# HELP fortiweb_db_status_available Whether the FortiWeb database status is available (1) or not (0).
# TYPE fortiweb_db_status_available gauge
fortiweb_db_status_available 1
# HELP fortiweb_server_policy_count Current number of server-policy objects configured on the vdom.
# TYPE fortiweb_server_policy_count gauge
fortiweb_server_policy_count{vdom="root"} 5
# HELP fortiweb_content_routing_count Current number of content-routing-policy objects configured on the vdom.
# TYPE fortiweb_content_routing_count gauge
fortiweb_content_routing_count{vdom="root"} 6
# HELP fortiweb_server_pool_count Current number of server-pool objects configured on the vdom.
# TYPE fortiweb_server_pool_count gauge
fortiweb_server_pool_count{vdom="root"} 7
`

	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected)); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}
}

func TestCollect_Failure(t *testing.T) {
	fake := &fakeStatusGetter{err: errors.New("connection refused")}
	collector := NewCollector(fake, "root")

	expected := `
# HELP fortiweb_up Whether the last scrape of the FortiWeb API succeeded (1 for success, 0 for failure).
# TYPE fortiweb_up gauge
fortiweb_up 0
`

	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected)); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}
}

func TestCollect_ResourceCountFailure(t *testing.T) {
	fake := &fakeStatusGetter{
		status: &fortiweb.SystemResourceStatus{
			CPU: 17,
		},
		countErr: errors.New("connection refused"),
	}
	collector := NewCollector(fake, "root")

	expected := `
# HELP fortiweb_up Whether the last scrape of the FortiWeb API succeeded (1 for success, 0 for failure).
# TYPE fortiweb_up gauge
fortiweb_up 0
`

	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected)); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}
}
