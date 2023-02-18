#!/bin/bash

(/usr/bin/wait-for localhost:${PUBSUB_PORT} --timeout=0 PUBSUB_EMULATOR_HOST=localhost:${PUBSUB_PORT} /usr/bin/pubsub-emulator -debug; nc -lkp 8682 >/dev/null) &


gcloud beta emulators pubsub start --project=${PUBSUB_PROJECT} --host-port=0.0.0.0:${PUBSUB_PORT} "$@"