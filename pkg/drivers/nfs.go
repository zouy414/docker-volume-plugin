package drivers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zouy414/docker-volume-plugin/pkg/drivers/apis"
	"github.com/zouy414/docker-volume-plugin/pkg/drivers/storage"
	"github.com/zouy414/docker-volume-plugin/pkg/log"
	"github.com/zouy414/docker-volume-plugin/pkg/utils"
)

func init() {
	registerFactory("nfs", nfsFactory)
}

// nfs is an implementation of the Driver interface for managing volumes on an NFS share.
type nfs struct {
	logger   *log.Logger
	opts     *nfsDriverOptions
	storage  *storage.Builtin
	rootPath string
}

type nfsDriverOptions struct {
	// Address of NFS server
	Address string `json:"address"`

	// RemotePath of NFS exported
	RemotePath string `json:"remotePath"`

	// MountOptions for NFS
	MountOptions []string `json:"mountOptions,omitempty"`

	// PurgeAfterDelete indicates whether to purge the volume data after deletion
	PurgeAfterDelete bool `json:"purgeAfterDelete,omitempty"`

	// Mock indicates whether to run in mock mode (no actual NFS mount)
	Mock bool `json:"mock,omitempty"`
}

func nfsFactory(ctx context.Context, logger *log.Logger, propagatedMountpoint string, driverOptions string) (apis.Driver, error) {
	opts := &nfsDriverOptions{
		MountOptions:     []string{"nfsvers=4", "rw", "noatime", "rsize=8192", "wsize=8192", "tcp", "timeo=14", "sync"},
		PurgeAfterDelete: false,
		Mock:             false,
	}
	if err := json.Unmarshal([]byte(driverOptions), opts); err != nil {
		return nil, fmt.Errorf("failed to parse driver options: %s", err)
	}

	// Mount NFS share to a local mount point
	if opts.Mock {
		logger.Warning("Mock mode enabled, no actual NFS mount will be performed")
		if err := utils.MountMock(propagatedMountpoint); err != nil {
			return nil, fmt.Errorf("failed to create mock mount point: %s", err)
		}
	} else {
		if err := utils.MountNFS(opts.Address, opts.RemotePath, propagatedMountpoint, opts.MountOptions); err != nil {
			return nil, fmt.Errorf("failed to mount NFS share: %s", err)
		}
	}

	return &nfs{
		logger:   logger,
		opts:     opts,
		storage:  storage.NewBuiltin(logger.WithService("storage").WithLogLevel(log.WarnLevel), propagatedMountpoint),
		rootPath: propagatedMountpoint,
	}, nil
}

func (driver *nfs) Create(name string, options map[string]string) error {
	spec := &apis.VolumeSpec{
		PurgeAfterDelete: driver.opts.PurgeAfterDelete,
	}
	if err := spec.Unmarshal(options); err != nil {
		return err
	}

	return driver.storage.CreateVolume(name, spec)
}

func (driver *nfs) List() (map[string]*apis.VolumeMetadata, error) {
	return driver.storage.GetVolumeMetadataMap()
}

func (driver *nfs) Get(name string) (*apis.VolumeMetadata, error) {
	return driver.storage.GetVolumeMetadata(name)
}

func (driver *nfs) Remove(name string) error {
	metadata, err := driver.storage.GetVolumeMetadata(name)
	if err != nil {
		return fmt.Errorf("failed to get volume metadata: %s", err)
	}

	err = driver.storage.DeleteVolumeMetadata(name)
	if err != nil {
		return fmt.Errorf("failed to delete volume metadata: %s", err)
	}

	if metadata.Spec.PurgeAfterDelete {
		err = driver.storage.DeleteVolume(name)
		if err != nil {
			return fmt.Errorf("failed to delete volume data: %s", err)
		}
	}

	return nil
}

func (driver *nfs) Path(name string) (string, error) {
	volumeMetadata, err := driver.storage.GetVolumeMetadata(name)
	if err != nil {
		return "", err
	}
	return volumeMetadata.Status.Mountpoint, err
}

func (driver *nfs) Mount(name string, id string) (string, error) {
	return driver.Path(name)
}

func (driver *nfs) Unmount(name string, id string) error {
	_, err := driver.storage.GetVolumeMetadata(name)
	return err
}

func (driver *nfs) Destroy() error {
	err := driver.storage.Close()
	if err != nil {
		return fmt.Errorf("failed to close storage: %s", err)
	}

	if !driver.opts.Mock {
		err = utils.Umount(driver.rootPath)
		if err != nil {
			return fmt.Errorf("failed to unmount NFS mount root path %s: %s", driver.rootPath, err)
		}
	}

	return nil
}
