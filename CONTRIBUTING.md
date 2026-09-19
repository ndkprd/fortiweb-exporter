# Contributing

## Setup

Requires Go 1.27+.

```sh
git clone git@gitlab.com:endekasoft/fortiweb-exporter.git
cd fortiweb-exporter
go build ./...
```

## Development

```sh
go build ./...      # build
go vet ./...         # static checks
gofmt -l .            # list unformatted files (should be empty)
go test ./...        # run tests
```

Run `gofmt -w <file>` on anything `gofmt -l` flags before committing.

A local run against a real appliance needs a `config.yml` — copy
`config.yml.example` and fill in your FortiWeb target(s). See `examples/`
for a runnable exporter + Prometheus + Grafana stack via `docker-compose`.

## Making changes

- Keep changes scoped to what the issue/task asks for — no drive-by
  refactors in the same commit.
- Add or update tests for any behavior change (see `*_test.go` files
  alongside the code they cover).
- Update `README.md`'s metrics table when adding or changing a metric, and
  the example Grafana dashboard (`examples/grafana/dashboards/`) when adding
  a metric worth visualizing.

## Commit messages

This repo follows [Conventional Commits](https://www.conventionalcommits.org/)
(`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`, ...) — `CHANGELOG.md`
is generated from commit history via [git-cliff](https://git-cliff.org/), so
non-conventional commits get silently dropped from it.

## Reporting bugs / requesting features

Open an issue on the
[GitLab issue tracker](https://gitlab.com/endekasoft/fortiweb-exporter/-/issues)
before writing any code — this includes bug reports and feature requests.

## Submitting changes

1. Start from an issue — every change should trace back to one. If no issue
   covers your change yet, open one first (see above).
2. Wait until the issue is assigned to you before starting work, to avoid
   duplicate effort on the same issue.
3. Branch off `main`.
4. Make sure `go build ./...`, `go vet ./...`, and `go test ./...` pass.
5. Open a merge request against `main`, linking the issue it resolves.

## License

By contributing, you agree your contributions are licensed under this
project's [MIT License](LICENSE).
