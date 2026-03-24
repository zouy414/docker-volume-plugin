package badger

import (
	"fmt"
	"slices"
	"strings"

	"github.com/zouy414/docker-volume-plugin/pkg/drivers/apis"
	"github.com/zouy414/docker-volume-plugin/pkg/log"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/gofrs/flock"
)

type ActionCallback func(volumeMetadata *apis.VolumeMetadata) error

type DB struct {
	path          string
	logger        *log.Logger
	flock         *flock.Flock
	badgerOptions badger.Options
}

// New creates a new instance of the DB struct with the provided logger and path.
// It initializes the badger options and sets up a file lock to ensure that only one instance of the database can be accessed at a time.
func New(logger *log.Logger, path string) *DB {
	defaultBadgerOptions := badger.DefaultOptions(path)
	defaultBadgerOptions.Logger = logger

	return &DB{
		path:          path,
		logger:        logger,
		flock:         flock.New(path + ".lock"),
		badgerOptions: defaultBadgerOptions,
	}
}

// CreateVolumeMetadata creates a new volume metadata entry in the database with the specified name and action callback.
func (b *DB) CreateVolumeMetadata(name string, action ActionCallback) error {
	if slices.Contains([]string{b.path, b.path + ".lock"}, name) {
		return fmt.Errorf("volume name %s is reserved, please choose a different name", name)
	}

	err := b.flock.Lock()
	if err != nil {
		return fmt.Errorf("failed to get flock: %v", err)
	}
	defer func() {
		if err := b.flock.Unlock(); err != nil {
			b.logger.Errorf("failed to unlock flock: %v", err)
		}
	}()

	db, err := badger.Open(b.badgerOptions)
	if err != nil {
		return fmt.Errorf("failed to open badger database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			b.logger.Errorf("failed to close badger database: %v", err)
		}
	}()

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(name))
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return err
		}
		if item != nil {
			return fmt.Errorf("volume %s already created", name)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to read badger database: %v", err)
	}

	txn := db.NewTransaction(true)
	defer txn.Discard()

	volumeMetadata := &apis.VolumeMetadata{}
	err = action(volumeMetadata)
	if err != nil {
		return fmt.Errorf("failed to execute action: %v", err)
	}

	value, err := volumeMetadata.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal volume metadata: %v", err)
	}

	err = txn.Set([]byte(name), []byte(value))
	if err != nil {
		return fmt.Errorf("failed to set volume metadata in database: %v", err)
	}

	return txn.Commit()
}

// GetVolumeMetadata retrieves the volume metadata for the specified volume name from the database.
func (b *DB) GetVolumeMetadata(name string) (*apis.VolumeMetadata, error) {
	err := b.flock.Lock()
	if err != nil {
		return &apis.VolumeMetadata{}, fmt.Errorf("failed to get flock: %v", err)
	}
	defer func() {
		if err := b.flock.Unlock(); err != nil {
			b.logger.Errorf("failed to unlock flock: %v", err)
		}
	}()

	db, err := badger.Open(b.badgerOptions)
	if err != nil {
		return &apis.VolumeMetadata{}, fmt.Errorf("failed to open badger database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			b.logger.Errorf("failed to close badger database: %v", err)
		}
	}()

	return getVolumeMetadata(db, name)
}

// GetVolumeMetadataMap retrieves a map of all volume metadata entries in the database, where the keys are the volume names and the values are the corresponding volume metadata.
func (b *DB) GetVolumeMetadataMap() (map[string]*apis.VolumeMetadata, error) {
	err := b.flock.Lock()
	if err != nil {
		return make(map[string]*apis.VolumeMetadata), nil
	}
	defer func() {
		if err := b.flock.Unlock(); err != nil {
			b.logger.Errorf("failed to unlock flock: %v", err)
		}
	}()

	db, err := badger.Open(b.badgerOptions)
	if err != nil {
		return make(map[string]*apis.VolumeMetadata), fmt.Errorf("failed to open badger database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			b.logger.Errorf("failed to close badger database: %v", err)
		}
	}()

	volumeMetadataMap := make(map[string]*apis.VolumeMetadata)
	err = db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			volumeMetadata := &apis.VolumeMetadata{}
			err = item.Value(func(val []byte) error { return volumeMetadata.Unmarshal(val) })
			if err != nil {
				return err
			}

			volumeMetadataMap[string(item.Key())] = volumeMetadata
		}

		return nil
	})

	return volumeMetadataMap, err
}

// SetVolumeMetadata updates the volume metadata for the specified volume name in the database using the provided action callback to modify the existing metadata.
func (b *DB) SetVolumeMetadata(name string, action ActionCallback) error {
	err := b.flock.Lock()
	if err != nil {
		return fmt.Errorf("failed to get flock: %v", err)
	}
	defer func() {
		if err := b.flock.Unlock(); err != nil {
			b.logger.Errorf("failed to unlock flock: %v", err)
		}
	}()

	db, err := badger.Open(b.badgerOptions)
	if err != nil {
		return fmt.Errorf("failed to open badger database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			b.logger.Errorf("failed to close badger database: %v", err)
		}
	}()

	txn := db.NewTransaction(true)
	defer txn.Discard()

	volumeMetadata, err := getVolumeMetadata(db, name)
	if err != nil {
		return fmt.Errorf("failed to get %s volume metadata: %v", name, err)
	}

	err = action(volumeMetadata)
	if err != nil {
		return fmt.Errorf("failed to execute action: %v", err)
	}

	value, err := volumeMetadata.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal volume metadata: %v", err)
	}

	err = txn.Set([]byte(name), []byte(value))
	if err != nil {
		return fmt.Errorf("failed to set volume metadata in database: %v", err)
	}

	return txn.Commit()
}

// DeleteVolumeMetadata deletes the volume metadata for the specified volume name from the database using the provided action callback to perform any necessary cleanup before deletion.
func (b *DB) DeleteVolumeMetadata(name string, action ActionCallback) error {
	err := b.flock.Lock()
	if err != nil {
		return fmt.Errorf("failed to get flock: %v", err)
	}
	defer func() {
		if err := b.flock.Unlock(); err != nil {
			b.logger.Errorf("failed to unlock flock: %v", err)
		}
	}()

	db, err := badger.Open(b.badgerOptions)
	if err != nil {
		return fmt.Errorf("failed to open badger database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			b.logger.Errorf("failed to close badger database: %v", err)
		}
	}()

	volumeMetadata, err := getVolumeMetadata(db, name)
	if err != nil {
		return fmt.Errorf("failed to get %s volume metadata: %v", name, err)
	}

	txn := db.NewTransaction(true)
	defer txn.Discard()

	err = txn.Delete([]byte(name))
	if err != nil {
		return fmt.Errorf("failed to delete volume metadata in database: %v", err)
	}

	err = action(volumeMetadata)
	if err != nil {
		return fmt.Errorf("failed to execute action: %v", err)
	}

	return txn.Commit()
}

// Close releases any resources held by the DB instance, such as the file lock. It should be called when the DB instance is no longer needed to ensure proper cleanup.
func (b *DB) Close() error {
	return b.flock.Close()
}

func getVolumeMetadata(db *badger.DB, name string) (*apis.VolumeMetadata, error) {
	volumeMetadata := &apis.VolumeMetadata{}

	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(name))
		if err != nil {
			return err
		}
		if item == nil {
			return fmt.Errorf("volume %s not found", name)
		}
		return item.Value(func(val []byte) error { return volumeMetadata.Unmarshal(val) })
	})

	return volumeMetadata, err
}
