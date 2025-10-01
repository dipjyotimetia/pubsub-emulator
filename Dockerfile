ARG GO_VERSION=1.24.0
ARG GCLOUD_SDK_VERSION=541.0.0
ARG PUBSUB_PROJECT="demo-project"
ARG PUBSUB_TOPIC="demo-topic"
ARG PUBSUB_SUBSCRIPTION="demo-sub"
ARG PUBSUB_PORT=8681
FROM golang:${GO_VERSION}-bullseye AS builder

LABEL maintainer="dipjyotimetia"
LABEL description="This is a custom image for GCP Pubsub Emulator"
LABEL repository="https://github.com/dipjyotimetia/pubsub-emulator"

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    && rm -rf /var/lib/apt/lists/*

RUN curl -s https://raw.githubusercontent.com/eficode/wait-for/master/wait-for -o /usr/bin/wait-for
RUN chmod +x /usr/bin/wait-for

WORKDIR /build

ENV GO111MODULE=on

COPY go.mod go.sum main.go ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build .

FROM google/cloud-sdk:${GCLOUD_SDK_VERSION}-emulators

ARG PUBSUB_PROJECT
ARG PUBSUB_TOPIC
ARG PUBSUB_SUBSCRIPTION
ARG PUBSUB_PORT

# Expose runtime env vars
ENV PUBSUB_PROJECT=${PUBSUB_PROJECT} \
    PUBSUB_TOPIC=${PUBSUB_TOPIC} \
    PUBSUB_SUBSCRIPTION=${PUBSUB_SUBSCRIPTION} \
    PUBSUB_EMULATOR_HOST=0.0.0.0:${PUBSUB_PORT}

COPY --from=builder /usr/bin/wait-for /usr/bin
COPY --from=builder /build/pubsub-emulator /usr/bin
COPY run.sh /run.sh

RUN apt-get update && apt-get install -y --no-install-recommends \
    netcat-openbsd \
    bash \
    openjdk-21-jre-headless \
    && rm -rf /var/lib/apt/lists/*

EXPOSE ${PUBSUB_PORT}

RUN chmod +x /run.sh

ENTRYPOINT [ "sh", "/run.sh"]