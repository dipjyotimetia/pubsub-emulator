# Google Cloud PubSub Emulator

[Google Cloud PubSub Emulator](https://cloud.google.com/pubsub/docs/emulator)

## Usage

Capability to create multiple topic and subscription

Set the `PUBSUB_PROJECT`, `PUBSUB_TOPIC`, `PUBSUB_SUBSCRIPTION` and `PUBSUB_PORT` environment variables when running the image.

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

---
Thanks to [RoryQ/spanner-emulator](https://github.com/RoryQ/spanner-emulator) for the motivation.
