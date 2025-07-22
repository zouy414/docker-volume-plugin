# Quick Start Docker Volume Plugin

## Install

```sh
$ make plugin # or `docker plugin install --alias docker-volume-plugin zouy/docker-volume-plugin --grant-all-permissions --disable`
$ docker plugin set docker-volume-plugin DRIVER_OPTIONS='{"address":"nfs-server.example.com","remotePath":"/exported/path"}'
$ docker plugin enable docker-volume-plugin
```

## Usage

### Command

```sh
$ docker volume create \
  --driver docker-volume-plugin \
  sample
```

### Compose

```yaml
volumes:
  sample:
    driver: docker-volume-plugin
```

## How to Upgrade

1. Drain target node by `docker node update <target-node> --availability drain`
2. Wait for the node to drain and then execute `docker plugin disable -f docker-volume-plugin`
3. Upgrade plugin by `docker plugin upgrade`
