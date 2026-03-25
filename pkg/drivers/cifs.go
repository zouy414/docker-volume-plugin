package drivers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/zouy414/docker-volume-plugin/pkg/drivers/apis"
	"github.com/zouy414/docker-volume-plugin/pkg/drivers/storage/badger"
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
	db       *badger.DB
	rootPath string
	lock     *sync.RWMutex
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
		db:       badger.New(logger.WithService("badger").WithLogLevel(log.WarnLevel), path.Join(propagatedMountpoint, "metadata.db")),
		rootPath: propagatedMountpoint,
		lock:     &sync.RWMutex{},
	}, nil
}

func (driver *cifs) Create(name string, options map[string]string) error {
	driver.lock.Lock()
	defer driver.lock.Unlock()

	spec := apis.VolumeSpec{
		PurgeAfterDelete: driver.opts.PurgeAfterDelete,
	}
	if err := spec.Unmarshal(options); err != nil {
		return err
	}

	return driver.db.CreateVolumeMetadata(name, func(volumeMetadata *apis.VolumeMetadata) error {
		*volumeMetadata = apis.VolumeMetadata{
			Mountpoint: path.Join(name, "_data"),
			CreatedAt:  time.Now(),
			Spec:       &spec,
			Status:     &apis.VolumeStatus{},
		}

		return os.MkdirAll(path.Join(driver.rootPath, volumeMetadata.Mountpoint), 0755)
	})
}

func (driver *cifs) List() (map[string]*apis.VolumeMetadata, error) {
	driver.lock.Lock()
	defer driver.lock.Unlock()

	return driver.db.GetVolumeMetadataMap()
}

func (driver *cifs) Get(name string) (*apis.VolumeMetadata, error) {
	driver.lock.Lock()
	defer driver.lock.Unlock()

	return driver.db.GetVolumeMetadata(name)
}

func (driver *cifs) Remove(name string) error {
	driver.lock.Lock()
	defer driver.lock.Unlock()

	return driver.db.DeleteVolumeMetadata(name, func(volumeMetadata *apis.VolumeMetadata) error {
		if !volumeMetadata.Spec.PurgeAfterDelete {
			return nil
		}

		return os.RemoveAll(path.Join(driver.rootPath, name))
	})
}

func (driver *cifs) Path(name string) (string, error) {
	driver.lock.Lock()
	defer driver.lock.Unlock()

	volumeMetadata, err := driver.db.GetVolumeMetadata(name)

	return volumeMetadata.Mountpoint, err
}

func (driver *cifs) Mount(name string, id string) (string, error) {
	driver.lock.Lock()
	defer driver.lock.Unlock()

	return path.Join(name, "_data"), driver.db.SetVolumeMetadata(name, func(volumeMetadata *apis.VolumeMetadata) error {
		// Do nothing
		return nil
	})
}

func (driver *cifs) Unmount(name string, id string) error {
	driver.lock.Lock()
	defer driver.lock.Unlock()

	return driver.db.SetVolumeMetadata(name, func(volumeMetadata *apis.VolumeMetadata) error {
		// Do nothing
		return nil
	})
}

func (driver *cifs) Destroy() error {
	if err := driver.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %s", err)
	}

	if !driver.opts.Mock {
		if err := utils.Umount(driver.rootPath); err != nil {
			return fmt.Errorf("failed to unmount CIFS mount root path %s: %s", driver.rootPath, err)
		}
	}

	return nil
}
