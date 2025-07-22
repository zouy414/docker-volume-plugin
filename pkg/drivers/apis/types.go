package apis

import (
	"time"
)

// Driver interface
type Driver interface {
	// Create a new volume with the given name and options.
	Create(name string, options map[string]string) error
	// List all volumes.
	// Returns a map of volume names to their metadata.
	List() (map[string]*VolumeMetadata, error)
	// Get retrieves the metadata for a volume by name.
	Get(name string) (*VolumeMetadata, error)
	// Remove deletes a volume by name.
	Remove(name string) error
	// Path returns the mount point for a volume by name.
	Path(name string) (string, error)
	// Mount mounts a volume by name and ID.
	Mount(name string, id string) (string, error)
	// Unmount unmounts a volume by name and ID.
	Unmount(name string, id string) error
	// Destroy cleans up any resources used by the driver.
	Destroy() error
}

type VolumeStatus struct {
	MountBy string `json:"mountBy,omitempty"`
}

type VolumeMetadata struct {
	Mountpoint string        `json:"mountpoint,omitempty"`
	CreatedAt  time.Time     `json:"createAt"`
	Status     *VolumeStatus `json:"status,omitempty"`
}
