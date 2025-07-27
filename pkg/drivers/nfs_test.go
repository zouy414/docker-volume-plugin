package drivers

import (
	"context"
	"docker-volume-plugin/pkg/log"
	"os"
	"path"
	"testing"
)

var localNFSServerDriverOptions string = `{
	"address": "nfs-server.mock",
	"remotePath": "/mock"
}`

func TestNFSDriver(t *testing.T) {
	propagatedMountpoint := path.Join(os.TempDir(), "net-volume-nfs-test")
	driver, err := New(context.Background(), log.New("test-nfs"), "nfs", propagatedMountpoint, localNFSServerDriverOptions)
	if err != nil {
		t.Fatalf("got error when new nfs driver: %v", err)
	}
	defer func() {
		if err := driver.Destroy(); err != nil {
			t.Errorf("got error when destroy nfs driver: %v", err)
		}
		if err := os.RemoveAll(propagatedMountpoint); err != nil {
			t.Errorf("got error when remove propagated mountpoint %s: %v", propagatedMountpoint, err)
		}
	}()

	// Test Create
	err = driver.Create("test", map[string]string{"purgeAfterDelete": "true"})
	if err != nil {
		t.Fatalf("got error when create volume test for nfs driver: %v", err)
	}

	// Test List
	volumeMetadataMap, err := driver.List()
	if err != nil {
		t.Fatalf("got error when list volume for nfs driver: %v", err)
	}
	if len(volumeMetadataMap) != 1 {
		t.Errorf("expected 1 volumes, got %d volume for nfs driver", len(volumeMetadataMap))
	}

	// Test Get
	_, err = driver.Get("test")
	if err != nil {
		t.Fatalf("got error when get volume tes for nfs driver: %v", err)
	}

	// Test Path
	mountpoint, err := driver.Path("test")
	if err != nil {
		t.Fatalf("got error when get volume test for nfs driver: %v", err)
	}
	if mountpoint != path.Join("test", "_data") {
		t.Errorf("expected mountpoint %s, got %s for nfs driver", path.Join("test", "_data"), mountpoint)
	}
	_, err = driver.Path("non-exist")
	if err == nil {
		t.Fatalf("expect got error when path volume non-exist")
	}

	// Test Mount
	for _, id := range []string{"1", "2"} {
		_, err = driver.Mount("test", id)
		if err != nil {
			t.Fatalf("got error when mount volume with %s: %v", id, err)
		}
		_, err = driver.Mount("test", id)
		if err == nil {
			t.Fatalf("expect got error when mount mounted volume")
		}
	}
	_, err = driver.Mount("non-exist", "1")
	if err == nil {
		t.Fatalf("expect got error when mount not existed volume")
	}

	// Test Remove mounted volume
	err = driver.Remove("test")
	if err == nil {
		t.Fatalf("expect got error when remove mounted volume")
	}

	// Test Unmount
	for _, id := range []string{"1", "2"} {
		err = driver.Unmount("test", id)
		if err != nil {
			t.Fatalf("got error when umount volume by %s: %v", id, err)
		}
		err = driver.Unmount("test", id)
		if err == nil {
			t.Fatalf("expect got error when umount unmounted volume test")
		}
	}

	// Test Remove unmounted volume
	err = driver.Remove("test")
	if err != nil {
		t.Fatalf("got error when remove volume test for nfs driver: %v", err)
	}

	// Test List after remove
	volumeMetadataMap, err = driver.List()
	if err != nil {
		t.Fatalf("got error when list volume for nfs driver: %v", err)
	}
	if len(volumeMetadataMap) != 0 {
		t.Errorf("expected 0 volumes, got %d volume for nfs driver", len(volumeMetadataMap))
	}
}
