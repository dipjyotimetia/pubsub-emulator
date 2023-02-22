#!/usr/bin/env bash

set -e

docker build . -t verify-emulator

docker run --rm --env PUBSUB_PROJECT=test-project \
  --env PUBSUB_TOPIC=test-topic \
  --env PUBSUB_SUBSCRIPTION=test-sub \
  --env PUBSUB_PORT=8085 \
  --detach --name verify verify-emulator

docker logs -f --until=10s verify > verifylogs

cat verifylogs

docker stop verify > /dev/null

echo verifying log output

grep "topic created: projects/test-project/topics/test-topic" verifylogs
grep "created subscription: projects/test-project/subscriptions/test-sub" verifylogs

echo logs contain expected output