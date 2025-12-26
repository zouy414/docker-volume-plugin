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
	opts := &nfsOptions{
		MountOptions:       []string{"nfsvers=4", "rw", "noatime", "rsize=8192", "wsize=8192", "tcp", "timeo=14", "sync"},
		PurgeAfterDelete:   false,
		AllowMultipleMount: true,
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

type nfsOptions struct {
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
	opts         *nfsOptions
	db           *badger.DB
	rootPath     string
	lock         *sync.RWMutex
	reservedPath []string
}

func (n *nfs) Create(name string, options map[string]string) error {
	n.lock.Lock()
	defer n.lock.Unlock()

	if slices.Contains(n.reservedPath, name) {
		return fmt.Errorf("volume name %s is reserved, please choose a different name", name)
	}

	spec := apis.VolumeSpec{
		PurgeAfterDelete:   n.opts.PurgeAfterDelete,
		AllowMultipleMount: n.opts.AllowMultipleMount,
	}
	if err := spec.Unmarshal(options); err != nil {
		return err
	}

	return n.db.CreateVolumeMetadata(name, func(volumeMetadata *apis.VolumeMetadata) error {
		*volumeMetadata = apis.VolumeMetadata{
			Mountpoint: path.Join(name, "_data"),
			CreatedAt:  time.Now(),
			Spec:       &spec,
			Status: &apis.VolumeStatus{
				MountBy: []string{},
			},
		}

		return os.MkdirAll(path.Join(n.rootPath, volumeMetadata.Mountpoint), 0755)
	})
}

func (n *nfs) List() (map[string]*apis.VolumeMetadata, error) {
	n.lock.Lock()
	defer n.lock.Unlock()

	return n.db.GetVolumeMetadataMap()
}

func (n *nfs) Get(name string) (*apis.VolumeMetadata, error) {
	n.lock.Lock()
	defer n.lock.Unlock()

	return n.db.GetVolumeMetadata(name)
}

func (n *nfs) Remove(name string) error {
	n.lock.Lock()
	defer n.lock.Unlock()

	return n.db.DeleteVolumeMetadata(name, func(volumeMetadata *apis.VolumeMetadata) error {
		if len(volumeMetadata.Status.MountBy) != 0 {
			return fmt.Errorf("volume %s is mounted by %s, unmount it before removing", name, volumeMetadata.Status.MountBy)
		}

		if volumeMetadata.Spec.PurgeAfterDelete {
			err := os.RemoveAll(path.Join(n.rootPath, name))
			if err != nil {
				return fmt.Errorf("failed to remove volume data: %s", err)
			}
		}
		return nil
	})
}

func (n *nfs) Path(name string) (string, error) {
	n.lock.Lock()
	defer n.lock.Unlock()

	volumeMetadata, err := n.db.GetVolumeMetadata(name)

	return volumeMetadata.Mountpoint, err
}

func (n *nfs) Mount(name string, id string) (string, error) {
	n.lock.Lock()
	defer n.lock.Unlock()

	return path.Join(name, "_data"), n.db.SetVolumeMetadata(name, func(volumeMetadata *apis.VolumeMetadata) error {
		if (!volumeMetadata.Spec.AllowMultipleMount && volumeMetadata.Status.IsMounted()) || volumeMetadata.Status.IsMountedBy(id) {
			return fmt.Errorf("volume %s is already mounted", name)
		}

		volumeMetadata.Status.AddMount(id)
		return nil
	})
}

func (n *nfs) Unmount(name string, id string) error {
	n.lock.Lock()
	defer n.lock.Unlock()

	return n.db.SetVolumeMetadata(name, func(volumeMetadata *apis.VolumeMetadata) error {
		if !volumeMetadata.Status.IsMountedBy(id) {
			return fmt.Errorf("volume %s is not mounted by %s", name, id)
		}

		volumeMetadata.Status.RemoveMount(id)
		return nil
	})
}

func (n *nfs) Destroy() error {
	err := n.db.Close()
	if err != nil {
		return fmt.Errorf("failed to close database: %s", err)
	}

	if !n.opts.Mock {
		err = utils.Umount(n.rootPath)
		if err != nil {
			return fmt.Errorf("failed to unmount NFS mount root path %s: %s", n.rootPath, err)
		}
	}

	return nil
}
