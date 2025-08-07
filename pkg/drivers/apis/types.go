package apis

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
)

var globalValidator *validator.Validate = validator.New()

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

func (vm *VolumeMetadata) Marshal() (data []byte, err error) {
	err = globalValidator.Struct(vm)
	if err != nil {
		return data, fmt.Errorf("failed to validate volume metadata: %v", err)
	}

	data, err = json.Marshal(vm)
	if err != nil {
		return data, fmt.Errorf("failed to unmarshal volume metadata: %v", err)
	}
	return data, err
}

func (vm *VolumeMetadata) Unmarshal(data []byte) (err error) {
	err = json.Unmarshal(data, vm)
	if err != nil {
		return fmt.Errorf("failed to unmarshal volume metadata: %v", err)
	}

	err = globalValidator.Struct(vm)
	if err != nil {
		return fmt.Errorf("failed to validate volume metadata: %v", err)
	}

	return nil
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

func (status *VolumeStatus) IsMounted() bool {
	return len(status.MountBy) > 0
}

func (status *VolumeStatus) IsMountedBy(id string) bool {
	return slices.Contains(status.MountBy, id)
}

func (status *VolumeStatus) AddMount(id string) {
	if !slices.Contains(status.MountBy, id) {
		status.MountBy = append(status.MountBy, id)
	}
}

func (status *VolumeStatus) RemoveMount(id string) {
	status.MountBy = slices.DeleteFunc(status.MountBy, func(mountID string) bool { return mountID == id })
}
