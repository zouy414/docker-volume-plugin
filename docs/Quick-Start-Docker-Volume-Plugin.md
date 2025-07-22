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
