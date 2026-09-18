package handler

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.com/endekasoft/fortiweb-exporter/internal/fortiweb"
)

type fakeStatusGetter struct {
	status *fortiweb.SystemResourceStatus
	err    error
}

func (f *fakeStatusGetter) GetSystemResourceStatus(context.Context) (*fortiweb.SystemResourceStatus, error) {
	return f.status, f.err
}

func (f *fakeStatusGetter) GetServerPolicyCount(context.Context) (int, error) {
	return 0, f.err
}

func (f *fakeStatusGetter) GetContentRoutingCount(context.Context) (int, error) {
	return 0, f.err
}

func (f *fakeStatusGetter) GetServerPoolCount(context.Context) (int, error) {
	return 0, f.err
}

func newTestClients() map[string]Target {
	return map[string]Target{
		"fwb-01.example.com": {
			Client: &fakeStatusGetter{
				status: &fortiweb.SystemResourceStatus{
					CPU: 17, Mem: 78, DiskUsage: 79,
					SessionCount: 21051, ConnCntPerSec: 482,
					LogDisk: "Available", DBStatus: "Available",
				},
			},
			VDOM: "root",
		},
		"fwb-02.example.com": {
			Client: &fakeStatusGetter{
				err: errors.New("connection refused"),
			},
			VDOM: "root",
		},
	}
}

func TestProbeServeHTTP_MissingTarget(t *testing.T) {
	h := NewProbeHandler(newTestClients())

	req := httptest.NewRequest("GET", ProbePath, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "target parameter is required") {
		t.Errorf("body = %q, want it to mention the missing target parameter", rec.Body.String())
	}
}

func TestProbeServeHTTP_UnknownTarget(t *testing.T) {
	h := NewProbeHandler(newTestClients())

	req := httptest.NewRequest("GET", ProbePath+"?target=does-not-exist", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "does-not-exist") {
		t.Errorf("body = %q, want it to mention the unknown target name", rec.Body.String())
	}
}

func TestProbeServeHTTP_KnownTarget_Success(t *testing.T) {
	h := NewProbeHandler(newTestClients())

	req := httptest.NewRequest("GET", ProbePath+"?target=fwb-01.example.com", nil)
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

func TestProbeServeHTTP_KnownTarget_UpstreamFailure(t *testing.T) {
	h := NewProbeHandler(newTestClients())

	req := httptest.NewRequest("GET", ProbePath+"?target=fwb-02.example.com", nil)
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

func TestProbeServeHTTP_TargetIsolation(t *testing.T) {
	h := NewProbeHandler(newTestClients())

	req := httptest.NewRequest("GET", ProbePath+"?target=fwb-01.example.com", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, "connection refused") {
		t.Errorf("fwb-01 response leaked fwb-02's failure:\n%s", body)
	}
}
