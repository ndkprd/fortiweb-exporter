# Plan: FortiWeb Prometheus Exporter (basic system-resource metrics)

## Overview
Build a small Go binary that queries a single FortiWeb appliance's
`api/v2.0/system/status.systemresource` REST endpoint on every Prometheus scrape and
exposes CPU, memory, disk, session, and connection-rate metrics on a `/metrics` HTTP
endpoint. Authentication credentials and target configuration live in a `config.yml`
file; auth is done with HTTP Basic Auth against the FortiWeb REST API.

### Flowchart
```mermaid
flowchart TD
    A[Prometheus server] -->|HTTP GET /metrics| B[fortiweb_exporter]
    B --> C{Load config.yml}
    C -->|target, creds, TLS opt| D[FortiWeb API client]
    D -->|GET /api/v2.0/system/status.systemresource\nAuthorization: Basic user:pass| E[FortiWeb appliance]
    E -->|JSON: cpu, mem, diskUsage,\nsessionCount, connCntPerSec,\nlogDisk, dbStatus| D
    D --> F[Prometheus Collector]
    F -->|gauges: fortiweb_up,\nfortiweb_cpu_usage_percent, ...| B
    B -->|metrics text/plain| A
```

### Sequence
```mermaid
sequenceDiagram
    participant P as Prometheus
    participant E as fortiweb_exporter
    participant F as FortiWeb API

    P->>E: GET /metrics
    E->>F: GET /api/v2.0/system/status.systemresource\n(Authorization: Basic base64(user:pass))
    alt success
        F-->>E: 200 OK { results: { cpu, mem, diskUsage, sessionCount, connCntPerSec, logDisk, dbStatus } }
        E-->>P: 200 OK metrics (fortiweb_up 1, fortiweb_cpu_usage_percent ...)
    else failure (network/auth/5xx)
        F-->>E: error / non-200
        E-->>P: 200 OK metrics (fortiweb_up 0 only)
    end
```

## Context
This is a brand-new, empty repository (`/home/ndkprd/devel/ndkprd/fortiweb_exporter`,
not yet a git repo) — there is no existing code, no prior plans, and no established
project conventions to follow. This plan establishes the initial structure from
scratch.

