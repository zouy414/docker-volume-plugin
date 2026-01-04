package drivers

import (
	"context"
	"docker-volume-plugin/pkg/drivers/apis"
	"docker-volume-plugin/pkg/drivers/store/badger"
	"docker-volume-plugin/pkg/log"
	"docker-volume-plugin/pkg/utils"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"slices"
	"sync"
	"time"
)

func init() {
	registerFactory("nfs", nfsFactory)
}

func nfsFactory(ctx context.Context, logger *log.Logger, propagatedMountpoint string, driverOptions string) (apis.Driver, error) {
	opts := &nfsDriverOptions{
		MountOptions:       []string{"nfsvers=4", "rw", "noatime", "rsize=8192", "wsize=8192", "tcp", "timeo=14", "sync"},
		PurgeAfterDelete:   false,
		AllowMultipleMount: true,
		Mock:               false,
	}
	err := json.Unmarshal([]byte(driverOptions), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse driver options: %s", err)
	}

	// Create local mount point directory if not exists
	err = os.MkdirAll(propagatedMountpoint, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create NFS mount point directory: %s", err)
	}

	// Mount NFS share to a local mount point
	if opts.Mock {
		logger.Warning("Mock mode enabled, no actual NFS mount will be performed")
	} else {
		err = utils.MountNFS(opts.Address, opts.RemotePath, propagatedMountpoint, opts.MountOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to mount NFS share: %s", err)
		}
	}

	return &nfs{
		logger: logger,
		opts:   opts,
		db: badger.New(
			logger.WithService("badger").WithLogLevel(log.WarnLevel),
			path.Join(propagatedMountpoint, "metadata.db"),
		),
		rootPath:     propagatedMountpoint,
		lock:         &sync.RWMutex{},
		reservedPath: []string{"metadata.db", "metadata.db.lock"},
	}, nil
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

	// AllowMultipleMount indicates whether to allow multiple containers to mount the same volume
	AllowMultipleMount bool `json:"allowMultipleMount,omitempty"`

	// Mock indicates whether to run in mock mode (no actual NFS mount)
	Mock bool `json:"mock,omitempty"`
}

type nfs struct {
	logger       *log.Logger
	opts         *nfsDriverOptions
	db           *badger.DB
	rootPath     string
	lock         *sync.RWMutex
	reservedPath []string
}

func (driver *nfs) Create(name string, options map[string]string) error {
	driver.lock.Lock()
	defer driver.lock.Unlock()

	if slices.Contains(driver.reservedPath, name) {
		return fmt.Errorf("volume name %s is reserved, please choose a different name", name)
	}

	spec := apis.VolumeSpec{
		PurgeAfterDelete:   driver.opts.PurgeAfterDelete,
		AllowMultipleMount: driver.opts.AllowMultipleMount,
	}
	if err := spec.Unmarshal(options); err != nil {
		return err
	}

	return driver.db.CreateVolumeMetadata(name, func(volumeMetadata *apis.VolumeMetadata) error {
		*volumeMetadata = apis.VolumeMetadata{
			Mountpoint: path.Join(name, "_data"),
			CreatedAt:  time.Now(),
			Spec:       &spec,
			Status: &apis.VolumeStatus{
				MountBy: []string{},
			},
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
		if (!volumeMetadata.Spec.AllowMultipleMount && volumeMetadata.Status.IsMounted()) || volumeMetadata.Status.IsMountedBy(id) {
			return fmt.Errorf("volume %s is already mounted", name)
		}

		volumeMetadata.Status.AddMount(id)
		return nil
	})
}

func (driver *nfs) Unmount(name string, id string) error {
	driver.lock.Lock()
	defer driver.lock.Unlock()

	return driver.db.SetVolumeMetadata(name, func(volumeMetadata *apis.VolumeMetadata) error {
		if !volumeMetadata.Status.IsMountedBy(id) {
			driver.logger.Warningf("volume %s is not mounted by %s", name, id)
			return nil
		}

		volumeMetadata.Status.RemoveMount(id)
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
