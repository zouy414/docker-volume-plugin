package badger

import (
	"encoding/json"
	"fmt"
	"strings"

	"net-volume-plugins/pkg/drivers/apis"
	"net-volume-plugins/pkg/log"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/gofrs/flock"
)

type ActionCallback func(volumeMetadata *apis.VolumeMetadata) error

type DB struct {
	path                 string
	flock                *flock.Flock
	defaultBadgerOptions badger.Options
}

func NewBadgerDB(logger *log.Logger, path string, lock string) *DB {
	defaultBadgerOptions := badger.DefaultOptions(path)
	defaultBadgerOptions.Logger = logger
	return &DB{
		path:                 path,
		flock:                flock.New(lock),
		defaultBadgerOptions: defaultBadgerOptions,
	}
}

func (b *DB) CreateVolumeMetadata(name string, action ActionCallback) error {
	err := b.flock.Lock()
	if err != nil {
		return fmt.Errorf("failed to get flock: %v", err)
	}
	defer b.flock.Unlock()

	db, err := badger.Open(b.defaultBadgerOptions)
	if err != nil {
		return fmt.Errorf("failed to open badger database: %v", err)
	}
	defer db.Close()

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
		return fmt.Errorf("failed to open badger database: %v", err)
	}

	txn := db.NewTransaction(true)
	defer txn.Discard()

	volumeMetadata := &apis.VolumeMetadata{}
	err = action(volumeMetadata)
	if err != nil {
		return fmt.Errorf("failed to execute action: %v", err)
	}

	value, err := json.Marshal(volumeMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal volume metadata: %v", err)
	}

	err = txn.Set([]byte(name), []byte(value))
	if err != nil {
		return fmt.Errorf("failed to set volume metadata in database: %v", err)
	}

	return txn.Commit()
}

func (b *DB) GetVolumeMetadata(name string) (*apis.VolumeMetadata, error) {
	err := b.flock.Lock()
	if err != nil {
		return &apis.VolumeMetadata{}, fmt.Errorf("failed to get flock: %v", err)
	}
	defer b.flock.Unlock()

	db, err := badger.Open(b.defaultBadgerOptions)
	if err != nil {
		return &apis.VolumeMetadata{}, fmt.Errorf("failed to open badger database: %v", err)
	}
	defer db.Close()

	return getVolumeMetadata(db, name)
}

func (b *DB) GetVolumeMetadataMap() (map[string]*apis.VolumeMetadata, error) {
	volumeMetadataMap := make(map[string]*apis.VolumeMetadata)

	err := b.flock.Lock()
	if err != nil {
		return volumeMetadataMap, nil
	}
	defer b.flock.Unlock()

	db, err := badger.Open(b.defaultBadgerOptions)
	if err != nil {
		return volumeMetadataMap, fmt.Errorf("failed to open badger database: %v", err)
	}
	defer db.Close()

	err = db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			volumeMetadata := &apis.VolumeMetadata{}
			err = item.Value(func(val []byte) error { return json.Unmarshal(val, volumeMetadata) })
			if err != nil {
				return err
			}

			volumeMetadataMap[string(item.Key())] = volumeMetadata
		}

		return nil
	})

	return volumeMetadataMap, err
}

func (b *DB) SetVolumeMetadata(name string, action ActionCallback) error {
	err := b.flock.Lock()
	if err != nil {
		return fmt.Errorf("failed to get flock: %v", err)
	}
	defer b.flock.Unlock()

	db, err := badger.Open(b.defaultBadgerOptions)
	if err != nil {
		return fmt.Errorf("failed to open badger database: %v", err)
	}
	defer db.Close()

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

	value, err := json.Marshal(volumeMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal volume metadata: %v", err)
	}

	err = txn.Set([]byte(name), []byte(value))
	if err != nil {
		return fmt.Errorf("failed to set volume metadata in database: %v", err)
	}

	return txn.Commit()
}

func (b *DB) DeleteVolumeMetadata(name string, action ActionCallback) error {
	err := b.flock.Lock()
	if err != nil {
		return fmt.Errorf("failed to get flock: %v", err)
	}
	defer b.flock.Unlock()

	db, err := badger.Open(b.defaultBadgerOptions)
	if err != nil {
		return fmt.Errorf("failed to open badger database: %v", err)
	}
	defer db.Close()

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
		return item.Value(func(val []byte) error { return json.Unmarshal(val, volumeMetadata) })
	})

	return volumeMetadata, err
}
