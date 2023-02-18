FROM golang:1.19-alpine as builder

ENV PUBSUB_PROJECT ${PUBSUB_PROJECT}
ENV PUBSUB_TOPIC ${PUBSUB_TOPIC}
ENV PUBSUB_SUBSCRIPTION ${PUBSUB_SUBSCRIPTION}
ENV PUBSUB_PORT 8085
ENV PUBSUB_EMULATOR_HOST ${PUBSUB_PORT}

WORKDIR /build
ENV GO111MODULE=on
COPY go.mod go.sum main.go ./
RUN go build .

RUN apk update && apk upgrade && apk add --no-cache curl git

RUN curl -s https://raw.githubusercontent.com/eficode/wait-for/master/wait-for -o /usr/bin/wait-for
RUN chmod +x /usr/bin/wait-for


FROM google/cloud-sdk:alpine

COPY --from=builder /usr/bin/wait-for /usr/bin
COPY --from=builder /build/pubsub-emulator /usr/bin
COPY start.sh /start.sh
    
RUN apk add openjdk11-jre gcompat bash netcat-openbsd && gcloud components install beta pubsub-emulator

ENV LD_PRELOAD=/lib/libgcompat.so.0

EXPOSE ${PUBSUB_PORT}

ENTRYPOINT ["/bin/sh","/start.sh"]