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

type VolumeMetadata struct {
	Mountpoint string        `json:"mountpoint" validate:"required"`
	CreatedAt  time.Time     `json:"createAt" validate:"required"`
	Spec       *VolumeSpec   `json:"spec" validate:"required"`
	Status     *VolumeStatus `json:"status" validate:"required"`
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
