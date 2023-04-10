FROM golang:1.19-buster as builder

LABEL MAINTAINER="dipjyotimetia"

ENV PUBSUB_PROJECT ${PUBSUB_PROJECT}
ENV PUBSUB_TOPIC ${PUBSUB_TOPIC}
ENV PUBSUB_SUBSCRIPTION ${PUBSUB_SUBSCRIPTION}
ENV PUBSUB_EMULATOR_HOST ${PUBSUB_PORT}

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
RUN go build .

FROM google/cloud-sdk:425.0.0-debian_component_based

COPY --from=builder /usr/bin/wait-for /usr/bin
COPY --from=builder /build/pubsub-emulator /usr/bin
COPY run.sh /run.sh

RUN gcloud components install beta pubsub-emulator

RUN apt-get update && apt-get install -y --no-install-recommends \
    netcat-openbsd \
    bash \
    openjdk-17-jre-headless \
    && rm -rf /var/lib/apt/lists/*

EXPOSE ${PUBSUB_PORT}

RUN chmod +x /run.sh

ENTRYPOINT [ "sh", "/run.sh"]