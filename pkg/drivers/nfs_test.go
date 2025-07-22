package drivers

import (
	"context"
	"net-volume-plugins/pkg/log"
	"os"
	"path"
	"testing"
)

var localNFSServerDriverOptions string = `{
	"address": "nfs-server.mock",
	"remotePath": "/mock",
	"purgeAfterDelete": true
}`

func TestNFSDriver(t *testing.T) {
	driver, err := New(context.Background(), log.New("test-nfs"), "nfs", path.Join(os.TempDir(), "net-volume-nfs"), localNFSServerDriverOptions)
	if err != nil {
		t.Fatalf("got error when new nfs driver: %v", err)
	}
	defer driver.Destroy()
	defer os.RemoveAll(path.Join(os.TempDir(), "net-volume-nfs"))

	// Test Create
	err = driver.Create("test", nil)
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
	_, err = driver.Mount("test", "1")
	if err != nil {
		t.Fatalf("got error when mount volume test0: %v", err)
	}
	_, err = driver.Mount("test", "1")
	if err == nil {
		t.Fatalf("expect got error when mount mounted volume test")
	}
	_, err = driver.Mount("test", "2")
	if err == nil {
		t.Fatalf("expect got error when mount mounted volume test")
	}
	_, err = driver.Mount("non-exist", "1")
	if err == nil {
		t.Fatalf("expect got error when mount volume non-exist")
	}

	// Test Remove mounted volume
	err = driver.Remove("test")
	if err == nil {
		t.Fatalf("expect got error when remove mounted volume test")
	}

	// Test Unmount
	_, err = driver.Mount("test", "2")
	if err == nil {
		t.Fatalf("expect got error when unmount volume mount by other id")
	}
	err = driver.Unmount("test", "1")
	if err != nil {
		t.Fatalf("got error when umount volume test: %v", err)
	}
	err = driver.Unmount("test", "1")
	if err == nil {
		t.Fatalf("expect got error when umount unmounted volume test")
	}
	err = driver.Unmount("non-exist", "1")
	if err == nil {
		t.Fatalf("expect got error when unmounted volume non-exist")
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
