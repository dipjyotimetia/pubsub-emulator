#!/bin/bash

# Check if PUBSUB_PROJECT environment variable is set
if [ -z "${PUBSUB_PROJECT}" ]; then
  echo "Missing PUBSUB_PROJECT environment variable" >&2
  exit 1
fi

# Check if PUBSUB_PORT environment variable is set
if [ -z "${PUBSUB_PORT}" ]; then
  echo "Missing PUBSUB_PORT environment variable" >&2
  exit 1
fi

# Start the Pub/Sub emulator and wait for it to be ready
(/usr/bin/wait-for localhost:${PUBSUB_PORT} -- env PUBSUB_EMULATOR_HOST=localhost:${PUBSUB_PORT} /usr/bin/pubsub-emulator; nc -lkp 8682 >/dev/null) &

# Start the Pub/Sub emulator using gcloud command
gcloud beta emulators pubsub start --host-port=0.0.0.0:${PUBSUB_PORT} --project=${PUBSUB_PROJECT} "$@"