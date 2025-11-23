# Docker Volume Plugin

[![CI](https://github.com/zouy414/docker-volume-plugin/actions/workflows/ci.yml/badge.svg)](https://github.com/zouy414/docker-volume-plugin/actions/workflows/ci.yml)
[![Release](https://github.com/zouy414/docker-volume-plugin/actions/workflows/release.yml/badge.svg)](https://github.com/zouy414/docker-volume-plugin/actions/workflows/release.yml)

NFS volume plugin for docker

## Quick Start

### Install

```sh
$ make image plugin # or install released version by `docker plugin install --alias docker-volume-plugin zouyu613/docker-volume-plugin:<tag> --grant-all-permissions --disable`
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
    driver_opts:
      purgeAfterDelete: "true" # optional
```

### Upgrade

1. Drain target node by `docker node update <target-node> --availability drain`
2. Wait for the node to drain and then disable plugin by `docker plugin disable docker-volume-plugin -f`
3. Upgrade plugin by `docker plugin upgrade docker-volume-plugin zouyu613/docker-volume-plugin:<tag> --grant-all-permissions`
4. Enable plugin by `docker plugin enable docker-volume-plugin`
5. Active target node by `docker node update <target-node> --availability active`

## Supported Net Volumes

|Name|Driver|Options|
|:-|:-|:-|
|NFS|nfs|[NFS-Driver.md](docs/NFS-Driver.md)|
