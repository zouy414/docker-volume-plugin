FROM registry.hub.docker.com/library/golang:1.25 AS builder
WORKDIR /workspace
COPY . /workspace
RUN PLUGIN=docker-volume-plugin make build

FROM alpine:3.23.3
WORKDIR /
COPY --from=builder /workspace/bin/docker-volume-plugin .
USER root
RUN apk add nfs-utils cifs-utils --no-cache
ENTRYPOINT ["/docker-volume-plugin"]
