# Google Cloud PubSub Emulator

[![Publish Docker image](https://github.com/dipjyotimetia/pubsub-emulator/actions/workflows/docker-publish.yaml/badge.svg)](https://github.com/dipjyotimetia/pubsub-emulator/actions/workflows/docker-publish.yaml)
[![codecov](https://codecov.io/github/dipjyotimetia/pubsub-emulator/graph/badge.svg?token=PO4ID8VSP5)](https://codecov.io/github/dipjyotimetia/pubsub-emulator)  
[Google Cloud PubSub Emulator](https://cloud.google.com/pubsub/docs/emulator) is a tool that allows you to run a local emulator of Google Cloud Pub/Sub, making it easy to develop and test Pub/Sub applications without incurring costs or interacting with the production environment.

## Usage

This emulator provides the capability to create multiple topics and subscriptions for testing purposes. To get started, follow these steps:

### Prerequisites

- Docker installed on your system.

### 1. Set Environment Variables

Before running the emulator, set the required environment variables in your shell:

- `PUBSUB_PROJECT`: Your Google Cloud project ID.
- `PUBSUB_TOPIC`: Comma-separated list of Pub/Sub topic names.
- `PUBSUB_SUBSCRIPTION`: Comma-separated list of Pub/Sub subscription names.
- `PUBSUB_PORT` (optional): Port to run the emulator on (default is `8085`).
- `DASHBOARD_PORT` (optional): Port for the web dashboard. **Leave unset to disable dashboard.**

### 2. Run the Container

Use the provided Docker image to run the emulator:

```yaml
services:
  pubsub-emulator:
    image: dipjyotimetia/pubsub-emulator:latest
    environment:
      - PUBSUB_PROJECT=test-project
      - PUBSUB_TOPIC=test-topic1,test-topic2,test-topic3
      - PUBSUB_SUBSCRIPTION=test-sub1,test-sub2,test-sub3
      - PUBSUB_PORT=8085
      - DASHBOARD_PORT=8080
    ports:
      - "8085:8085"
      - "8080:8080"
```

Replace the values in the environment variables with your own project, topic, and subscription names.

## Web Dashboard (Optional)

The emulator includes an optional built-in web dashboard for monitoring, debugging, and managing your Pub/Sub resources.

### Enabling the Dashboard

Set the `DASHBOARD_PORT` environment variable:

```yaml
services:
  pubsub-emulator:
    image: dipjyotimetia/pubsub-emulator:latest
    environment:
      - PUBSUB_PROJECT=test-project
      - PUBSUB_TOPIC=test-topic1,test-topic2
      - PUBSUB_SUBSCRIPTION=test-sub1,test-sub2
      - DASHBOARD_PORT=8080  # Dashboard enabled on port 8080
    ports:
      - "8085:8085"
      - "8080:8080"
```

Access at: `http://localhost:8080`

### Running Without Dashboard (Backward Compatible)

Simply omit the `DASHBOARD_PORT` variable:

```yaml
services:
  pubsub-emulator:
    image: dipjyotimetia/pubsub-emulator:latest
    environment:
      - PUBSUB_PROJECT=test-project
      - PUBSUB_TOPIC=test-topic1,test-topic2
      - PUBSUB_SUBSCRIPTION=test-sub1,test-sub2
    ports:
      - "8085:8085"  # Only emulator, no dashboard
```

### Dashboard Features

- **📊 Live Statistics**: Real-time overview of topics, subscriptions, and messages
- **🔍 Search & Filter**: Search messages by content or filter by topic
- **✉️ Publish Messages**: Send messages directly from the UI
- **➕ Create Resources**: Create topics and subscriptions via the dashboard
- **📥 Export Data**: Export messages to JSON or CSV
- **🔄 Message Replay**: Replay any historical message
- **⚡ Live Updates**: WebSocket connection for instant message notifications
- **📬 Message Viewer**: View up to 1000 recent messages with full details

### 3. Develop and Test

With the emulator running, you can now develop and test your Google Cloud Pub/Sub applications locally. Any messages published to the emulator will be processed as if they were sent to the actual Google Cloud Pub/Sub service.

---

**Note**: The emulator is an excellent tool for local development and testing, but remember that it may not fully replicate all features of the production environment. Be sure to thoroughly test your code in the actual Google Cloud environment before deploying it to production.

---

**Special Thanks**: This project was inspired by [RoryQ/spanner-emulator](https://github.com/RoryQ/spanner-emulator). Thanks for the motivation!
