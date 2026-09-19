## [unreleased]

### 📚 Documentation

- Add CONTRIBUTING.md and AGENTS.md

### ⚙️ Miscellaneous Tasks

- Fix cliff.toml to add a blank line between release headings

## [0.2.0] - 2026-09-19

### 🚀 Features

- Add FortiGuard license expiry/validity metrics

## [0.1.0] - 2026-09-18

### 🚀 Features

- Add config loading with validation and defaults
- Add FortiWeb API client with Basic Auth
- Add Prometheus collector for FortiWeb system status
- Wire up main entrypoint
- Support multiple FortiWeb targets via ?target= param
- Serve a landing page at / listing configured targets
- Split /metrics into exporter self-metrics, add /probe
- Add Grafana dashboard and container to examples
- Add per-vdom server-policy/content-routing/server-pool count metrics
- Add 4 more per-vdom resource-count metrics

### 🐛 Bug Fixes

- Use FortiWeb's actual Authorization header scheme

### 📚 Documentation

- Add initial plan for FortiWeb Prometheus exporter
- Add MIT LICENSE
- Add README
- Mark all plan tasks completed
- Add Fortinet trademark disclaimer to README
- Add examples/ with a runnable exporter + Prometheus stack
- Add plan for per-vdom resource-count metrics
- Amend plan 002 with 4 more resource-count metrics
- Add resource-count panels to example Grafana dashboard

### 🚜 Refactor

- Rename Go module to gitlab.com/endekasoft/fortiweb_exporter
- Rename project references from fortiweb_exporter to fortiweb-exporter

### 🧪 Testing

- Add unit tests for FortiWeb API client

### ⚙️ Miscellaneous Tasks

- Scaffold Go module and project layout
- Add GitLab CI pipeline for building/publishing the image
- Keep .agents/ tracked in git, ignore .codegraph/ index only
- Untrack .agents/ going forward

### 💼 Other

- Add multi-stage Dockerfile
- Add docker-compose.yml
