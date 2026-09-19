# AGENTS.md

Guidance for AI coding agents working in this repo. Human contributors: see
[CONTRIBUTING.md](CONTRIBUTING.md).

## Project

A Prometheus exporter for FortiWeb appliances, following the
`blackbox_exporter` pattern (`/probe?target=NAME` per-appliance, `/metrics`
for the exporter's own runtime metrics). Go, no web frontend.

## Layout

- `cmd/fortiweb-exporter/` — entrypoint (`main.go`).
- `internal/config/` — YAML config loading/validation (`config.yml`).
- `internal/fortiweb/` — HTTP client for the FortiWeb REST API.
- `internal/collector/` — `prometheus.Collector` adapting the client into
  metrics; all-or-nothing per scrape (any upstream error → `fortiweb_up 0`
  only, no partial metric sets — preserve this convention).
- `internal/handler/` — HTTP handlers (`/`, `/probe`, `/metrics`).
- `examples/` — runnable exporter + Prometheus + Grafana stack via
  `docker-compose`; `examples/grafana/dashboards/` has the example dashboard.
- `.agents/plans/` — numbered implementation plans (`NNN-slug.md`); check
  here for prior design decisions before making structural changes.

## Build / test / verify

```sh
go build ./...
go vet ./...
gofmt -l .      # must be empty; gofmt -w <file> to fix
go test ./...
```

Run all four before considering a change done.

## Conventions

- Every `StatusGetter` method addition (new client call) needs: the client
  method + its unit test, the interface method on
  `collector.StatusGetter`, wiring into `Collector.Describe`/`Collect`, a
  collector unit test, and — if the plan calls for it — a `README.md`
  metrics-table row and a Grafana dashboard panel. See
  `.agents/plans/003-add-fortiguard-license-metrics.md` for a worked example
  of this full chain.
- New metrics: no vdom label unless the data is genuinely per-vdom (matches
  existing resource-count metrics); appliance-global metrics carry no
  labels beyond what's needed to distinguish services/resources.
- Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)
  — `CHANGELOG.md` is generated from them via `git-cliff` (`cliff.toml`);
  non-conventional commits are dropped from it silently.
- Every change should trace back to a GitLab issue (see CONTRIBUTING.md) —
  don't start work on unassigned issues.
