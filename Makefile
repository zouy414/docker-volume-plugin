# Read user specially environment file
ifneq (,$(wildcard .env))
	include .env
	export
endif

export PLUGIN ?= docker-volume-plugin#@variables The target plugin to run or build, supported: docker-volume-plugin
export IMAGE ?= $(PLUGIN)#@variables The image name of builded image, default to plugin name
export TAG ?= latest#@variables The tag of builded image

HELP_PREFIX = @help
VARIABLES_PREFIX = @variables

.PHONY: help
help: #@help Display this help
	@echo "Usage: make \033[36m<target>\033[0m"
	@echo "Targets:"
	@awk -F ':.*#$(HELP_PREFIX)' '/^.+:.*#$(HELP_PREFIX)/ {printf "    \033[36m%-23s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) 2>/dev/null
	@echo "Variables:"
	@awk -F '\?=.*#$(VARIABLES_PREFIX)' '/export .+\?=.*#$(VARIABLES_PREFIX)/ {var=$$1; gsub(/export +/, "", var); printf "    \033[36m%-23s\033[0m %s\n", var, $$2}' $(MAKEFILE_LIST) 2>/dev/null

.PHONY: run
run: #@help Run specified plugin on local
	go run cmd/$(PLUGIN)/main.go

.PHONY: build
build: bin/$(PLUGIN) #@help Build binary

bin/%:
	CGO_ENABLED=0 GO111MODULE=on go build -a -o bin/$* cmd/$*/main.go

.PHONY: image
image: #@help Build image
	sudo docker rmi -f $(IMAGE):$(TAG)
	sudo docker build \
		-t $(IMAGE):$(TAG) \
		--build-arg http_proxy=$(http_proxy) \
		--build-arg https_proxy=$(https_proxy) \
		-f cmd/$(PLUGIN)/Dockerfile .

.PHONY: plugin
plugin: #@help Create docker plugin from image
	sudo docker plugin disable -f $(IMAGE):$(TAG) || echo "plugin $(IMAGE):$(TAG) already disabled"
	sudo docker plugin rm -f $(IMAGE):$(TAG) || echo "plugin $(IMAGE):$(TAG) already removed"
	rm -rf bin/plugin
	mkdir -p bin/plugin/rootfs
	sudo docker export $(shell sudo docker create $(IMAGE):$(TAG) --name plugin_$(PLUGIN)) | tar -x -C bin/plugin/rootfs
	cp cmd/docker-volume-plugin/config.json bin/plugin
	sudo docker rm -vf plugin_$(PLUGIN)
	sudo docker plugin create $(IMAGE):$(TAG) bin/plugin

.PHONY: unit
unit: #@help Run unit test
	go test -v ./... -coverprofile=cover.out
	go tool cover -html=cover.out -o coverage.html

.PHONY: clean
clean: #@help Clean unnecessary files
	rm -f ./cover.out
	rm -f ./coverage.html
	rm -rf ./bin/*
