package adapters

import (
	"context"
	"docker-volume-plugin/pkg/drivers"
	"docker-volume-plugin/pkg/drivers/apis"
	"docker-volume-plugin/pkg/log"
	"path"
	"strings"
	"time"

	"github.com/docker/go-plugins-helpers/volume"
)

type VolumePlugin struct {
	driverInstance apis.Driver
	logger         *log.Logger
	mountpointBase string
	volume.Driver
}

func NewVolumePlugin(ctx context.Context, logger *log.Logger, driver string, driverOptions string) (*VolumePlugin, error) {
	driverInstance, err := drivers.New(ctx, logger.WithService("nfs"), driver, volume.DefaultDockerRootDirectory, driverOptions)
	if err != nil {
		return nil, err
	}

	return &VolumePlugin{
		driverInstance: driverInstance,
		logger:         logger,
		mountpointBase: volume.DefaultDockerRootDirectory,
	}, nil
}

func (d *VolumePlugin) Create(req *volume.CreateRequest) error {
	err := d.driverInstance.Create(req.Name, req.Options)
	if strings.Contains(err.Error(), "already created") {
		d.logger.Infof("volume %s already exists, skipping creation", req.Name)
		return nil
	}

	return err
}

func (d *VolumePlugin) List() (*volume.ListResponse, error) {
	listResponse := &volume.ListResponse{
		Volumes: make([]*volume.Volume, 0),
	}

	volumeMetadataMap, err := d.driverInstance.List()
	if err != nil {
		d.logger.Errorf("failed to list volumes: %v", err)
		return listResponse, err
	}

	for name, metadata := range volumeMetadataMap {
		listResponse.Volumes = append(listResponse.Volumes, &volume.Volume{
			Name:       name,
			Mountpoint: path.Join(d.mountpointBase, metadata.Mountpoint),
			CreatedAt:  metadata.CreatedAt.Local().Format(time.RFC3339),
			Status: map[string]interface{}{
				"mountBy": metadata.Status.MountBy,
			},
		})
	}

	d.logger.Infof("find %d volumes", len(listResponse.Volumes))

	return listResponse, nil
}

func (d *VolumePlugin) Get(req *volume.GetRequest) (*volume.GetResponse, error) {
	getResponse := &volume.GetResponse{}

	metadata, err := d.driverInstance.Get(req.Name)
	if err != nil {
		d.logger.Errorf("failed to get volume %s: %v", req.Name, err)
		return getResponse, err
	}

	getResponse.Volume = &volume.Volume{
		Name:       req.Name,
		Mountpoint: path.Join(d.mountpointBase, metadata.Mountpoint),
		CreatedAt:  metadata.CreatedAt.Local().Format(time.RFC3339),
		Status: map[string]interface{}{
			"mountBy": metadata.Status.MountBy,
		},
	}

	return getResponse, nil
}

func (d *VolumePlugin) Remove(req *volume.RemoveRequest) error {
	return d.driverInstance.Remove(req.Name)
}

func (d *VolumePlugin) Path(req *volume.PathRequest) (*volume.PathResponse, error) {
	pathResponse := &volume.PathResponse{}

	mountpoint, err := d.driverInstance.Path(req.Name)
	if err != nil {
		d.logger.Errorf("failed to get path of volume %s: %v", req.Name, err)
		return pathResponse, err
	}

	pathResponse.Mountpoint = path.Join(d.mountpointBase, mountpoint)

	return pathResponse, nil
}

func (d *VolumePlugin) Mount(req *volume.MountRequest) (*volume.MountResponse, error) {
	mountResponse := &volume.MountResponse{}

	mountpoint, err := d.driverInstance.Mount(req.Name, req.ID)
	if err != nil {
		d.logger.Errorf("failed to mount volume %s: %v", req.Name, err)
		return mountResponse, err
	}
	mountResponse.Mountpoint = path.Join(d.mountpointBase, mountpoint)

	return mountResponse, nil
}

func (d *VolumePlugin) Unmount(req *volume.UnmountRequest) error {
	return d.driverInstance.Unmount(req.Name, req.ID)
}

func (d *VolumePlugin) Capabilities() *volume.CapabilitiesResponse {
	return &volume.CapabilitiesResponse{Capabilities: volume.Capability{Scope: "local"}}
}

func (d *VolumePlugin) Destroy() error {
	// Cleanup resources if needed
	if d.driverInstance != nil {
		return d.driverInstance.Destroy()
	}
	return nil
}
