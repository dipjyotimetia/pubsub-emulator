ARG GO_VERSION=1.24.0
ARG GCLOUD_SDK_VERSION=514.0.0
FROM golang:${GO_VERSION}-bullseye as builder

LABEL maintainer="dipjyotimetia"
LABEL version="3.0"
LABEL description="This is a custom image for GCP Pubsub Emulator"
LABEL repository="https://github.com/dipjyotimetia/pubsub-emulator"

ENV PUBSUB_EMULATOR_HOST ${PUBSUB_PORT}

RUN apk add --no-cache ca-certificates curl git

# Download wait-for script
RUN curl -s https://raw.githubusercontent.com/eficode/wait-for/master/wait-for -o /usr/bin/wait-for \
    && chmod +x /usr/bin/wait-for

WORKDIR /build

COPY go.mod go.sum main.go ./

# Download dependencies and build application
RUN go mod download \
    && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o pubsub-emulator .

# Use distroless image for the final stage
FROM google/cloud-sdk:${GCLOUD_SDK_VERSION}-emulators

# Copy only necessary files from builder
COPY --from=builder /usr/bin/wait-for /usr/bin/
COPY --from=builder /build/pubsub-emulator /usr/bin/
COPY run.sh /run.sh

# Install only required packages with minimal layers
RUN apt-get update && apt-get install -y --no-install-recommends \
    netcat-openbsd \
    bash \
    openjdk-17-jre-headless \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && chmod +x /run.sh

# Environment variables should be set at runtime for flexibility
ENV PUBSUB_PROJECT="" \
    PUBSUB_TOPIC="" \
    PUBSUB_SUBSCRIPTION="" \
    PUBSUB_EMULATOR_HOST=""

EXPOSE ${PUBSUB_PORT}

ENTRYPOINT [ "sh", "/run.sh"]
