# fortiweb_exporter

A Prometheus exporter for [FortiWeb](https://www.fortinet.com/products/web-application-firewall/fortiweb)
appliances. It scrapes the `api/v2.0/system/status.systemresource` REST
endpoint on every Prometheus scrape and exposes CPU, memory, disk, session,
and connection-rate metrics.

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

When the FortiWeb API call fails (network error, auth failure, non-2xx
response), `/metrics` still returns `200 OK` with only `fortiweb_up 0` set —
the other metrics are simply omitted for that scrape, following Prometheus's
standard "target down" convention.

## Configuration

The exporter reads a single target's connection details and credentials from
a YAML config file (default path `config.yml`, overridable with `--config`).
See [`config.yml.example`](config.yml.example):

```yaml
fortiweb:
  url: "https://fortiweb.example.com"
  username: "admin"
  password: "changeme"
  insecure_skip_verify: false
listen_address: ":9633"
metrics_path: "/metrics"
```

| Field | Required | Default | Description |
|---|---|---|---|
| `fortiweb.url` | yes | — | Base URL of the FortiWeb appliance's REST API. |
| `fortiweb.username` | yes | — | Admin username, sent via HTTP Basic Auth. |
| `fortiweb.password` | yes | — | Admin password, sent via HTTP Basic Auth. |
| `fortiweb.insecure_skip_verify` | no | `false` | Skip TLS certificate verification — useful for appliances with self-signed certs. |
| `listen_address` | no | `:9633` | Address the exporter's HTTP server listens on. |
| `metrics_path` | no | `/metrics` | Path the Prometheus metrics are served on. |

`config.yml` is gitignored — never commit real credentials. Copy the example
file and edit it:

```sh
cp config.yml.example config.yml
```

## Usage

Build and run locally:

```sh
go build -o fortiweb_exporter ./cmd/fortiweb_exporter
./fortiweb_exporter --config config.yml
```

Or run directly without a separate build step:

```sh
go run ./cmd/fortiweb_exporter --config config.yml
```

Then scrape it:

```sh
curl http://localhost:9633/metrics
```

Point Prometheus at it with a `scrape_config` like:

```yaml
scrape_configs:
  - job_name: fortiweb
    static_configs:
      - targets: ["localhost:9633"]
```

## Docker

Build the image:

```sh
docker build -t fortiweb_exporter .
```

Run it, mounting your `config.yml` read-only and publishing the metrics port:

```sh
docker run --rm \
  -v "$(pwd)/config.yml:/etc/fortiweb_exporter/config.yml:ro" \
  -p 9633:9633 \
  fortiweb_exporter
```

## Development

```sh
go build ./...   # build everything
go vet ./...     # static checks
go test ./...    # unit tests
```

## License

[MIT](LICENSE)
