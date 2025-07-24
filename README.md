# Docker Volume Plugin

[![CI](https://github.com/zouy414/docker-volume-plugin/actions/workflows/ci.yml/badge.svg)](https://github.com/zouy414/docker-volume-plugin/actions/workflows/ci.yml)

NFS volume plugin for docker

## Quick Start

### Install

```sh
$ make plugin # or `docker plugin install --alias docker-volume-plugin zouyu613/docker-volume-plugin:<tag> --grant-all-permissions --disable`
$ docker plugin set docker-volume-plugin DRIVER=nfs DRIVER_OPTIONS='{"address":"nfs-server.example.com","remotePath":"/exported/path"}'
$ docker plugin enable docker-volume-plugin
```

### Usage

#### Command

```sh
$ docker volume create --driver docker-volume-plugin sample
```

#### Compose

```yaml
volumes:
  sample:
    driver: docker-volume-plugin
```

### How to Upgrade

1. Drain target node by `docker node update <target-node> --availability drain`
2. Wait for the node to drain and then execute `docker plugin disable -f docker-volume-plugin`
3. Upgrade plugin by `docker plugin upgrade`

## Supported Net Volumes

|Name|Driver|Options|
|:-|:-|:-|
|NFS|nfs|[NFS-Driver-Options.md](docs/NFS-Driver-Options.md)|
