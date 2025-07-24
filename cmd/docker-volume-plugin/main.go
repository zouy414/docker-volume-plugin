package main

import (
	"context"
	"docker-volume-plugin/pkg/adapters"
	"docker-volume-plugin/pkg/log"
	"flag"
	"os"
	"strings"

	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-plugins-helpers/volume"
)

func main() {
	var logger = log.New("main")

	var logLevel string
	var unixEndpoint string
	var driver string
	var driverOptions string
	flag.StringVar(&logLevel, "log-level", os.Getenv("LOG_LEVEL"), "set the log level (debug, info, warn, error)")
	flag.StringVar(&unixEndpoint, "unit-endpoint", os.Getenv("UNIX_ENDPOINT"), "specify a UNIX endpoint to listen on")
	flag.StringVar(&driver, "driver", os.Getenv("DRIVER"), "specify a driver to use")
	flag.StringVar(&driverOptions, "driver-options", os.Getenv("DRIVER_OPTIONS"), "specify a json string of driver options")
	flag.Parse()

	switch strings.ToLower(logLevel) {
	case "debug":
		logger = logger.WithLogLevel(log.DebugLevel)
	case "info":
		logger = logger.WithLogLevel(log.InfoLevel)
	case "warn":
		logger = logger.WithLogLevel(log.WarnLevel)
	case "error":
		logger = logger.WithLogLevel(log.ErrorLevel)
	default:
		logger.Fatalf("invalid log level: %s", logLevel)
	}

	driverAdapter, err := adapters.NewVolumePlugin(context.Background(), logger.WithService("docker-volume-plugin"), driver, driverOptions)
	if err != nil {
		logger.Fatalf("failed to create docker volume plugin adapter: %v", err)
	}
	defer func() {
		if err := driverAdapter.Destroy(); err != nil {
			logger.Errorf("failed to destroy driver adapter: %v", err)
		}
	}()

	listener, err := sockets.NewUnixSocket(unixEndpoint, 0)
	if err != nil {
		logger.Fatalf("failed to create unix socket: %v", err)
	}
	defer listener.Close()

	handler := volume.NewHandler(driverAdapter)
	err = handler.Serve(listener)
	if err != nil {
		logger.Fatalf("failed to serve volume handler: %v", err)
	}
}
