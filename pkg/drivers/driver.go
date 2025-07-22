package drivers

import (
	"context"
	"fmt"
	"net-volume-plugins/pkg/drivers/apis"
	"net-volume-plugins/pkg/log"
)

type driverFactory func(ctx context.Context, logger *log.Logger, propagatedMountpoint string, driverOptions string) (apis.Driver, error)

var driverFactories map[string]driverFactory = map[string]driverFactory{}

// registerFactory to register factory
func registerFactory(name string, factory driverFactory) {
	driverFactories[name] = factory
}

// New creates a new driver instance
func New(ctx context.Context, logger *log.Logger, name string, propagatedMountpoint string, driverOptions string) (apis.Driver, error) {
	factory := driverFactories[name]
	if factory == nil {
		return nil, fmt.Errorf("driver %s is invalid", name)
	}

	return factory(ctx, logger.WithService(name), propagatedMountpoint, driverOptions)
}
