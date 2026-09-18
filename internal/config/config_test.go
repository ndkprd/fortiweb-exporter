package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfigFile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("writing test config file: %v", err)
	}
	return path
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeConfigFile(t, `
fortiweb:
  fwb-01.example.com:
    url: "https://fortiweb-01.example.com"
    username: "admin"
    password: "changeme"
    vdom: "customvdom"
    insecure_skip_verify: true
  fwb-02.example.com:
    url: "https://fortiweb-02.example.com"
    username: "admin2"
    password: "changeme2"
    insecure_skip_verify: false
listen_address: ":9999"
metrics_path: "/custom-metrics"
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	want := Config{
		FortiWeb: map[string]FortiWebConfig{
			"fwb-01.example.com": {
				URL:                "https://fortiweb-01.example.com",
				Username:           "admin",
				Password:           "changeme",
				VDOM:               "customvdom",
				InsecureSkipVerify: true,
			},
			"fwb-02.example.com": {
				URL:                "https://fortiweb-02.example.com",
				Username:           "admin2",
				Password:           "changeme2",
				VDOM:               defaultVDOM,
				InsecureSkipVerify: false,
			},
		},
		ListenAddress: ":9999",
		MetricsPath:   "/custom-metrics",
	}
	if len(cfg.FortiWeb) != len(want.FortiWeb) {
		t.Fatalf("Load() FortiWeb = %+v, want %+v", cfg.FortiWeb, want.FortiWeb)
	}
	for name, wantTarget := range want.FortiWeb {
		gotTarget, ok := cfg.FortiWeb[name]
		if !ok {
			t.Fatalf("Load() missing target %q", name)
		}
		if gotTarget != wantTarget {
			t.Errorf("Load() target %q = %+v, want %+v", name, gotTarget, wantTarget)
		}
	}
	if cfg.ListenAddress != want.ListenAddress {
		t.Errorf("ListenAddress = %q, want %q", cfg.ListenAddress, want.ListenAddress)
	}
	if cfg.MetricsPath != want.MetricsPath {
		t.Errorf("MetricsPath = %q, want %q", cfg.MetricsPath, want.MetricsPath)
	}
}

func TestLoad_DefaultsApplied(t *testing.T) {
	path := writeConfigFile(t, `
fortiweb:
  fwb-01.example.com:
    url: "https://fortiweb-01.example.com"
    username: "admin"
    password: "changeme"
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.ListenAddress != defaultListenAddress {
		t.Errorf("ListenAddress = %q, want default %q", cfg.ListenAddress, defaultListenAddress)
	}
	if cfg.MetricsPath != defaultMetricsPath {
		t.Errorf("MetricsPath = %q, want default %q", cfg.MetricsPath, defaultMetricsPath)
	}
	if got := cfg.FortiWeb["fwb-01.example.com"].VDOM; got != defaultVDOM {
		t.Errorf("VDOM = %q, want default %q", got, defaultVDOM)
	}
}

func TestLoad_EmptyTargetMap(t *testing.T) {
	tests := []struct {
		name     string
		contents string
	}{
		{name: "fortiweb key omitted", contents: `listen_address: ":9633"`},
		{name: "fortiweb key empty map", contents: "fortiweb: {}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfigFile(t, tt.contents)

			if _, err := Load(path); err == nil {
				t.Fatal("Load() returned nil error, want non-nil for an empty target map")
			}
		})
	}
}

func TestLoad_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		contents string
	}{
		{
			name: "missing url",
			contents: `
fortiweb:
  fwb-01.example.com:
    username: "admin"
    password: "changeme"
`,
		},
		{
			name: "missing username",
			contents: `
fortiweb:
  fwb-01.example.com:
    url: "https://fortiweb-01.example.com"
    password: "changeme"
`,
		},
		{
			name: "missing password",
			contents: `
fortiweb:
  fwb-01.example.com:
    url: "https://fortiweb-01.example.com"
    username: "admin"
`,
		},
		{
			name: "one valid target, one invalid target",
			contents: `
fortiweb:
  fwb-01.example.com:
    url: "https://fortiweb-01.example.com"
    username: "admin"
    password: "changeme"
  fwb-02.example.com:
    url: "https://fortiweb-02.example.com"
    username: "admin2"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfigFile(t, tt.contents)

			if _, err := Load(path); err == nil {
				t.Fatal("Load() returned nil error, want non-nil for missing required field")
			}
		})
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	nonexistent := filepath.Join(t.TempDir(), "does-not-exist.yml")

	if _, err := Load(nonexistent); err == nil {
		t.Fatal("Load() returned nil error, want non-nil for a nonexistent file")
	}
}
