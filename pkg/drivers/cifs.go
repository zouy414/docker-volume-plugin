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
	registerFactory("cifs", cifsFactory)
}

// cifs is an implementation of the Driver interface for managing volumes on a CIFS share.
type cifs struct {
	logger   *log.Logger
	opts     *cifsDriverOptions
	storage  *storage.Builtin
	rootPath string
}

type cifsDriverOptions struct {
	// Address of CIFS server
	Address string `json:"address"`

	// RemotePath of CIFS exported
	RemotePath string `json:"remotePath"`

	// Username for CIFS authentication
	Username string `json:"username"`

	// Password for CIFS authentication
	Password string `json:"password,omitempty"`

	// MountOptions for CIFS
	MountOptions []string `json:"mountOptions,omitempty"`

	// PurgeAfterDelete indicates whether to purge the volume data after deletion
	PurgeAfterDelete bool `json:"purgeAfterDelete,omitempty"`

	// Mock indicates whether to run in mock mode (no actual CIFS mount)
	Mock bool `json:"mock,omitempty"`
}

func cifsFactory(ctx context.Context, logger *log.Logger, propagatedMountpoint string, driverOptions string) (apis.Driver, error) {
	opts := &cifsDriverOptions{
		MountOptions:     []string{},
		PurgeAfterDelete: false,
		Mock:             false,
	}
	if err := json.Unmarshal([]byte(driverOptions), opts); err != nil {
		return nil, fmt.Errorf("failed to parse driver options: %s", err)
	}

	// Mount CIFS share to a local mount point
	if opts.Mock {
		logger.Warning("Mock mode enabled, no actual CIFS mount will be performed")
		if err := utils.MountMock(propagatedMountpoint); err != nil {
			return nil, fmt.Errorf("failed to create mock mount point: %s", err)
		}
	} else {
		if err := utils.MountCIFS(opts.Address, opts.RemotePath, propagatedMountpoint, opts.Username, opts.Password, opts.MountOptions); err != nil {
			logger.Warning(err)
			return nil, fmt.Errorf("failed to mount CIFS share: %s", err)
		}
	}

	return &cifs{
		logger:   logger,
		opts:     opts,
		storage:  storage.NewBuiltin(logger.WithService("storage").WithLogLevel(log.WarnLevel), propagatedMountpoint),
		rootPath: propagatedMountpoint,
	}, nil
}

func (driver *cifs) Create(name string, options map[string]string) error {
	spec := &apis.VolumeSpec{
		PurgeAfterDelete: driver.opts.PurgeAfterDelete,
	}
	if err := spec.Unmarshal(options); err != nil {
		return err
	}

	return driver.storage.CreateVolume(name, spec)
}

func (driver *cifs) List() (map[string]*apis.VolumeMetadata, error) {
	return driver.storage.GetVolumeMetadataMap()
}

func (driver *cifs) Get(name string) (*apis.VolumeMetadata, error) {
	return driver.storage.GetVolumeMetadata(name)
}

func (driver *cifs) Remove(name string) error {
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

func (driver *cifs) Path(name string) (string, error) {
	metadata, err := driver.storage.GetVolumeMetadata(name)
	if err != nil {
		return "", err
	}
	return metadata.Status.Mountpoint, err
}

func (driver *cifs) Mount(name string, id string) (string, error) {
	return driver.Path(name)
}

func (driver *cifs) Unmount(name string, id string) error {
	_, err := driver.storage.GetVolumeMetadata(name)
	return err
}

func (driver *cifs) Destroy() error {
	err := driver.storage.Close()
	if err != nil {
		return fmt.Errorf("failed to close storage: %s", err)
	}

	if !driver.opts.Mock {
		err = utils.Umount(driver.rootPath)
		if err != nil {
			return fmt.Errorf("failed to unmount CIFS mount root path %s: %s", driver.rootPath, err)
		}
	}

	return nil
}
