package handler

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ndkprd/fortiweb_exporter/internal/collector"
	"github.com/ndkprd/fortiweb_exporter/internal/fortiweb"
)

type fakeStatusGetter struct {
	status *fortiweb.SystemResourceStatus
	err    error
}

func (f *fakeStatusGetter) GetSystemResourceStatus(context.Context) (*fortiweb.SystemResourceStatus, error) {
	return f.status, f.err
}

func newTestClients() map[string]collector.StatusGetter {
	return map[string]collector.StatusGetter{
		"fwb-01.example.com": &fakeStatusGetter{
			status: &fortiweb.SystemResourceStatus{
				CPU: 17, Mem: 78, DiskUsage: 79,
				SessionCount: 21051, ConnCntPerSec: 482,
				LogDisk: "Available", DBStatus: "Available",
			},
		},
		"fwb-02.example.com": &fakeStatusGetter{
			err: errors.New("connection refused"),
		},
	}
}

func TestServeHTTP_MissingTarget(t *testing.T) {
	h := NewMetricsHandler(newTestClients())

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "target parameter is required") {
		t.Errorf("body = %q, want it to mention the missing target parameter", rec.Body.String())
	}
}

func TestServeHTTP_UnknownTarget(t *testing.T) {
	h := NewMetricsHandler(newTestClients())

	req := httptest.NewRequest("GET", "/metrics?target=does-not-exist", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "does-not-exist") {
		t.Errorf("body = %q, want it to mention the unknown target name", rec.Body.String())
	}
}

func TestServeHTTP_KnownTarget_Success(t *testing.T) {
	h := NewMetricsHandler(newTestClients())

	req := httptest.NewRequest("GET", "/metrics?target=fwb-01.example.com", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "fortiweb_up 1") {
		t.Errorf("body missing fortiweb_up 1:\n%s", body)
	}
	if !strings.Contains(body, "fortiweb_cpu_usage_percent 17") {
		t.Errorf("body missing fortiweb_cpu_usage_percent 17:\n%s", body)
	}
}

func TestServeHTTP_KnownTarget_UpstreamFailure(t *testing.T) {
	h := NewMetricsHandler(newTestClients())

	req := httptest.NewRequest("GET", "/metrics?target=fwb-02.example.com", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200 (scrape failures degrade gracefully); body: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "fortiweb_up 0") {
		t.Errorf("body missing fortiweb_up 0:\n%s", body)
	}
	if strings.Contains(body, "fortiweb_cpu_usage_percent") {
		t.Errorf("body should not contain other metrics on failure:\n%s", body)
	}
}

func TestServeHTTP_TargetIsolation(t *testing.T) {
	h := NewMetricsHandler(newTestClients())

	req := httptest.NewRequest("GET", "/metrics?target=fwb-01.example.com", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, "connection refused") {
		t.Errorf("fwb-01 response leaked fwb-02's failure:\n%s", body)
	}
}
