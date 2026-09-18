# fortiweb-exporter examples

A ready-to-run stack — fortiweb-exporter, Prometheus, and Grafana — wired
together with Docker Compose. Useful for trying the exporter out quickly, or
as a starting point for your own deployment.

## Quick start

1. Edit [`config.yml`](config.yml) with your real FortiWeb appliance(s):
   `url`, `username`, `password`, and `vdom` (use `"root"` if you don't use
   virtual domains). Add or remove target entries as needed — the map key
   (e.g. `fwb-01.example.com`) is what you'll see as the `instance` label in
   Prometheus.

2. Update [`prometheus.yml`](prometheus.yml)'s `fortiweb` job so its
   `targets` list matches the target names you used in `config.yml`.

3. Start the stack:

   ```sh
   docker compose up -d
   ```

   This pulls `registry.gitlab.com/endekasoft/fortiweb-exporter:latest` (no
   local build required), `prom/prometheus:latest`, and `grafana/grafana:latest`.

4. Open:
   - Exporter landing page: <http://localhost:9633/> — lists every
     configured target with a link to its `/probe` endpoint.
   - Prometheus UI: <http://localhost:9090/targets> — confirm both the
     `fortiweb` job (one series per target, via `/probe`) and the
     `fortiweb-exporter` job (the exporter's own health, via `/metrics`)
     show as `UP`.
   - Grafana: <http://localhost:3000> — log in with the default `admin` /
     `admin` credentials (Grafana will prompt you to change it). The
     **FortiWeb Exporter** dashboard is already provisioned and selects the
     Prometheus datasource automatically.

5. Try a query in Prometheus, e.g. `fortiweb_cpu_usage_percent`.

## Files

| File | Purpose |
|---|---|
| `docker-compose.yml` | Runs the published exporter image alongside Prometheus and Grafana, all configured via the files below. |
| `config.yml` | Exporter config — **edit with your real FortiWeb targets before use.** Contains placeholder credentials only; never commit real ones. |
| `prometheus.yml` | Scrapes each FortiWeb target through the exporter's `/probe` endpoint using the standard multi-target relabeling pattern, plus a second job for the exporter's own `/metrics`. |
| `grafana/dashboards/fortiweb-exporter.json` | The **FortiWeb Exporter** dashboard, auto-provisioned into Grafana on startup. |
| `grafana/provisioning/` | Grafana's datasource (points at the `prometheus` service) and dashboard-provider config — no manual setup needed. |

See the [main README](../README.md) for the full configuration reference and
how each endpoint behaves.

## Dashboard

The provisioned **FortiWeb Exporter** dashboard covers all 8 `/probe`
metrics: target/log-disk/DB-status stats up top, CPU/memory/disk gauges and
trend graphs, and session/connection-rate traffic graphs — filterable by an
`instance` template variable. It uses a blue/yellow/orange threshold color
scheme (rather than the usual green/yellow/red) for utilization and status
coloring.

## Tear down

```sh
docker compose down
```
