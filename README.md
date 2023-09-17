# Google Cloud PubSub Emulator

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

### 2. Run the Container

Use the provided Docker image to run the emulator:

```yaml
version: '3.8'
services:
  pubsub-emulator:
    image: dipjyotimetia/pubsub-emulator:latest
    environment:
      - PUBSUB_PROJECT=test-project
      - PUBSUB_TOPIC=test-topic1,test-topic2,test-topic3
      - PUBSUB_SUBSCRIPTION=test-sub1,test-sub2,test-sub3
      - PUBSUB_PORT=8085
```

Replace the values in the environment variables with your own project, topic, and subscription names.

### 3. Develop and Test

With the emulator running, you can now develop and test your Google Cloud Pub/Sub applications locally. Any messages published to the emulator will be processed as if they were sent to the actual Google Cloud Pub/Sub service.

---

**Note**: The emulator is an excellent tool for local development and testing, but remember that it may not fully replicate all features of the production environment. Be sure to thoroughly test your code in the actual Google Cloud environment before deploying it to production.

---

**Special Thanks**: This project was inspired by [RoryQ/spanner-emulator](https://github.com/RoryQ/spanner-emulator). Thanks for the motivation!
