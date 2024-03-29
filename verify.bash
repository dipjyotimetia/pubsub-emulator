#!/usr/bin/env bash

# Exit on any error
set -e

# Build the Docker container and tag it as "verify-emulator"
docker build . -t verify-emulator

# Run the Docker container with the required environment variables and detach it
docker run --rm --env PUBSUB_PROJECT=test-project \
  --env PUBSUB_TOPIC=test-topic1,test-topic2,test-topic3 \
  --env PUBSUB_SUBSCRIPTION=test-sub1,test-sub2,test-sub3 \
  --env PUBSUB_PORT=8085 \
  --detach \
  --name verify \
  verify-emulator

# Sleep for 10 seconds to allow the container to start
sleep 10

# Capture the container logs to a file
docker logs verify &> verifylogs

# Display the container logs
cat verifylogs

# Stop the Docker container
docker stop verify > /dev/null

# Verify log output
echo "Verifying log output..."
if grep -q "Server started, listening on 8085" verifylogs &&
   grep -q "Topic created: projects/test-project/topics/test-topic1" verifylogs &&
   grep -q "Created subscription: projects/test-project/subscriptions/test-sub1" verifylogs &&
   grep -q "Published message with ID: 1" verifylogs &&
   grep -q "Received message: Hello, Pub/Sub emulator!" verifylogs; then
  echo "Logs contain expected output."
else
  echo "Logs do not contain expected output."
fi
