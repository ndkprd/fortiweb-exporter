// Package config loads and validates the exporter's YAML configuration.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	defaultListenAddress = ":9633"
	defaultMetricsPath   = "/metrics"
	defaultVDOM          = "root"
)

// Config is the top-level exporter configuration.
type Config struct {
	FortiWeb      map[string]FortiWebConfig `yaml:"fortiweb"`
	ListenAddress string                    `yaml:"listen_address"`
	MetricsPath   string                    `yaml:"metrics_path"`
}

// FortiWebConfig holds the target FortiWeb appliance's connection details.
type FortiWebConfig struct {
	URL                string `yaml:"url"`
	Username           string `yaml:"username"`
	Password           string `yaml:"password"`
	VDOM               string `yaml:"vdom"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
}

// Load reads and parses the YAML config file at path, applying defaults and
// validating that the required FortiWeb credentials are present.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validating config file %q: %w", path, err)
	}

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.ListenAddress == "" {
		c.ListenAddress = defaultListenAddress
	}
	if c.MetricsPath == "" {
		c.MetricsPath = defaultMetricsPath
	}
	for name, target := range c.FortiWeb {
		if target.VDOM == "" {
			target.VDOM = defaultVDOM
			c.FortiWeb[name] = target
		}
	}
}

func (c *Config) validate() error {
	if len(c.FortiWeb) == 0 {
		return fmt.Errorf("at least one fortiweb target must be configured")
	}
	for name, target := range c.FortiWeb {
		if err := target.validate(); err != nil {
			return fmt.Errorf("target %q: %w", name, err)
		}
	}
	return nil
}

func (t FortiWebConfig) validate() error {
	if t.URL == "" {
		return fmt.Errorf("url is required")
	}
	if t.Username == "" {
		return fmt.Errorf("username is required")
	}
	if t.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}
