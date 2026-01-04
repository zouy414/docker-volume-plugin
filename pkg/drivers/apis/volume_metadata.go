package apis

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"time"

	"github.com/docker/go-plugins-helpers/volume"
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

func (vm *VolumeMetadata) ToVolume(name string, mountpointBase string) *volume.Volume {
	return &volume.Volume{
		Name:       name,
		Mountpoint: path.Join(mountpointBase, vm.Mountpoint),
		CreatedAt:  vm.CreatedAt.Local().Format(time.RFC3339),
		Status:     map[string]interface{}{},
	}
}

type VolumeSpec struct {
	PurgeAfterDelete bool `json:"purgeAfterDelete,omitempty"`
}

func (spec *VolumeSpec) Unmarshal(data map[string]string) (err error) {
	for key, value := range data {
		switch key {
		case "purgeAfterDelete":
			spec.PurgeAfterDelete, err = strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid value for purgeAfterDelete: %v", err)
			}
		default:
			return fmt.Errorf("unknown option %s with value %s", key, value)
		}
	}

	return nil
}

type VolumeStatus struct{}
