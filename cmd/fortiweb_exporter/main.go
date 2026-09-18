// Command fortiweb_exporter serves Prometheus metrics scraped from a
// FortiWeb appliance's system resource status.
package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/ndkprd/fortiweb_exporter/internal/collector"
	"github.com/ndkprd/fortiweb_exporter/internal/config"
	"github.com/ndkprd/fortiweb_exporter/internal/fortiweb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	configPath := flag.String("config", "config.yml", "path to the exporter's config.yml")
	flag.Parse()

	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal().Str("event", "config_load_failed").Err(err).Msg("failed to load config")
	}

	client := fortiweb.NewClient(cfg.FortiWeb.URL, cfg.FortiWeb.Username, cfg.FortiWeb.Password, cfg.FortiWeb.InsecureSkipVerify)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector.NewCollector(client))

	mux := http.NewServeMux()
	mux.Handle(cfg.MetricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	log.Info().
		Str("event", "fortiweb_exporter_starting").
		Str("listen_address", cfg.ListenAddress).
		Str("metrics_path", cfg.MetricsPath).
		Str("fortiweb_target", cfg.FortiWeb.URL).
		Msg("starting fortiweb_exporter")

	if err := http.ListenAndServe(cfg.ListenAddress, mux); err != nil {
		log.Fatal().Str("event", "http_server_failed").Err(err).Msg("http server stopped")
	}
}
