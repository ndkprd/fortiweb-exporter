# fortiweb-exporter

A Prometheus exporter for [FortiWeb](https://www.fortinet.com/products/web-application-firewall/fortiweb)
appliances. It scrapes the `api/v2.0/system/status.systemresource` REST
endpoint and exposes CPU, memory, disk, session, and connection-rate metrics.
It supports multiple FortiWeb appliances from a single exporter instance —
each scrape selects which one via a `target` query parameter
(`/metrics?target=NAME`), the same pattern used by `blackbox_exporter`.

## Metrics

| Metric | Type | Description |
|---|---|---|
| `fortiweb_up` | gauge | Whether the last scrape of the FortiWeb API succeeded (`1`) or failed (`0`). |
| `fortiweb_cpu_usage_percent` | gauge | FortiWeb CPU usage percentage. |
| `fortiweb_memory_usage_percent` | gauge | FortiWeb memory usage percentage. |
| `fortiweb_disk_usage_percent` | gauge | FortiWeb disk usage percentage. |
| `fortiweb_session_count` | gauge | Current FortiWeb session count. |
| `fortiweb_connections_per_second` | gauge | FortiWeb connections per second. |
| `fortiweb_log_disk_available` | gauge | Whether the FortiWeb log disk is available (`1`) or not (`0`). |
| `fortiweb_db_status_available` | gauge | Whether the FortiWeb database status is available (`1`) or not (`0`). |

When a `target` param is missing or doesn't match a configured target,
`/metrics` returns `400 Bad Request`. When the `target` is valid but the
FortiWeb API call itself fails (network error, auth failure, non-2xx
response), `/metrics` still returns `200 OK` with only `fortiweb_up 0` set —
the other metrics are simply omitted for that scrape, following Prometheus's
standard "target down" convention.

## Configuration

The exporter reads a map of named FortiWeb targets — each with its own
connection details and credentials — from a YAML config file (default path
`config.yml`, overridable with `--config`). The map key is an arbitrary name
you choose (typically a hostname); it's what you pass as `?target=` and what
Prometheus's scrape config uses to select which appliance to scrape. See
[`config.yml.example`](config.yml.example):

```yaml
fortiweb:
  fwb-01.example.com:
    url: "https://10.0.1.10"
    username: "admin"
    password: "changeme"
    vdom: "root"
    insecure_skip_verify: true
  fwb-02.example.com:
    url: "https://10.0.2.10"
    username: "admin"
    password: "changeme"
    vdom: "root"
    insecure_skip_verify: true
listen_address: ":9633"
metrics_path: "/metrics"
```

| Field | Required | Default | Description |
|---|---|---|---|
| `fortiweb.<name>.url` | yes | — | Base URL of that FortiWeb appliance's REST API. |
| `fortiweb.<name>.username` | yes | — | Admin username. |
| `fortiweb.<name>.password` | yes | — | Admin password. |
| `fortiweb.<name>.vdom` | no | `root` | Virtual domain to authenticate into — use `root` if the appliance doesn't use VDOMs. |
| `fortiweb.<name>.insecure_skip_verify` | no | `false` | Skip TLS certificate verification — useful for appliances with self-signed certs. |
| `listen_address` | no | `:9633` | Address the exporter's HTTP server listens on. |
| `metrics_path` | no | `/metrics` | Path the Prometheus metrics are served on. |

At least one target must be configured. `config.yml` is gitignored — never
commit real credentials. Copy the example file and edit it:

```sh
cp config.yml.example config.yml
```

## Usage

Build and run locally:

```sh
go build -o fortiweb-exporter ./cmd/fortiweb-exporter
./fortiweb-exporter --config config.yml
```

Or run directly without a separate build step:

```sh
go run ./cmd/fortiweb-exporter --config config.yml
```

The root path (`http://localhost:9633/`) serves a basic landing page listing
every configured target as a link to its `/metrics?target=NAME` endpoint.

Or scrape a specific target directly:

```sh
curl 'http://localhost:9633/metrics?target=fwb-01.example.com'
```

Point Prometheus at it using the standard multi-target relabeling pattern,
so each configured target name becomes its own scraped `instance`:

```yaml
scrape_configs:
  - job_name: fortiweb
    static_configs:
      - targets:
          - fwb-01.example.com
          - fwb-02.example.com
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9633 # the fortiweb-exporter's own address
```

## Docker

Build the image:

```sh
docker build -t fortiweb-exporter .
```

Run it, mounting your `config.yml` read-only and publishing the metrics port:

```sh
docker run --rm \
  -v "$(pwd)/config.yml:/etc/fortiweb-exporter/config.yml:ro" \
  -p 9633:9633 \
  fortiweb-exporter
```

Or use Docker Compose, which builds the image and mounts `config.yml` for you:

```sh
docker compose up -d
```

## Development

```sh
go build ./...   # build everything
go vet ./...     # static checks
go test ./...    # unit tests
```

## License

[MIT](LICENSE)