Reference for the FortiWeb REST API auth shape: the
[fortinet-ansible-dev FortiWeb Ansible collection](https://github.com/fortinet-ansible-dev/ansible-galaxy-fortiweb-collection)
authenticates to the FortiWeb REST API using the device admin username/password.
Per user decision below, this exporter uses **HTTP Basic Auth** (`Authorization: Basic
base64(username:password)`) on each request rather than the collection's
session-cookie + CSRF-token login flow, since FortiWeb's REST API supports Basic Auth
directly and it keeps the exporter fully stateless between scrapes.

FortiWeb appliances commonly run self-signed TLS certificates out of the box, so the
HTTP client must support an opt-in, config-driven `insecure_skip_verify` flag.

## Objective
Running `fortiweb_exporter --config config.yml` starts an HTTP server that:
1. Loads target URL, username, password, and TLS option from `config.yml`.
2. On every `GET /metrics`, calls the FortiWeb `system/status.systemresource`
   endpoint using HTTP Basic Auth.
3. Exposes these Prometheus gauges: `fortiweb_up`, `fortiweb_cpu_usage_percent`,
   `fortiweb_memory_usage_percent`, `fortiweb_disk_usage_percent`,
   `fortiweb_session_count`, `fortiweb_connections_per_second`,
   `fortiweb_log_disk_available`, `fortiweb_db_status_available`.
4. Degrades gracefully: if the FortiWeb API call fails, `/metrics` still returns
   `200 OK` with only `fortiweb_up 0` set (Prometheus's standard "target failed"
   convention), rather than erroring the whole scrape.

Done = `go build ./...` succeeds, `go test ./...` passes (including the FortiWeb
client unit tests), a manually-run binary serves valid Prometheus exposition format
on `/metrics`, and `docker build` produces a working image.

## References
- No existing files — greenfield project. All paths below are new files this plan
  creates.
- `https://github.com/fortinet-ansible-dev/ansible-galaxy-fortiweb-collection` —
  reference for FortiWeb REST API username/password auth shape (used to validate the
  request/response contract, not copied directly since this plan uses Basic Auth
  instead of the collection's session+CSRF flow).
- `https://github.com/prometheus/client_golang` — official Prometheus Go
  instrumentation library, used for the collector and `/metrics` HTTP handler.

## Constraints / Scope

**In scope:**
- Single FortiWeb target, configured via `config.yml` (no multi-target/multi-device
  support yet).
- Only the `api/v2.0/system/status.systemresource` endpoint.
- HTTP Basic Auth against the FortiWeb REST API.
- Config-driven `insecure_skip_verify` TLS option (default `false`).
- On-demand query per scrape (no background poller/cache).
- `go.mod`-based Go module, `github.com/prometheus/client_golang` for metrics,
  `gopkg.in/yaml.v3` for config parsing.
- Sample `config.yml.example` + `README.md` documenting setup/build/run.
- `Dockerfile` for containerized builds.
- Unit tests for the FortiWeb API client (login/request + response parsing) using
  `net/http/httptest`.
- MIT `LICENSE` file.

**Out of scope:**
- Multiple FortiWeb targets / multi-target scraping (`?target=` param pattern).
- Any FortiWeb API endpoint other than `system/status.systemresource`.
- Session-cookie/CSRF-token auth flow (explicitly deferred in favor of Basic Auth).
- Background polling, caching, or configurable scrape intervals inside the exporter
  itself (Prometheus's own scrape_interval governs frequency).
- Metrics persistence, alerting rules, or Grafana dashboards.
- CI/CD pipeline setup, releases, or Helm charts.

**Non-negotiables:**
- Credentials (username/password) must only ever be read from `config.yml`, never
  hardcoded or passed as CLI flags (avoids leaking secrets into shell history/process
  list).
- `/metrics` must never panic or return a non-200 due to the upstream FortiWeb call
  failing — failures must surface as `fortiweb_up 0`.
- No secrets committed to the repo — `config.yml` (the real one) must be gitignored;
  only `config.yml.example` with placeholder values is committed.

## Tasks

### Task 1: Scaffold the Go module and project layout
- **Status**: completed
- **Date**: 2026-09-18
- **Related file**: `go.mod`, `.gitignore`, directory structure
- **Objective**: Initialize the Go module (`github.com/ndkprd/fortiweb_exporter`, Go
  1.22+), create the directory skeleton (`cmd/fortiweb_exporter/`,
  `internal/config/`, `internal/fortiweb/`, `internal/collector/`), and add a
  `.gitignore` that excludes `config.yml`, built binaries, and standard Go/OS cruft.
  Initialize git (`git init`) since the directory is not yet a repo.
- **Verification**: Run `go build ./...` from the repo root — no source files exist
  yet so it should succeed with no errors, and `test -f go.mod && git rev-parse
  --is-inside-work-tree` both succeed.

### Task 2: Implement config loading
- **Status**: completed
- **Date**: 2026-09-18
- **Related file**: `internal/config/config.go`, `internal/config/config_test.go`,
  `config.yml.example`
- **Objective**: Define a `Config` struct (fortiweb target URL, username, password,
  `insecure_skip_verify` bool, HTTP listen address, metrics path) and a
  `Load(path string) (*Config, error)` function that reads and parses YAML via
  `gopkg.in/yaml.v3`, applying sensible defaults (listen address `:9633`, metrics
  path `/metrics`) when omitted. Create `config.yml.example` with placeholder values
  matching the struct fields. *(depends on Task 1)*
- **Verification**: Run `go test ./internal/config/...` — a test loading
  `config.yml.example` (or an inline fixture) must assert the parsed struct fields
  match expected values, including that defaults apply when optional fields are
  omitted.

### Task 3: Implement the FortiWeb API client
- **Status**: completed
- **Date**: 2026-09-18
- **Related file**: `internal/fortiweb/client.go`
- **Objective**: Implement `fortiweb.Client` with a constructor taking base URL,
  username, password, and `insecureSkipVerify bool`. Add
  `GetSystemResourceStatus(ctx context.Context) (*SystemResourceStatus, error)` that
  issues `GET {baseURL}/api/v2.0/system/status.systemresource` with header
  `Authorization: Basic base64(username:password)`, and unmarshals the
  `{"results": {"cpu":..., "mem":..., "logDisk":..., "dbStatus":..., "diskUsage":...,
  "sessionCount":..., "connCntPerSec":...}}` envelope into a typed
  `SystemResourceStatus` struct (`CPU, Mem, DiskUsage float64`; `SessionCount int64`;
  `ConnCntPerSec int64`; `LogDisk, DBStatus string`). The underlying `http.Client`
  must use a `tls.Config{InsecureSkipVerify: insecureSkipVerify}` transport per the
  constructor flag. Non-2xx responses return a descriptive error. *(depends on
  Task 1)*
- **Verification**: `go build ./internal/fortiweb/...` succeeds and the package
  exports `Client`, `NewClient`, `SystemResourceStatus`, and
  `(*Client).GetSystemResourceStatus` (checkable via `go doc ./internal/fortiweb`).

### Task 4: Unit tests for the FortiWeb API client
- **Status**: completed
- **Date**: 2026-09-18
- **Related file**: `internal/fortiweb/client_test.go`
- **Objective**: Using `net/http/httptest.NewServer`, write tests that (a) assert the
  client sends the correct `Authorization: Basic ...` header and requests the
  correct path, (b) assert a well-formed JSON response is parsed into the expected
  `SystemResourceStatus` values matching the sample payload from this plan's
  Overview, and (c) assert a non-200 response (e.g. 401) surfaces as a non-nil error.
  *(depends on Task 3)*
- **Verification**: Run `go test ./internal/fortiweb/... -v` — all tests pass.

### Task 5: Implement the Prometheus collector
- **Status**: completed
- **Date**: 2026-09-18
- **Related file**: `internal/collector/collector.go`, `internal/collector/collector_test.go`
- **Objective**: Implement a `prometheus.Collector` that wraps a `fortiweb.Client`
  (accept it as/through a small interface so it's mockable). On `Collect`, call
  `GetSystemResourceStatus`; on success emit `fortiweb_up 1` plus gauges
  `fortiweb_cpu_usage_percent`, `fortiweb_memory_usage_percent`,
  `fortiweb_disk_usage_percent`, `fortiweb_session_count`,
  `fortiweb_connections_per_second`, and boolean-as-gauge
  `fortiweb_log_disk_available` / `fortiweb_db_status_available` (1 if the string
  value is `"Available"`, else 0). On failure, emit only `fortiweb_up 0` and log the
  error (do not panic or return an error from `Collect`). *(depends on Task 3)*
- **Verification**: Run `go test ./internal/collector/...` — a test using a fake
  client returning fixed `SystemResourceStatus` values, verified with
  `github.com/prometheus/client_golang/prometheus/testutil.CollectAndCompare`
  against an expected exposition-format fixture; a second test with a fake client
  returning an error asserts only `fortiweb_up 0` is emitted.

### Task 6: Wire up `main.go`
- **Status**: completed
- **Date**: 2026-09-18
- **Related file**: `cmd/fortiweb_exporter/main.go`
- **Objective**: Add a `--config` flag (default `config.yml`), load config via
  `internal/config`, construct the `fortiweb.Client` and collector, register the
  collector with a `prometheus.Registry`, and start an `http.Server` serving
  `promhttp.HandlerFor(registry, ...)` on the configured metrics path/address. Log
  startup info (listen address, target URL) to stdout. *(depends on Task 2, Task 5)*
- **Verification**: Run `go build -o /tmp/fortiweb_exporter ./cmd/fortiweb_exporter`
  — build succeeds. Then run it against a `config.yml` pointing at a throwaway
  `httptest`-style stub or an unreachable URL and `curl -s
  http://localhost:9633/metrics | grep fortiweb_up` — expect `fortiweb_up 0` (proves
  the server starts and degrades gracefully without a real device).

### Task 7: Add a Dockerfile
- **Status**: completed
- **Date**: 2026-09-18
- **Related file**: `Dockerfile`
- **Objective**: Write a multi-stage Dockerfile (Go build stage on
  `golang:1.22-alpine` or newer, minimal `gcr.io/distroless/static` or
  `alpine`-based final stage) that builds `cmd/fortiweb_exporter` and runs it as a
  non-root user, exposing the default metrics port. *(depends on Task 6)*
- **Verification**: Run `docker build -t fortiweb_exporter:test .` from the repo
  root — build succeeds and `docker run --rm fortiweb_exporter:test --help` (or
  equivalent) starts without a crash loop.

### Task 8: Write README.md
- **Status**: completed
- **Date**: 2026-09-18
- **Related file**: `README.md`
- **Objective**: Document what the exporter does, the exposed metric names/types,
  `config.yml` format (with reference to `config.yml.example`), how to build/run
  locally (`go build`, `go run`), how to build/run via Docker, and a sample
  `scrape_config` snippet for `prometheus.yml`. *(depends on Task 6, Task 7)*
- **Verification**: `grep -qE "^## (Configuration|Metrics|Usage|Docker)" README.md`
  succeeds for each of those section headers.

### Task 9: Add MIT LICENSE
- **Status**: completed
- **Date**: 2026-09-18
- **Related file**: `LICENSE`
- **Objective**: Add the standard MIT License text with the current year (2026) and
  copyright holder attributed to the repo owner.
- **Verification**: `grep -q "MIT License" LICENSE` succeeds.

## FAQ

1. **Is the FortiWeb `system/status.systemresource` JSON response shape confirmed?**
   The plan's Overview and Task 3 specify a flat envelope
   (`{"results": {"cpu", "mem", "diskUsage", "sessionCount", "connCntPerSec",
   "logDisk", "dbStatus"}}`) with `cpu`/`mem`/`diskUsage` as `float64` and
   `logDisk`/`dbStatus` as plain strings compared against `"Available"`. The only
   cited reference is the FortiWeb Ansible collection, which documents auth, not this
   endpoint's response body. Real FortiWeb firmware sometimes nests these values
   (e.g. per-core CPU arrays, or `mem`/`disk` as objects with multiple sub-fields)
   depending on version. Should I implement exactly the flat shape as literally
   specified in the plan (treating it as the authoritative contract for this task),
   with the understanding that it may need adjustment once tested against a real
   appliance? (Assumption if no answer: yes — implement the flat shape as specified.)

   **Answer:** Yes. Treat the flat shape as the authoritative contract for this
   plan — it's taken directly from a real response the user provided (see Q6's
   canonical fixture below), not a guess. If a real appliance later returns a
   different shape (nested per-core CPU, etc.), that's a follow-up plan, not a
   blocker here.

2. **Should the outbound FortiWeb API call have a request timeout?**
   Task 3/6 describe an on-demand call per scrape with graceful degradation to
   `fortiweb_up 0` on failure, but no timeout is specified for the HTTP client. Without
   one, a hung or slow FortiWeb device would block the `/metrics` handler
   indefinitely instead of degrading gracefully. Should the client use a fixed
   default timeout (e.g. 10s) via `http.Client{Timeout: ...}`, or should it be a
   configurable `config.yml` field? (Assumption if no answer: hardcode a reasonable
   default, e.g. 10s, on the `http.Client`.)

   **Answer:** Hardcode it — `http.Client{Timeout: 10 * time.Second}` in the
   `fortiweb.Client` constructor. Keep this plan's `config.yml` surface area minimal
   (per the original scope decision); a configurable timeout can be added later if
   it turns out to matter in practice.

3. **Should `config.Load` validate required fields?**
   Task 2 doesn't say whether `Load` should return an error when `target`,
   `username`, or `password` are missing/empty in `config.yml`, versus loading
   successfully with zero values and letting the first API call fail (surfacing as
   `fortiweb_up 0`). This affects both the config test in Task 2 and startup
   behavior in Task 6 (main.go exiting fast on a broken config vs. starting an
   always-down exporter). (Assumption if no answer: `Load` returns an error if
   `target`, `username`, or `password` is empty, since a config missing these can
   never succeed and failing fast at startup is more operable.)

   **Answer:** Confirmed — `Load` returns an error when `target`, `username`, or
   `password` is empty. Fail fast at startup rather than running an exporter that
   can never succeed.

4. **Should the target URL be normalized to avoid double slashes?**
   Task 3 builds the request URL as `{baseURL}/api/v2.0/system/status.systemresource`.
   If a user's `config.yml` sets `target: https://fw.example.com/` (trailing slash),
   naive concatenation produces `.../fw.example.com//api/v2.0/...`. Should the
   client/config trim a trailing slash from the configured target before use?
   (Assumption if no answer: yes — trim any trailing `/` from the configured target
   URL before building the request path.)

   **Answer:** Confirmed — trim any trailing `/` from the configured target URL
   (do it once in `fortiweb.NewClient`, so it's normalized regardless of caller).

5. **Who is the copyright holder for the MIT `LICENSE`?**
   Task 9 says to attribute the copyright line to "the repo owner" but doesn't name
   them. The module path uses GitHub handle `ndkprd`. Should the `LICENSE` copyright
   line use that handle, a real name, or something else? (Assumption if no answer:
   use `ndkprd` as the copyright holder, matching the `go.mod` module path owner.)

   **Answer:** Use `ndkprd` as the copyright holder name in `LICENSE`.

6. **Task 4 references "the sample payload from this plan's Overview" — but the Overview has no concrete example values.**
   The Overview's flowchart/sequence diagrams list field *names* only (`cpu, mem,
   diskUsage, sessionCount, connCntPerSec, logDisk, dbStatus`), not example values.
   Task 4's verification implies a specific fixture to match against. Since none
   exists, I'll author a representative fixture (e.g. `cpu: 12.5, mem: 45.2,
   diskUsage: 10.1, sessionCount: 100, connCntPerSec: 5, logDisk: "Available",
   dbStatus: "Available"`) for the client and collector tests unless told otherwise.

   **Answer:** Use the exact sample the user provided as the canonical fixture
   across Task 3/4/5 tests, rather than inventing new values:
   ```json
   {
     "results": {
       "cpu": 17,
       "mem": 78,
       "logDisk": "Available",
       "dbStatus": "Available",
       "diskUsage": 79,
       "sessionCount": 21051,
       "connCntPerSec": 482
     }
   }
   ```
   These integers unmarshal cleanly into the `float64`/`int64` struct fields
   specified in Task 3 — no shape change needed.
