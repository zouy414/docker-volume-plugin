package apis

import (
	"fmt"
	"strconv"
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

type VolumeMetadata struct {
	Mountpoint string        `json:"mountpoint,omitempty"`
	CreatedAt  time.Time     `json:"createAt"`
	Spec       *VolumeSpec   `json:"spec"`
	Status     *VolumeStatus `json:"status"`
}

type VolumeSpec struct {
	PurgeAfterDelete   bool `json:"purgeAfterDelete,omitempty"`
	AllowMultipleMount bool `json:"allowMultipleMount,omitempty"`
}

func (spec *VolumeSpec) Unmarshal(data map[string]string) (err error) {
	for key, value := range data {
		switch key {
		case "purgeAfterDelete":
			spec.PurgeAfterDelete, err = strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid value for purgeAfterDelete: %v", err)
			}
		case "allowMultipleMount":
			spec.AllowMultipleMount, err = strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid value for allowMultipleMount: %v", err)
			}
		default:
			return fmt.Errorf("unknown option %s with value %s", key, value)
		}
	}

	return nil
}

type VolumeStatus struct {
	MountBy []string `json:"mountBy,omitempty"`
}
