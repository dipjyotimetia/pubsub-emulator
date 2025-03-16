#!/bin/bash
set -e

# Validate required environment variables
if [ -z "${PUBSUB_PROJECT}" ]; then
  echo "ERROR: Missing PUBSUB_PROJECT environment variable" >&2
  exit 1
fi

# Set default port if not provided
if [ -z "${PUBSUB_PORT}" ]; then
  PUBSUB_PORT="8085"
  echo "INFO: Using default PUBSUB_PORT=${PUBSUB_PORT}"
fi

# Export the emulator host for the client application
export PUBSUB_EMULATOR_HOST="localhost:${PUBSUB_PORT}"

# Start the pubsub-emulator client in background with proper error handling
(
  # Wait for emulator to be ready before starting client
  echo "INFO: Waiting for emulator to be ready on port ${PUBSUB_PORT}..."
  /usr/bin/wait-for localhost:${PUBSUB_PORT} -t 30 -- \
    echo "INFO: Emulator detected, starting client application" && \
    /usr/bin/pubsub-emulator
  
  # Keep container running even if application exits
  echo "INFO: Starting network listener to keep container alive"
  nc -lkp 8682 >/dev/null
) &

# Start the actual emulator in foreground
echo "INFO: Starting PubSub emulator on port ${PUBSUB_PORT} for project ${PUBSUB_PROJECT}"
exec gcloud beta emulators pubsub start --host-port=0.0.0.0:${PUBSUB_PORT} --project=${PUBSUB_PROJECT} "$@"