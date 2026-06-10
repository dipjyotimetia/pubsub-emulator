#!/usr/bin/env bash

# Exit on any error
set -euo pipefail

# Build the Docker container and tag it as "verify-emulator"
docker build . -t verify-emulator

# Always stop the container and dump logs on exit
cleanup() {
  echo "--- container logs ---"
  docker logs verify 2>&1 || true
  docker stop verify > /dev/null 2>&1 || true
}
trap cleanup EXIT

# Run the Docker container with the required environment variables and detach it.
# DASHBOARD_PORT enables the web dashboard, which is the surface this script verifies.
docker run --rm --env PUBSUB_PROJECT=test-project \
  --env PUBSUB_TOPIC=test-topic1,test-topic2,test-topic3 \
  --env PUBSUB_SUBSCRIPTION=test-sub1,test-sub2,test-sub3 \
  --env PUBSUB_PORT=8085 \
  --env DASHBOARD_PORT=8080 \
  --publish 8080:8080 \
  --publish 8085:8085 \
  --detach \
  --name verify \
  verify-emulator

# Wait for the dashboard to come up. The dashboard server starts only after the
# emulator is ready and the configured topics/subscriptions have been created,
# so a successful health response proves the whole client -> emulator -> dashboard
# path is working.
echo "Waiting for dashboard health endpoint..."
for i in $(seq 1 30); do
  if curl -fsS -m 3 http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "Dashboard is healthy."
    break
  fi
  if [ "$i" -eq 30 ]; then
    echo "ERROR: dashboard did not become healthy in time." >&2
    exit 1
  fi
  sleep 2
done

# Verify the dashboard reports the topics created at startup. This confirms the
# Go client connected to the emulator and provisioned the configured resources.
echo "Verifying dashboard stats..."
stats=$(curl -fsS -m 5 http://localhost:8080/api/stats)
echo "$stats"
if ! echo "$stats" | grep -q "test-topic1"; then
  echo "ERROR: dashboard stats did not report the configured topics." >&2
  exit 1
fi

echo "Verification succeeded: emulator is running and the dashboard is serving on port 8080."
