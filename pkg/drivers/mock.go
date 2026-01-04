package drivers

import (
	"context"
	"docker-volume-plugin/pkg/drivers/apis"
	"docker-volume-plugin/pkg/log"
	"fmt"
	"os"
	"path"
	"time"
)

func init() {
	registerFactory("mock", mockFactory)
}

func mockFactory(ctx context.Context, logger *log.Logger, propagatedMountpoint string, driverOptions string) (apis.Driver, error) {
	err := os.MkdirAll(propagatedMountpoint, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create mount point directory: %s", err)
	}
	return &mock{
		logger:               logger,
		propagatedMountpoint: propagatedMountpoint,
		volumeMetadataMap:    make(map[string]*apis.VolumeMetadata),
	}, nil
}

type mock struct {
	logger               *log.Logger
	propagatedMountpoint string
	volumeMetadataMap    map[string]*apis.VolumeMetadata
}

func (driver *mock) Create(name string, options map[string]string) error {
	driver.volumeMetadataMap[name] = &apis.VolumeMetadata{
		Mountpoint: name,
		CreatedAt:  time.Now(),
		Spec:       &apis.VolumeSpec{},
		Status:     &apis.VolumeStatus{},
	}
	return os.MkdirAll(path.Join(driver.propagatedMountpoint, driver.volumeMetadataMap[name].Mountpoint), 0755)
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
	return volumeMetadata.Mountpoint, nil
}

func (driver *mock) Mount(name string, id string) (string, error) {
	volumeMetadata, existed := driver.volumeMetadataMap[name]
	if !existed {
		return "", fmt.Errorf("volume %s does not exist", name)
	}
	return volumeMetadata.Mountpoint, nil
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
