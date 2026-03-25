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
	registerFactory("nfs", nfsFactory)
}

// nfs is an implementation of the Driver interface for managing volumes on an NFS share.
type nfs struct {
	logger   *log.Logger
	opts     *nfsDriverOptions
	db       *badger.DB
	rootPath string
	lock     *sync.RWMutex
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
		db:       badger.New(logger.WithService("badger").WithLogLevel(log.WarnLevel), path.Join(propagatedMountpoint, "metadata.db")),
		rootPath: propagatedMountpoint,
		lock:     &sync.RWMutex{},
	}, nil
}

func (driver *nfs) Create(name string, options map[string]string) error {
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

func (driver *nfs) List() (map[string]*apis.VolumeMetadata, error) {
	driver.lock.Lock()
	defer driver.lock.Unlock()

	return driver.db.GetVolumeMetadataMap()
}

func (driver *nfs) Get(name string) (*apis.VolumeMetadata, error) {
	driver.lock.Lock()
	defer driver.lock.Unlock()

	return driver.db.GetVolumeMetadata(name)
}

func (driver *nfs) Remove(name string) error {
	driver.lock.Lock()
	defer driver.lock.Unlock()

	return driver.db.DeleteVolumeMetadata(name, func(volumeMetadata *apis.VolumeMetadata) error {
		if !volumeMetadata.Spec.PurgeAfterDelete {
			return nil
		}

		return os.RemoveAll(path.Join(driver.rootPath, name))
	})
}

func (driver *nfs) Path(name string) (string, error) {
	driver.lock.Lock()
	defer driver.lock.Unlock()

	volumeMetadata, err := driver.db.GetVolumeMetadata(name)

	return volumeMetadata.Mountpoint, err
}

func (driver *nfs) Mount(name string, id string) (string, error) {
	driver.lock.Lock()
	defer driver.lock.Unlock()

	return path.Join(name, "_data"), driver.db.SetVolumeMetadata(name, func(volumeMetadata *apis.VolumeMetadata) error {
		// Do nothing
		return nil
	})
}

func (driver *nfs) Unmount(name string, id string) error {
	driver.lock.Lock()
	defer driver.lock.Unlock()

	return driver.db.SetVolumeMetadata(name, func(volumeMetadata *apis.VolumeMetadata) error {
		// Do nothing
		return nil
	})
}

func (driver *nfs) Destroy() error {
	err := driver.db.Close()
	if err != nil {
		return fmt.Errorf("failed to close database: %s", err)
	}

	if !driver.opts.Mock {
		err = utils.Umount(driver.rootPath)
		if err != nil {
			return fmt.Errorf("failed to unmount NFS mount root path %s: %s", driver.rootPath, err)
		}
	}

	return nil
}
