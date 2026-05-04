package storage

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/gofrs/flock"
	"github.com/zouy414/docker-volume-plugin/pkg/drivers/apis"
	"github.com/zouy414/docker-volume-plugin/pkg/log"
)

type Builtin struct {
	logger           *log.Logger
	rootPath         string
	dataDirName      string
	metadataFileName string
	metadataLockName string
}

func (s *Builtin) getMountpointPath(name string) string {
	return path.Join(name, s.dataDirName)
}

func (s *Builtin) getDataDirPath(name string) string {
	return path.Join(s.rootPath, s.getMountpointPath(name))
}

func (s *Builtin) getMetadataFilePath(name string) string {
	return path.Join(s.rootPath, name, s.metadataFileName)
}

func (s *Builtin) acquireMetadataLock(name string) (*flock.Flock, error) {
	lock := flock.New(path.Join(s.rootPath, name, s.metadataLockName))

	err := lock.Lock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire metadata lock: %v", err)
	}

	return lock, nil
}

// New creates a new instance of the Storage struct with the provided logger and path.
// It initializes the badger options and sets up a file lock to ensure that only one instance of the database can be accessed at a time.
func NewBuiltin(logger *log.Logger, rootPath string) *Builtin {
	return &Builtin{
		logger:           logger,
		rootPath:         rootPath,
		dataDirName:      "_data",
		metadataFileName: "_metadata.json",
		metadataLockName: "_metadata.json.lock",
	}
}

// CreateVolume creates a volume entry
func (s *Builtin) CreateVolume(name string, spec *apis.VolumeSpec) error {
	metadata := &apis.VolumeMetadata{
		CreatedAt: time.Now(),
		Spec:      spec,
		Status: &apis.VolumeStatus{
			Mountpoint: s.getMountpointPath(name),
		},
	}

	err := os.MkdirAll(s.getDataDirPath(name), 0755)
	if err != nil {
		return fmt.Errorf("failed to create volume directory: %v", err)
	}

	lock, err := s.acquireMetadataLock(name)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %v", err)
	}
	defer lock.Unlock()

	if _, err := os.Stat(s.getMetadataFilePath(name)); err == nil {
		s.logger.Warningf("volume %s already exists, skipping creation", name)
		return nil
	}

	data, err := metadata.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal volume metadata: %v", err)
	}

	return os.WriteFile(s.getMetadataFilePath(name), data, 0644)
}

// GetVolumeMetadata retrieves the volume metadata for the specified volume name
func (s *Builtin) GetVolumeMetadata(name string) (*apis.VolumeMetadata, error) {
	data, err := os.ReadFile(s.getMetadataFilePath(name))
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %v", err)
	}

	metadata := &apis.VolumeMetadata{}
	err = metadata.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal volume metadata: %v", err)
	}

	return metadata, nil
}

// GetVolumeMetadataMap retrieves a map of all volume metadata entries, where the keys are the volume names and the values are the corresponding volume metadata.
func (s *Builtin) GetVolumeMetadataMap() (map[string]*apis.VolumeMetadata, error) {
	volumeMetadataMap := make(map[string]*apis.VolumeMetadata)
	entries, err := os.ReadDir(s.rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read root directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		metadata, err := s.GetVolumeMetadata(entry.Name())
		if err != nil {
			s.logger.Warningf("failed to get metadata for volume %s: %v", entry.Name(), err)
			continue
		}

		volumeMetadataMap[entry.Name()] = metadata
	}

	return volumeMetadataMap, nil
}

// DeleteVolumeMetadata deletes the volume metadata for the specified volume name
func (s *Builtin) DeleteVolumeMetadata(name string) error {
	lock, err := s.acquireMetadataLock(name)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %v", err)
	}
	defer lock.Unlock()

	return os.Remove(s.getMetadataFilePath(name))
}

// DeleteVolume deletes the volume and its metadata for the specified volume name
func (s *Builtin) DeleteVolume(name string) error {
	return os.RemoveAll(path.Join(s.rootPath, name))
}

// Close releases any resources held by the DB instance, such as the file lock. It should be called when the DB instance is no longer needed to ensure proper cleanup.
func (s *Builtin) Close() error {
	// Do nothing
	return nil
}
