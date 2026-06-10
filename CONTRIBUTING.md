# Contributing

Thanks for your interest in improving the Pub/Sub Emulator!

## Getting Started

```bash
git clone https://github.com/dipjyotimetia/pubsub-emulator.git
cd pubsub-emulator
go mod download
```

Requires Go (see the version in [`go.mod`](go.mod)).

## Development Workflow

Common tasks are wrapped in the [`Makefile`](Makefile):

```bash
make build    # build the binary
make test     # run tests with the race detector
make lint     # run golangci-lint
make verify   # fmt + vet + lint + test (run this before pushing)
```

Run the dashboard locally:

```bash
PUBSUB_PROJECT=test-project \
PUBSUB_TOPIC=orders,payments \
PUBSUB_SUBSCRIPTION=orders-sub,payments-sub \
DASHBOARD_PORT=8080 \
go run .
```

## Pull Requests

1. Branch off `main`.
2. Keep changes focused; add or update tests for behavior changes.
3. Ensure `make verify` passes (CI runs tests, lint, build, and a vulnerability scan).
4. Use clear, conventional commit messages where possible.

## Reporting Issues

- Bugs and feature requests: open a [GitHub issue](https://github.com/dipjyotimetia/pubsub-emulator/issues).
- Security vulnerabilities: see [SECURITY.md](SECURITY.md) — please do not file public issues.
