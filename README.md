# Google Cloud Pub/Sub Emulator

<div align="center">

[![Publish Docker image](https://github.com/dipjyotimetia/pubsub-emulator/actions/workflows/docker-publish.yaml/badge.svg)](https://github.com/dipjyotimetia/pubsub-emulator/actions/workflows/docker-publish.yaml)
[![codecov](https://codecov.io/github/dipjyotimetia/pubsub-emulator/graph/badge.svg?token=PO4ID8VSP5)](https://codecov.io/github/dipjyotimetia/pubsub-emulator)
[![Go Report Card](https://goreportcard.com/badge/github.com/dipjyotimetia/pubsub-emulator)](https://goreportcard.com/report/github.com/dipjyotimetia/pubsub-emulator)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

A local emulator for [Google Cloud Pub/Sub](https://cloud.google.com/pubsub/docs/emulator) with a web dashboard. Run Pub/Sub locally for development and testing without cloud credentials or internet connection.

## Features

- Drop-in replacement for Google Cloud Pub/Sub
- Web dashboard with live message monitoring
- Manage multiple topics and subscriptions
- Docker images ready to use
- Works offline, no cloud credentials needed

## Quick Start

### What you'll need

- Docker or Docker Compose
- Familiarity with Pub/Sub (topics and subscriptions)

### Docker Compose (easiest way)

Create `docker-compose.yml`:

```yaml
services:
  pubsub-emulator:
    image: dipjyotimetia/pubsub-emulator:latest
    container_name: pubsub-emulator
    environment:
      # Required
      - PUBSUB_PROJECT=test-project
      - PUBSUB_TOPIC=orders,payments,notifications
      - PUBSUB_SUBSCRIPTION=orders-sub,payments-sub,notifications-sub

      # Optional
      - PUBSUB_PORT=8085          # Emulator port (default: 8085)
      - DASHBOARD_PORT=8080       # Dashboard port (omit to disable)
      - MESSAGE_TO_PUBLISH=       # Auto-publish test message (optional)
    ports:
      - "8085:8085"  # Pub/Sub emulator
      - "8080:8080"  # Web dashboard
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/api/health"]
      interval: 10s
      timeout: 5s
      retries: 3
```

Then start it:

```bash
docker-compose up -d
```

### Or use Docker directly

```bash
docker run -d \
  --name pubsub-emulator \
  -p 8085:8085 \
  -p 8080:8080 \
  -e PUBSUB_PROJECT=test-project \
  -e PUBSUB_TOPIC=topic1,topic2 \
  -e PUBSUB_SUBSCRIPTION=sub1,sub2 \
  -e DASHBOARD_PORT=8080 \
  dipjyotimetia/pubsub-emulator:latest
```

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PUBSUB_PROJECT` | Yes | - | Google Cloud project ID |
| `PUBSUB_TOPIC` | Yes | - | Comma-separated list of topic names |
| `PUBSUB_SUBSCRIPTION` | Yes | - | Comma-separated list of subscription names (must match topic count) |
| `PUBSUB_PORT` | No | `8085` | Port for Pub/Sub emulator gRPC endpoint |
| `DASHBOARD_PORT` | No | _disabled_ | Port for web dashboard (omit to disable) |
| `MESSAGE_TO_PUBLISH` | No | - | Optional message to auto-publish on startup |

### Topic-Subscription Pairing

Topics and subscriptions are paired by position:

```bash
PUBSUB_TOPIC=orders,payments,notifications
PUBSUB_SUBSCRIPTION=orders-sub,payments-sub,notifications-sub
# Pairs: orders↔orders-sub, payments↔payments-sub, notifications↔notifications-sub
```

## Web Dashboard

The emulator comes with a built-in web UI. Set `DASHBOARD_PORT=8080` to enable it, then open:

```
http://localhost:8080
```

### What you can do

- View live stats (topics, subscriptions, message counts)
- Browse recent messages (up to 1,000)
- Search and filter messages
- Publish test messages
- Create topics and subscriptions on the fly
- Replay messages for testing
- Real-time updates via WebSocket
- Dark mode toggle

## Using it in your code

Just point your Pub/Sub client to `localhost:8085`:

## Why use this?

- Develop locally without cloud credentials or costs
- Test message flows in your CI/CD pipeline
- Debug issues using the web dashboard
- Run demos without internet
- Learn Pub/Sub concepts offline

## Limitations

This is meant for local development and testing. Keep in mind:

- Some advanced GCP features aren't implemented
- Not optimized for production-level throughput
- Messages are stored in memory (no persistence)
- No authentication or IAM
- Single instance only

Always test against real GCP Pub/Sub before going to production.

## Contributing

Pull requests are welcome! Here's the process:

1. Fork the repo
2. Create a branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Run the linter (`golangci-lint run`)
6. Commit and push
7. Open a PR

Please write tests for new features and update docs as needed.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Credits

Inspired by [RoryQ/spanner-emulator](https://github.com/RoryQ/spanner-emulator). Built with the [Google Cloud Pub/Sub Go Client](https://pkg.go.dev/cloud.google.com/go/pubsub) and styled with [Pico CSS](https://picocss.com/).

## Need help?

- Found a bug? [Open an issue](https://github.com/dipjyotimetia/pubsub-emulator/issues)
- Have questions? [Start a discussion](https://github.com/dipjyotimetia/pubsub-emulator/discussions)
- Want to learn more? Check the [official Pub/Sub docs](https://cloud.google.com/pubsub/docs)

