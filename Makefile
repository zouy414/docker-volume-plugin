PLUGIN ?= docker-volume-plugin#@variables The plugin of running
IMAGE ?= $(PLUGIN)#@variables The image name of builded image
TAG ?= latest#@variables The tag of builded image

HELP_PREFIX = @help
VARIABLES_PREFIX = @variables

.PHONY: help
help: #@help Display this help
	@echo "Usage: make \033[36m<target>\033[0m"
	@echo "Targets:"
	@awk -F ':.*#$(HELP_PREFIX)' '/^[a-zA-Z0-9_-]+:.*#$(HELP_PREFIX)/ {printf "  \033[36m%-13s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo "Variables:"
	@awk -F '?=.*#$(VARIABLES_PREFIX)' '/[a-zA-Z0-9_-]+ *\?=.*#$(VARIABLES_PREFIX)/ {var=$$1; gsub(/export +/, "", var); printf "  \033[36m%-23s\033[0m %s\n", var, $$2}' $(MAKEFILE_LIST)

run: #@help Run specified plugin on local
	go run cmd/$(PLUGIN)/main.go

build: #@help Build binary
	CGO_ENABLED=0 GO111MODULE=on go build -a -o bin/$(PLUGIN) cmd/$(PLUGIN)/main.go

image: #@help Build image
	sudo docker rmi -f $(IMAGE):$(TAG)
	sudo docker build \
		-t $(IMAGE):$(TAG) \
		--build-arg http_proxy=$(http_proxy) \
		--build-arg https_proxy=$(https_proxy) \
		-f cmd/$(PLUGIN)/Dockerfile .

plugin: image #@help Create docker plugin
	docker plugin disable -f $(IMAGE):$(TAG) || echo "plugin already disabled"
	docker plugin rm -f $(IMAGE):$(TAG) || echo "plugin already removed"
	rm -rf bin/plugin
	mkdir -p bin/plugin/rootfs
	docker export $(shell sudo docker create $(IMAGE):$(TAG) --name plugin_$(PLUGIN)) | tar -x -C bin/plugin/rootfs
	cp cmd/docker-volume-plugin/config.json bin/plugin
	sudo docker rm -vf plugin_$(PLUGIN)
	sudo docker plugin create $(IMAGE):$(TAG) bin/plugin

unit: #@help Run unit test
	go test -v ./... -coverprofile=cover.out
	go tool cover -html=cover.out -o coverage.html

clean: #@help Clean unnecessary files
	rm -f ./cover.out
	rm -f ./coverage.html
	rm -rf ./bin/*
