package drivers

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/zouy414/docker-volume-plugin/pkg/drivers/apis"
	"github.com/zouy414/docker-volume-plugin/pkg/log"
	"github.com/zouy414/docker-volume-plugin/pkg/utils"
)

func init() {
	registerFactory("mock", mockFactory)
}

func mockFactory(ctx context.Context, logger *log.Logger, propagatedMountpoint string, driverOptions string) (apis.Driver, error) {
	if err := utils.MountMock(propagatedMountpoint); err != nil {
		return nil, fmt.Errorf("failed to create mock mount point: %s", err)
	}

	return &mock{
		logger:               logger,
		propagatedMountpoint: propagatedMountpoint,
		volumeMetadataMap:    make(map[string]*apis.VolumeMetadata),
	}, nil
}

// mock is a simple implementation of the Driver interface for testing purposes.
// It simulates volume management by maintaining an in-memory map of volume metadata
// and creating corresponding directories on the filesystem.
// This allows for testing the plugin's functionality without relying on an actual storage backend.
type mock struct {
	logger               *log.Logger
	propagatedMountpoint string
	volumeMetadataMap    map[string]*apis.VolumeMetadata
}

func (driver *mock) Create(name string, options map[string]string) error {
	driver.volumeMetadataMap[name] = &apis.VolumeMetadata{
		CreatedAt: time.Now(),
		Spec:      &apis.VolumeSpec{},
		Status: &apis.VolumeStatus{
			Mountpoint: path.Join(driver.propagatedMountpoint, name),
		},
	}
	return os.MkdirAll(path.Join(driver.propagatedMountpoint, driver.volumeMetadataMap[name].Status.Mountpoint), 0755)
}

func (driver *mock) List() (map[string]*apis.VolumeMetadata, error) {
	return driver.volumeMetadataMap, nil
}

func (driver *mock) Get(name string) (*apis.VolumeMetadata, error) {
	return driver.volumeMetadataMap[name], nil
}

func (driver *mock) Remove(name string) error {
	if driver.volumeMetadataMap[name] == nil {
		return fmt.Errorf("volume %s does not exist", name)
	}
	delete(driver.volumeMetadataMap, name)
	return nil
}

func (driver *mock) Path(name string) (string, error) {
	volumeMetadata, existed := driver.volumeMetadataMap[name]
	if !existed {
		return "", fmt.Errorf("volume %s does not exist", name)
	}
	return volumeMetadata.Status.Mountpoint, nil
}

func (driver *mock) Mount(name string, id string) (string, error) {
	volumeMetadata, existed := driver.volumeMetadataMap[name]
	if !existed {
		return "", fmt.Errorf("volume %s does not exist", name)
	}
	return volumeMetadata.Status.Mountpoint, nil
}

func (driver *mock) Unmount(name string, id string) error {
	_, existed := driver.volumeMetadataMap[name]
	if !existed {
		return fmt.Errorf("volume %s does not exist", name)
	}
	return nil
}

func (driver *mock) Destroy() error {
	return nil
}
