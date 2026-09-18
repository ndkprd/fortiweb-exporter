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
  url: "https://fortiweb.example.com"
  username: "admin"
  password: "changeme"
  insecure_skip_verify: true
listen_address: ":9999"
metrics_path: "/custom-metrics"
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	want := Config{
		FortiWeb: FortiWebConfig{
			URL:                "https://fortiweb.example.com",
			Username:           "admin",
			Password:           "changeme",
			InsecureSkipVerify: true,
		},
		ListenAddress: ":9999",
		MetricsPath:   "/custom-metrics",
	}
	if *cfg != want {
		t.Fatalf("Load() = %+v, want %+v", *cfg, want)
	}
}

func TestLoad_DefaultsApplied(t *testing.T) {
	path := writeConfigFile(t, `
fortiweb:
  url: "https://fortiweb.example.com"
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
  username: "admin"
  password: "changeme"
`,
		},
		{
			name: "missing username",
			contents: `
fortiweb:
  url: "https://fortiweb.example.com"
  password: "changeme"
`,
		},
		{
			name: "missing password",
			contents: `
fortiweb:
  url: "https://fortiweb.example.com"
  username: "admin"
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
