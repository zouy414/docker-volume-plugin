FROM registry.hub.docker.com/library/golang:1.23 AS builder
WORKDIR /workspace
COPY . /workspace
ENV EXECUTABLE=docker-volume-plugin
RUN make build

FROM alpine:3.22.0
WORKDIR /
COPY --from=builder /workspace/bin/docker-volume-plugin .
USER root
RUN apk add nfs-utils --no-cache
ENTRYPOINT ["/docker-volume-plugin"]
