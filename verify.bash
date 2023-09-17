#!/usr/bin/env bash

set -e

docker build . -t verify-emulator

docker run --rm --env PUBSUB_PROJECT=test-project \
  --env PUBSUB_TOPIC=test-topic1,test-topic2,test-topic3 \
  --env PUBSUB_SUBSCRIPTION=test-sub1,test-sub2,test-sub3 \
  --env PUBSUB_PORT=8085 \
  --detach \
  --name verify \
  verify-emulator

sleep 10

docker logs verify &> verifylogs

cat verifylogs

docker stop verify > /dev/null

echo verifying log output

grep "Server started, listening on 8085" verifylogs
grep "Topic created: projects/test-project/topics/test-topic1" verifylogs
grep "Created subscription: projects/test-project/subscriptions/test-sub1" verifylogs
grep "Published message with ID: 1" verifylogs
grep "Received message: Hello, Pub/Sub emulator!" verifylogs

echo logs contain expected output