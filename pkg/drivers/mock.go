package drivers

import (
	"context"
	"docker-volume-plugin/pkg/drivers/apis"
	"docker-volume-plugin/pkg/log"
	"fmt"
	"os"
	"path"
	"slices"
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

func (m *mock) Create(name string, options map[string]string) error {
	m.volumeMetadataMap[name] = &apis.VolumeMetadata{
		Mountpoint: name,
		CreatedAt:  time.Now(),
		Spec:       &apis.VolumeSpec{},
		Status:     &apis.VolumeStatus{MountBy: []string{}},
	}
	return os.MkdirAll(path.Join(m.propagatedMountpoint, m.volumeMetadataMap[name].Mountpoint), 0755)
}

func (m *mock) List() (map[string]*apis.VolumeMetadata, error) {
	return m.volumeMetadataMap, nil
}

func (m *mock) Get(name string) (*apis.VolumeMetadata, error) {
	return m.volumeMetadataMap[name], nil
}

func (m *mock) Remove(name string) error {
	if m.volumeMetadataMap[name] == nil {
		return fmt.Errorf("volume %s does not exist", name)
	}
	if len(m.volumeMetadataMap[name].Status.MountBy) != 0 {
		return fmt.Errorf("volume %s is mounted by %s, unmount it before removing", name, m.volumeMetadataMap[name].Status.MountBy)
	}
	delete(m.volumeMetadataMap, name)
	return nil
}

func (m *mock) Path(name string) (string, error) {
	volumeMetadata, existed := m.volumeMetadataMap[name]
	if !existed {
		return "", fmt.Errorf("volume %s does not exist", name)
	}
	return volumeMetadata.Mountpoint, nil
}

func (m *mock) Mount(name string, id string) (string, error) {
	volumeMetadata, existed := m.volumeMetadataMap[name]
	if !existed {
		return "", fmt.Errorf("volume %s does not exist", name)
	}
	if slices.Contains(volumeMetadata.Status.MountBy, id) {
		return "", fmt.Errorf("volume %s is already mounted by %s", name, id)
	}
	volumeMetadata.Status.MountBy = append(volumeMetadata.Status.MountBy, id)
	return volumeMetadata.Mountpoint, nil
}

func (m *mock) Unmount(name string, id string) error {
	volumeMetadata, existed := m.volumeMetadataMap[name]
	if !existed {
		return fmt.Errorf("volume %s does not exist", name)
	}
	if !slices.Contains(volumeMetadata.Status.MountBy, id) {
		return fmt.Errorf("volume %s is not mounted by %s", name, id)
	}
	volumeMetadata.Status.MountBy = slices.DeleteFunc(volumeMetadata.Status.MountBy, func(mountID string) bool { return mountID == id })
	return nil
}

func (n *mock) Destroy() error {
	return nil
}
