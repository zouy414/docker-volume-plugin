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
	driverInstance, err := drivers.New(ctx, logger.WithService(driver), driver, volume.DefaultDockerRootDirectory, driverOptions)
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
	d.logger.Debugf("creating volume %s with options %v", req.Name, req.Options)

	err := d.driverInstance.Create(req.Name, req.Options)
	if err != nil && strings.Contains(err.Error(), "already created") {
		d.logger.Warning("volume %s already exists, skipping creation", req.Name)
		return nil
	}

	d.logger.Debugf("created volume %s: %v", req.Name)

	return err
}

func (d *VolumePlugin) convertVolumeMetadataToVolume(name string, metadata *apis.VolumeMetadata) *volume.Volume {
	return &volume.Volume{
		Name:       name,
		Mountpoint: path.Join(d.mountpointBase, metadata.Mountpoint),
		CreatedAt:  metadata.CreatedAt.Local().Format(time.RFC3339),
		Status: map[string]interface{}{
			"mountBy": metadata.Status.MountBy,
		},
	}
}

func (d *VolumePlugin) List() (*volume.ListResponse, error) {
	d.logger.Debug("listing all volumes")

	listResponse := &volume.ListResponse{
		Volumes: make([]*volume.Volume, 0),
	}
	volumeMetadataMap, err := d.driverInstance.List()
	if err != nil {
		return listResponse, err
	}
	for name, metadata := range volumeMetadataMap {
		listResponse.Volumes = append(listResponse.Volumes, d.convertVolumeMetadataToVolume(name, metadata))
	}

	d.logger.Debugf("listed volumes: %v", listResponse.Volumes)

	return listResponse, nil
}

func (d *VolumePlugin) Get(req *volume.GetRequest) (*volume.GetResponse, error) {
	d.logger.Debugf("getting volume %s", req.Name)

	getResponse := &volume.GetResponse{}
	metadata, err := d.driverInstance.Get(req.Name)
	if err != nil {
		return getResponse, err
	}
	getResponse.Volume = d.convertVolumeMetadataToVolume(req.Name, metadata)

	d.logger.Debugf("got volume %s: %v", req.Name, getResponse.Volume)

	return getResponse, nil
}

func (d *VolumePlugin) Remove(req *volume.RemoveRequest) error {
	d.logger.Debugf("removing volume %s", req.Name)

	return d.driverInstance.Remove(req.Name)
}

func (d *VolumePlugin) Path(req *volume.PathRequest) (*volume.PathResponse, error) {
	d.logger.Debugf("getting path for volume %s", req.Name)

	pathResponse := &volume.PathResponse{}
	mountpoint, err := d.driverInstance.Path(req.Name)
	if err != nil {
		return pathResponse, err
	}
	pathResponse.Mountpoint = path.Join(d.mountpointBase, mountpoint)

	d.logger.Debugf("path for volume %s is %s", req.Name, pathResponse.Mountpoint)

	return pathResponse, nil
}

func (d *VolumePlugin) Mount(req *volume.MountRequest) (*volume.MountResponse, error) {
	d.logger.Debugf("mounting volume %s with ID %s", req.Name, req.ID)

	mountResponse := &volume.MountResponse{}
	mountpoint, err := d.driverInstance.Mount(req.Name, req.ID)
	if err != nil {
		return mountResponse, err
	}
	mountResponse.Mountpoint = path.Join(d.mountpointBase, mountpoint)

	d.logger.Debugf("mounted volume %s with ID %s at %s", req.Name, req.ID, mountResponse.Mountpoint)

	return mountResponse, nil
}

func (d *VolumePlugin) Unmount(req *volume.UnmountRequest) error {
	d.logger.Debugf("unmounting volume %s with ID %s", req.Name, req.ID)

	return d.driverInstance.Unmount(req.Name, req.ID)
}

func (d *VolumePlugin) Capabilities() *volume.CapabilitiesResponse {
	d.logger.Debug("getting capabilities of the volume plugin")

	return &volume.CapabilitiesResponse{Capabilities: volume.Capability{Scope: "global"}}
}

func (d *VolumePlugin) Destroy() error {
	return d.driverInstance.Destroy()
}
