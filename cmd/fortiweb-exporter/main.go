// Command fortiweb-exporter serves Prometheus metrics scraped from a
// FortiWeb appliance's system resource status.
package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/endekasoft/fortiweb-exporter/internal/collector"
	"gitlab.com/endekasoft/fortiweb-exporter/internal/config"
	"gitlab.com/endekasoft/fortiweb-exporter/internal/fortiweb"
	"gitlab.com/endekasoft/fortiweb-exporter/internal/handler"
)

func main() {
	configPath := flag.String("config", "config.yml", "path to the exporter's config.yml")
	flag.Parse()

	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal().Str("event", "config_load_failed").Err(err).Msg("failed to load config")
	}

	clients := make(map[string]collector.StatusGetter, len(cfg.FortiWeb))
	targetNames := make([]string, 0, len(cfg.FortiWeb))
	for name, target := range cfg.FortiWeb {
		clients[name] = fortiweb.NewClient(target.URL, target.Username, target.Password, target.VDOM, target.InsecureSkipVerify)
		targetNames = append(targetNames, name)
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.MetricsPath, handler.NewMetricsHandler(clients))
	mux.Handle("/", handler.NewIndexHandler(targetNames, cfg.MetricsPath))

	log.Info().
		Str("event", "fortiweb_exporter_starting").
		Str("listen_address", cfg.ListenAddress).
		Str("metrics_path", cfg.MetricsPath).
		Int("target_count", len(clients)).
		Msg("starting fortiweb-exporter")

	if err := http.ListenAndServe(cfg.ListenAddress, mux); err != nil {
		log.Fatal().Str("event", "http_server_failed").Err(err).Msg("http server stopped")
	}
}
