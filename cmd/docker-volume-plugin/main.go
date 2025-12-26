package main

import (
	"context"
	"docker-volume-plugin/pkg/adapters"
	"docker-volume-plugin/pkg/log"
	"flag"
	"os"

	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-plugins-helpers/volume"
)

func main() {
	var logLevel string
	var unixEndpoint string
	var driver string
	var driverOptions string

	// Parse flags
	flag.StringVar(&logLevel, "log-level", os.Getenv("LOG_LEVEL"), "set the log level (debug, info, warn, error)")
	flag.StringVar(&unixEndpoint, "unit-endpoint", os.Getenv("UNIX_ENDPOINT"), "specify a UNIX endpoint to listen on")
	flag.StringVar(&driver, "driver", os.Getenv("DRIVER"), "specify a driver to use")
	flag.StringVar(&driverOptions, "driver-options", os.Getenv("DRIVER_OPTIONS"), "specify a json string of driver options")
	flag.Parse()

	// Create logger
	var logger = log.NewWithLogLevel("main", log.StringToLogLevel(logLevel))

	// Create driver adapter
	driverAdapter, err := adapters.NewVolumePlugin(context.Background(), logger.WithService("docker-volume-plugin"), driver, driverOptions)
	if err != nil {
		logger.Fatalf("failed to create docker volume plugin adapter: %v", err)
	}
	defer func() {
		if err := driverAdapter.Destroy(); err != nil {
			logger.Errorf("failed to destroy driver adapter: %v", err)
		}
	}()

	// Create unix socket listener
	listener, err := sockets.NewUnixSocket(unixEndpoint, 0)
	if err != nil {
		logger.Fatalf("failed to create unix socket: %v", err)
	}
	defer listener.Close()

	// Bind driver adapter to volume handler
	handler := volume.NewHandler(driverAdapter)
	err = handler.Serve(listener)
	if err != nil {
		logger.Fatalf("failed to serve volume handler: %v", err)
	}
}
