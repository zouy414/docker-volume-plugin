package storage

import (
	"fmt"
	"os"
	"path"
	"sync"
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
	waitGroup        sync.WaitGroup
}

// New creates a new instance of the Storage struct with the provided logger and path.
func NewBuiltin(logger *log.Logger, rootPath string) *Builtin {
	return &Builtin{
		logger:           logger,
		rootPath:         rootPath,
		dataDirName:      "_data",
		metadataFileName: "_metadata.json",
		metadataLockName: "_metadata.json.lock",
		waitGroup:        sync.WaitGroup{},
	}
}

// CreateVolume creates a volume entry
func (s *Builtin) CreateVolume(name string, spec *apis.VolumeSpec) error {
	s.waitGroup.Add(1)
	defer s.waitGroup.Done()

	metadata := &apis.VolumeMetadata{
		CreatedAt: time.Now(),
		Spec:      spec,
		Status: &apis.VolumeStatus{
			Mountpoint: s.getMountpointPath(name),
		},
	}

	// Create the volume directory if it doesn't exist
	err := os.MkdirAll(s.getDataDirPath(name), 0755)
	if err != nil {
		return fmt.Errorf("failed to create volume directory: %v", err)
	}

	// Acquire a lock on the metadata file to prevent concurrent modifications
	lock, err := s.acquireMetadataLock(name)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %v", err)
	}
	defer func() {
		if err := lock.Unlock(); err != nil {
			s.logger.Errorf("failed to unlock flock: %v", err)
		}
	}()

	// Check if the metadata file already exists, which indicates that the volume already exists
	if _, err := os.Stat(s.getMetadataFilePath(name)); err == nil {
		s.logger.Warningf("volume %s already exists, skipping creation", name)
		return nil
	}

	// Marshal the volume metadata to JSON format and write it to the metadata file
	data, err := metadata.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal volume metadata: %v", err)
	}
	return os.WriteFile(s.getMetadataFilePath(name), data, 0644)
}

// FetchVolumeMetadata retrieves the volume metadata for the specified volume name
func (s *Builtin) FetchVolumeMetadata(name string) (*apis.VolumeMetadata, error) {
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

// ListVolumeMetadataMap retrieves a map of all volume metadata entries, where the keys are the volume names and the values are the corresponding volume metadata.
func (s *Builtin) ListVolumeMetadata() (map[string]*apis.VolumeMetadata, error) {
	volumeMetadataMap := make(map[string]*apis.VolumeMetadata)
	entries, err := os.ReadDir(s.rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read root directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		metadata, err := s.FetchVolumeMetadata(entry.Name())
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
	s.waitGroup.Add(1)
	defer s.waitGroup.Done()

	lock, err := s.acquireMetadataLock(name)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %v", err)
	}
	defer func() {
		if err := lock.Unlock(); err != nil {
			s.logger.Errorf("failed to unlock flock: %v", err)
		}
	}()

	return os.Remove(s.getMetadataFilePath(name))
}

// DeleteVolume deletes the volume and its metadata for the specified volume name
func (s *Builtin) DeleteVolume(name string) error {
	s.waitGroup.Add(1)
	defer s.waitGroup.Done()

	return os.RemoveAll(path.Join(s.rootPath, name))
}

// Close releases any resources held by the DB instance, such as the file lock. It should be called when the DB instance is no longer needed to ensure proper cleanup.
func (s *Builtin) Close() error {
	s.waitGroup.Wait()

	// Do nothing
	return nil
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
