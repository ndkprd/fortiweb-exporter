package collector

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"gitlab.com/endekasoft/fortiweb-exporter/internal/fortiweb"
)

type fakeStatusGetter struct {
	status *fortiweb.SystemResourceStatus
	err    error

	serverPolicyCount      int
	contentRoutingCount    int
	serverPoolCount        int
	protectionProfileCount int
	allowedHostsCount      int
	virtualServerCount     int
	signatureCount         int
	countErr               error

	fortiGuardStatus *fortiweb.FortiGuardStatus
	fortiGuardErr    error
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

func (f *fakeStatusGetter) GetProtectionProfileCount(context.Context) (int, error) {
	return f.protectionProfileCount, f.countErr
}

func (f *fakeStatusGetter) GetAllowedHostsCount(context.Context) (int, error) {
	return f.allowedHostsCount, f.countErr
}

func (f *fakeStatusGetter) GetVirtualServerCount(context.Context) (int, error) {
	return f.virtualServerCount, f.countErr
}

func (f *fakeStatusGetter) GetSignatureCount(context.Context) (int, error) {
	return f.signatureCount, f.countErr
}

func (f *fakeStatusGetter) GetFortiGuardStatus(context.Context) (*fortiweb.FortiGuardStatus, error) {
	return f.fortiGuardStatus, f.fortiGuardErr
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
		serverPolicyCount:      5,
		contentRoutingCount:    6,
		serverPoolCount:        7,
		protectionProfileCount: 8,
		allowedHostsCount:      9,
		virtualServerCount:     10,
		signatureCount:         11,
		fortiGuardStatus:       fakeFortiGuardStatus(),
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
# HELP fortiweb_protection_profile_count Current number of protection-profile objects configured on the vdom.
# TYPE fortiweb_protection_profile_count gauge
fortiweb_protection_profile_count{vdom="root"} 8
# HELP fortiweb_allowed_hosts_count Current number of allowed-hosts objects configured on the vdom.
# TYPE fortiweb_allowed_hosts_count gauge
fortiweb_allowed_hosts_count{vdom="root"} 9
# HELP fortiweb_virtual_server_count Current number of virtual-server objects configured on the vdom.
# TYPE fortiweb_virtual_server_count gauge
fortiweb_virtual_server_count{vdom="root"} 10
# HELP fortiweb_signature_count Current number of signature objects configured on the vdom.
# TYPE fortiweb_signature_count gauge
fortiweb_signature_count{vdom="root"} 11
# HELP fortiweb_registered Whether the FortiWeb appliance is registered with FortiGuard (1) or not (0).
# TYPE fortiweb_registered gauge
fortiweb_registered 1
# HELP fortiweb_license_expiry_timestamp_seconds Unix timestamp when the FortiGuard license for a service expires.
# TYPE fortiweb_license_expiry_timestamp_seconds gauge
fortiweb_license_expiry_timestamp_seconds{service="security"} 1.8264096e+09
fortiweb_license_expiry_timestamp_seconds{service="antivirus"} 1.8264096e+09
fortiweb_license_expiry_timestamp_seconds{service="reputation"} 1.8264096e+09
fortiweb_license_expiry_timestamp_seconds{service="credential_stuffing_defense"} 0
fortiweb_license_expiry_timestamp_seconds{service="sbcl"} 0
fortiweb_license_expiry_timestamp_seconds{service="dlp_signature"} 0
fortiweb_license_expiry_timestamp_seconds{service="geodb"} 1.8264096e+09
fortiweb_license_expiry_timestamp_seconds{service="fuzzy_webshell"} 1.8264096e+09
fortiweb_license_expiry_timestamp_seconds{service="threat_analytics"} 0
fortiweb_license_expiry_timestamp_seconds{service="advanced_bot_protection"} 0
# HELP fortiweb_license_valid Whether the FortiGuard license for a service is currently valid (1) or not (0).
# TYPE fortiweb_license_valid gauge
fortiweb_license_valid{service="security"} 1
fortiweb_license_valid{service="antivirus"} 1
fortiweb_license_valid{service="reputation"} 1
fortiweb_license_valid{service="credential_stuffing_defense"} 0
fortiweb_license_valid{service="sbcl"} 0
fortiweb_license_valid{service="dlp_signature"} 0
fortiweb_license_valid{service="geodb"} 1
fortiweb_license_valid{service="fuzzy_webshell"} 1
fortiweb_license_valid{service="threat_analytics"} 0
fortiweb_license_valid{service="advanced_bot_protection"} 0
`

	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected)); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}
}

func fakeFortiGuardStatus() *fortiweb.FortiGuardStatus {
	licensed := time.Date(2027, time.November, 17, 0, 0, 0, 0, time.UTC)
	unlicensed := time.Unix(0, 0).UTC()
	return &fortiweb.FortiGuardStatus{
		IsRegistered: true,
		Licenses: []fortiweb.LicenseStatus{
			{Service: "security", Expiry: licensed, Valid: true},
			{Service: "antivirus", Expiry: licensed, Valid: true},
			{Service: "reputation", Expiry: licensed, Valid: true},
			{Service: "credential_stuffing_defense", Expiry: unlicensed, Valid: false},
			{Service: "sbcl", Expiry: unlicensed, Valid: false},
			{Service: "dlp_signature", Expiry: unlicensed, Valid: false},
			{Service: "geodb", Expiry: licensed, Valid: true},
			{Service: "fuzzy_webshell", Expiry: licensed, Valid: true},
			{Service: "threat_analytics", Expiry: unlicensed, Valid: false},
			{Service: "advanced_bot_protection", Expiry: unlicensed, Valid: false},
		},
	}
}

func TestCollect_FortiGuardFailure(t *testing.T) {
	fake := &fakeStatusGetter{
		status:        &fortiweb.SystemResourceStatus{CPU: 17},
		fortiGuardErr: errors.New("connection refused"),
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
